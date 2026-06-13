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

To add a tool (say, Python or Rust):

1. Create `internal/plugins/<tool>/<tool>.go` implementing `Plugin`.
2. Register it in `internal/app/registry.go`.
3. Add a small test that parses real `--version` output (see existing plugins).

The registry topologically sorts plugins by `Dependencies()`, so prerequisites
(e.g. Node before npm globals) always install first.

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
