# Contributing to Pitara

Thanks for wanting to help — Pitara gets better every time someone adds the tool
*they* use. Contributions of any size are welcome.

## Good first contribution: add a plugin

Pitara is built around one small interface. Each tool category (a runtime or a
package manager) implements it for both directions:

```go
type Plugin interface {
    Name() string
    Scan(ctx context.Context) (ScanResult, error)
    Restore(ctx context.Context, data json.RawMessage, opts RestoreOptions) (RestoreResult, error)
    Dependencies() []string   // plugins that must restore first
    SupportedOS() []OS        // platforms this plugin can restore on
}
```

To add a tool (say, Python or Rust), you touch a few small spots. Copy an
existing plugin of the same kind — a **runtime** like `internal/plugins/golang`,
or a **global-package manager** like `internal/plugins/npm`.

1. **Create the plugin** `internal/plugins/<tool>/<tool>.go` implementing `Plugin`.
2. **Add a parser test** `<tool>_test.go` that parses real command output (e.g.
   `--version`, or `list` output). If your parser ranges over a map, sort before
   comparing — Go map order is randomized and will make the test flaky otherwise.
3. **Register it** in `internal/app/registry.go`.
4. **Add a snapshot field** in `internal/snapshot/snapshot.go` — to `Languages`
   for a runtime, or `Packages` for a global-package manager.
5. **Store the scan** in `internal/snapshot/builder.go` — add a `case "<name>"`
   in `applyScanResult` that writes your data into the snapshot field.
6. **Wire up restore** in `internal/restore/engine.go` — add a `case "<name>"`
   in **both** `pluginPayload` and `hasRestoreData`.
7. **Add a report label** in `internal/report/report.go` (`titleFor`).

> Steps 4–7 are easy to miss: skip them and the plugin compiles, but
> `pitara scan` fails with `unknown plugin "<name>"`. Run `go run ./cmd/pitara scan`
> to confirm your tool shows up.

The registry topologically sorts plugins by `Dependencies()`, so prerequisites
(e.g. Node before npm globals) always install first. A global-package plugin
should depend on its runtime — e.g. `Dependencies() []string { return []string{"node"} }`.

## Before opening a PR

```bash
go build ./...
go test ./...
```

Please keep changes focused, match the surrounding code style, and add a test for
any new parser. If you're unsure about an approach, open an issue first — happy to
discuss.

## Reporting bugs & ideas

Open an [issue](https://github.com/sailingsam/pitara/issues) with what you ran, what
you expected, and what happened. For a parsing bug, the raw output of the tool's
`--version` (or `ls -g`) command is the most useful thing you can include.

## License

By contributing, you agree your contributions are licensed under the
[MIT License](LICENSE).
