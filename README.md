# Pitara

> GitHub backs up your code. **Pitara backs up your development environment.**

Set up a new laptop in minutes instead of a lost day. Pitara captures the language
runtimes and global CLI tools on one machine and reinstalls them on another. It
captures **Node, Go, Java, and Bun**, plus **npm, pnpm, and bun** global packages.

> ### 🔒 You own it. We never see it.
>
> - **No server, no database** — backups live in *your own* GitHub repo
> - **No secrets stored** — snapshots hold only tool names and versions
> - **Your token never leaves your machine** (`~/.pitara`)
> - **Open source** — audit every line yourself

```bash
pitara login     # one-time: sign in with GitHub
pitara backup    # save this machine
pitara restore   # rebuild it on a new machine
```

That's the whole flow. Your backups land in a private `pitara-snapshots` repo
Pitara creates in your account.

---

## Why

Switching laptops or reinstalling an OS means rebuilding a setup scattered across
runtimes, package managers, and global CLIs. Existing tools each cover a slice —
dotfiles for config, Settings Sync for editors, Brewfiles for Homebrew, Nix for
full reproducibility (with a steep learning curve). Pitara takes the simple middle:
back up your **installation state** and restore it with one command.

---

## Install

**npm** (works with npm, pnpm, and bun — all use the npm registry):

```bash
npm install -g pitara
```

**From source** (requires [Go](https://go.dev/dl/) 1.23+):

```bash
go install github.com/sailingsam/pitara/cmd/pitara@latest
```

Prebuilt binaries for macOS, Linux, and Windows are attached to every
[release](https://github.com/sailingsam/pitara/releases).

---

## Usage

### Back up and restore

```bash
pitara login            # one-time: sign in with GitHub
pitara backup           # save this machine
pitara restore          # rebuild this setup on a new machine
pitara snapshots list   # see your saved machines
```

Backing up more than one machine? Add `--label <name>` to `backup` and `restore`
(e.g. `--label work-laptop`). With a single machine you never need it.

### Inspect a snapshot, or work offline

`pitara scan` prints the snapshot Pitara *would* create — handy to see exactly
what gets captured, or to save it as a file and move it yourself without an account:

```bash
pitara scan                          # print the snapshot (JSON) to the screen
pitara scan -o snapshot.json         # save it to a file instead
pitara restore --from snapshot.json  # restore from that file (add --dry-run to preview)
```

### The restore report

Restore installs in dependency order (Node before npm globals, and so on), skips
anything whose prerequisite failed, and prints a summary:

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

A snapshot is a small, versioned JSON document. Each package manager's globals are
kept in **separate lists** — Pitara never guesses which manager installed a tool,
and on restore reinstalls each via its original manager.

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

---

## How storage works

Your snapshots live in a private `pitara-snapshots` repo in **your own** GitHub
account (created on first backup). The CLI talks directly to the GitHub API — there
is **no Pitara server and no database** — and your token stays on your machine
(`~/.pitara`, `0600`). Every backup is a commit, so your history is just the git log.

Login uses GitHub's device flow (like the `gh` CLI) and requests the `repo` scope,
which is needed to create the private repo and commit snapshots. Because Pitara is
open source and runs no server, you can audit exactly what it does — and revoke
access anytime at [github.com/settings/applications](https://github.com/settings/applications).

---

## Architecture

Pitara is built around a small **plugin** abstraction. Each tool category implements
one interface for both directions:

```go
type Plugin interface {
    Name() string
    Scan(ctx context.Context) (ScanResult, error)
    Restore(ctx context.Context, data json.RawMessage, opts RestoreOptions) (RestoreResult, error)
    Dependencies() []string   // plugins that must restore first
    SupportedOS() []OS        // platforms this plugin can restore on
}
```

Adding a new tool is: implement `Plugin`, register it. The registry topologically
sorts plugins by `Dependencies()` so prerequisites install first.

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
  github/            GitHub device-login + REST client
  executil/          Command helpers
```

The CLI depends only on Cobra; the GitHub client is pure standard library, so the
installed binary stays small.

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
implement the `Plugin` interface, and register it. Please run `go build ./...` and
`go test ./...` before opening a PR.

---

## License

[MIT](LICENSE) © sailingsam
