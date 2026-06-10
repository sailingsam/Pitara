#!/usr/bin/env node
// postinstall: download the Pitara binary for THIS machine's OS/arch from the
// matching GitHub Release, and place it at bin/pitara (bin/pitara.exe on Windows).
// The release tag is derived from this package's version (v<version>).
"use strict";

const fs = require("fs");
const path = require("path");
const https = require("https");

const { version } = require("./package.json");

const OS = { darwin: "darwin", linux: "linux", win32: "windows" }[process.platform];
const ARCH = { x64: "amd64", arm64: "arm64" }[process.arch];

if (!OS || !ARCH) {
  console.error(`pitara: unsupported platform ${process.platform}/${process.arch}`);
  process.exit(1);
}

const ext = process.platform === "win32" ? ".exe" : "";
const asset = `pitara_${OS}_${ARCH}${ext}`;
const url = `https://github.com/sailingsam/pitara/releases/download/v${version}/${asset}`;

const binDir = path.join(__dirname, "bin");
const dest = path.join(binDir, "pitara" + ext);

fs.mkdirSync(binDir, { recursive: true });

function fail(msg) {
  console.error(`pitara: could not download binary — ${msg}`);
  console.error(`Download it manually from https://github.com/sailingsam/pitara/releases/tag/v${version}`);
  process.exit(1);
}

function download(u, redirects) {
  if (redirects > 10) return fail("too many redirects");
  https
    .get(u, { headers: { "User-Agent": "pitara-npm-installer" } }, (res) => {
      // GitHub release downloads redirect to a storage URL — follow it.
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        return download(res.headers.location, redirects + 1);
      }
      if (res.statusCode !== 200) return fail(`HTTP ${res.statusCode} for ${u}`);

      const file = fs.createWriteStream(dest, { mode: 0o755 });
      res.pipe(file);
      file.on("finish", () => file.close(() => console.log(`pitara: installed ${asset}`)));
      file.on("error", (e) => fail(e.message));
    })
    .on("error", (e) => fail(e.message));
}

download(url, 0);
