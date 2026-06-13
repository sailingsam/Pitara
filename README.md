# Pitara

> GitHub backs up your code. **Pitara backs up your development environment.**

Pitara is a CLI that scans a machine for language runtimes and globally installed
CLI tools, captures them as a portable snapshot, and restores them on a new
machine — so setting up a fresh laptop is closer to one command than one lost day.

```bash
pitara scan -o snapshot.json   # capture this machine
pitara restore --from snapshot.json   # rebuild it elsewhere
```

---

## Status

Pitara works today. Local backup/restore is fully offline; cloud backup uses
**your own GitHub** — no Pitara server, no database.

- ✅ `pitara scan` — discover Node.js, Go, Java, and Bun runtimes plus npm, pnpm,
  and bun global packages; output a versioned snapshot
- ✅ `pitara restore --from <file>` — restore from a local snapshot, with `--dry-run`
- ✅ Plugin architecture with dependency-ordered restore and a clear ✓/⚠/✗ report
- ✅ `pitara login` — GitHub device-flow auth (like `gh`)
- ✅ `pitara backup` — commit a snapshot to a private `pitara-snapshots` repo in
  your own account (auto-created on first use)
- ✅ `pitara snapshots list` / `pitara restore` — list and restore your backups

What's planned (see [Roadmap](#roadmap)):

- Cross-OS restore warnings and broader version-manager detection
- Retry/recovery logic

---

## Why

Switching laptops, reinstalling an OS, or onboarding to a new machine means
recreating a setup that's scattered across runtimes, package managers, and global
CLIs. Existing tools each solve a slice — dotfiles for config, Settings Sync for
editors, Brewfiles for Homebrew, Nix for full reproducibility (with a steep curve).
Pitara aims for the simple middle: back up your **installation state** and restore
it with one command.

It intentionally does **not** handle secrets — no SSH keys, auth tokens, or env
vars. Snapshots are safe to store and share.

---

## Install

**npm** (works with npm, pnpm, and bun — all use the npm registry):

```bash
npm install -g pitara
```

**From source** (requires [Go](https://go.dev/dl/) 1.23+):

```bash
go install github.com/sailingsam/pitara/cmd/pitara@latest

# or
git clone https://github.com/sailingsam/pitara.git
cd pitara && go build -o pitara ./cmd/pitara
```

Prebuilt binaries for macOS, Linux, and Windows are attached to each
[GitHub Release](https://github.com/sailingsam/pitara/releases).

---

## Usage

### Scan

```bash
pitara scan                      # print snapshot JSON to stdout
pitara scan -o snapshot.json     # write to a file
pitara scan --label work-laptop  # tag the snapshot with a machine label
```

### Restore

```bash
pitara restore --from snapshot.json             # restore runtimes + globals
pitara restore --from snapshot.json --dry-run   # show the plan, install nothing
```

### Cloud

Pitara stores backups in **your own GitHub** — a private `pitara-snapshots` repo
it creates for you. No Pitara server or database is involved; the token lives only
on your machine and talks directly to GitHub. See [Cloud](#cloud) for setup.

```bash
pitara login          # authorize via GitHub (device flow, like `gh`)
pitara backup         # scan + commit this machine to your GitHub
pitara snapshots list # see your saved machines
pitara restore        # restore this machine's latest backup
```

Multiple machines? Add `--label <name>` to `backup`/`restore`. Most people with
one machine never need it.

Restore runs plugins in dependency order (e.g. Node before npm globals), skips
anything with a failed prerequisite, and prints a summary report:

```text
Pitara Restore Report
───────────────────────
Snapshot:  work-laptop (linux, amd64)
Created:   2026-06-10 14:00 UTC

Runtimes
  ✓ Node.js 22.15.0

Global Packages (npm)
  ✓ npm: typescript@5.4.5
  ✓ npm: tsx@4.7.0

Restore completed.
```

---

## Snapshot format

A snapshot is a small, versioned JSON document:

```json
{
  "schemaVersion": 1,
  "createdAt": "2026-06-10T14:00:00Z",
  "machine": { "label": "work-laptop", "os": "linux", "arch": "amd64" },
  "languages": {
    "node": { "version": "22.15.0", "manager": "nvm" }
  },
  "packages": {
    "npm": {
      "globals": [
        { "name": "typescript", "version": "5.4.5" },
        { "name": "tsx", "version": "4.7.0" }
      ]
    }
  }
}
```

Each package manager's globals are recorded in **separate lists** — Pitara never
guesses which manager installed a tool. On restore, each is reinstalled via its
original manager.

---

## Architecture

Pitara is built around a small **plugin** abstraction. Each tool category
implements one interface for both directions:

```go
type Plugin interface {
    Name() string
    Scan(ctx context.Context) (ScanResult, error)
    Restore(ctx context.Context, data json.RawMessage, opts RestoreOptions) (RestoreResult, error)
    Dependencies() []string   // plugins that must restore first
    SupportedOS() []OS        // platforms this plugin can restore on
}
```

Adding support for a new tool is: implement `Plugin`, register it. The registry
topologically sorts plugins by `Dependencies()` so prerequisites install first.

```text
cmd/pitara/          CLI entrypoint
internal/
  cli/               Cobra commands (scan, restore, login, backup, ...)
  discovery/         Scan orchestrator
  restore/           Restore planner + executor
  snapshot/          Schema, validation, builder
  report/            Restore report formatter
  plugins/           Plugin interface + registry (node, go, java, bun, npm, ...)
  auth/              Local session storage (~/.pitara)
  github/            GitHub device-login + REST client (cloud storage)
  executil/          Command helpers
```

The CLI depends only on Cobra; the GitHub client is pure standard library, so the
installed binary stays small.

---

## Cloud

Pitara's cloud backup is **GitHub-as-storage**: your snapshots live in a private
`pitara-snapshots` repo in your own account. There is **no Pitara server or
database** — the CLI talks directly to the GitHub API, and your token never leaves
your machine (`~/.pitara`, `0600`).

Just run `pitara login` and approve in your browser — no setup needed:

```bash
pitara login    # authorize via GitHub (device flow)
pitara backup   # creates your private pitara-snapshots repo and commits a snapshot
```

Pitara requests the `repo` scope (needed to create the private repo and commit
snapshots). Because the project is open source and runs no server, you can audit
exactly what it does, and revoke access anytime at
[github.com/settings/applications](https://github.com/settings/applications).

---

## Roadmap

| Phase | Scope | Status |
|-------|-------|--------|
| 1 | CLI foundation: `scan`, local `restore`, Node + npm plugins | ✅ |
| 2 | Go, Java, Bun runtimes; pnpm + bun globals; per-OS install paths | ✅ |
| 3 | Cloud: GitHub-backed `login`, `backup`, `snapshots list`, restore | ✅ |
| 4 | Cross-OS warnings, version-manager detection, retry/recovery | ⏳ |

---

## Contributing

Contributions are welcome. Good first additions are new plugins — pick a tool,
implement the `Plugin` interface, and register it. Please run `go build ./...`
and `go test ./...` before opening a PR.

---

## License

[MIT](LICENSE) © sailingsam
