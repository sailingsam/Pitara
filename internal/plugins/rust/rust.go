package rust

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sailingsam/pitara/internal/executil"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "rust" }

func (p *Plugin) Dependencies() []string { return nil }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("rustc") {
		result.Warnings = append(result.Warnings, "rustc not found on PATH")
		result.Data, _ = marshalRust(nil)
		return result, nil
	}

	out, err := executil.Run(ctx, "rustc", "--version")
	if err != nil {
		return result, err
	}

	rt := &snapshot.Runtime{
		Version: parseRustVersion(out),
		Manager: detectManager(),
	}
	data, err := marshalRust(rt)
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Rust *snapshot.Runtime `json:"rust"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid rust snapshot data"
		return result, err
	}
	if payload.Rust == nil || payload.Rust.Version == "" {
		result.Status = plugins.StatusSkipped
		result.Message = "no rust version in snapshot"
		return result, nil
	}

	target := payload.Rust.Version
	current := currentVersion(ctx)

	if current == target {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("Rust %s already installed", target)
		result.Details = append(result.Details, fmt.Sprintf("✓ Rust %s", target))
		return result, nil
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install Rust %s (current: %s)", target, orNone(current))
		result.Details = append(result.Details, fmt.Sprintf("→ Rust %s", target))
		return result, nil
	}

	if err := installRust(ctx, target); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = err.Error()
		result.Details = append(result.Details, fmt.Sprintf("✗ Rust %s — %s", target, err.Error()))
		return result, err
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("Rust %s installed", target)
	result.Details = append(result.Details, fmt.Sprintf("✓ Rust %s", target))
	return result, nil
}

func marshalRust(rt *snapshot.Runtime) (json.RawMessage, error) {
	return json.Marshal(struct {
		Rust *snapshot.Runtime `json:"rust"`
	}{Rust: rt})
}

// parseRustVersion turns "rustc 1.78.0 (9b00956e5 2024-04-29)" into "1.78.0".
func parseRustVersion(out string) string {
	fields := strings.Fields(out)
	if len(fields) >= 2 && fields[0] == "rustc" {
		return fields[1]
	}
	return strings.TrimSpace(out)
}

func currentVersion(ctx context.Context) string {
	if !executil.Available("rustc") {
		return ""
	}
	out, err := executil.Run(ctx, "rustc", "--version")
	if err != nil {
		return ""
	}
	return parseRustVersion(out)
}

func detectManager() string {
	if executil.Available("rustup") {
		return "rustup"
	}
	return "system"
}

func installRust(ctx context.Context, version string) error {
	// rustup is the canonical way; if it's already there just pin the toolchain.
	if executil.Available("rustup") {
		if _, err := executil.Run(ctx, "rustup", "toolchain", "install", version); err == nil {
			_, _ = executil.Run(ctx, "rustup", "default", version)
			return nil
		}
	}
	return fmt.Errorf("could not install Rust %s automatically (install rustup from https://rustup.rs)", version)
}

func orNone(v string) string {
	if v == "" {
		return "not installed"
	}
	return v
}
