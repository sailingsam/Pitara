#!/usr/bin/env node
// Thin launcher: forwards all args to the downloaded native binary and mirrors its exit code.
"use strict";

const { spawnSync } = require("child_process");
const path = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = path.join(__dirname, "bin", "pitara" + ext);

const result = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(`pitara: failed to run binary — ${result.error.message}`);
  console.error("Try reinstalling: npm install -g pitara");
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
