package cargoglobals

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

func (p *Plugin) Name() string { return "cargo-globals" }

func (p *Plugin) Dependencies() []string { return []string{"rust"} }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("cargo") {
		result.Warnings = append(result.Warnings, "cargo not found on PATH")
		result.Data, _ = marshalCargo(emptyGlobals())
		return result, nil
	}

	out, err := executil.Run(ctx, "cargo", "install", "--list")
	if err != nil && out == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("cargo install --list failed: %v", err))
		result.Data, _ = marshalCargo(emptyGlobals())
		return result, nil
	}

	globals := parseCargoList(out)
	data, err := marshalCargo(&snapshot.GlobalPackages{Globals: globals})
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Cargo *snapshot.GlobalPackages `json:"cargo"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid cargo snapshot data"
		return result, err
	}
	if payload.Cargo == nil || len(payload.Cargo.Globals) == 0 {
		result.Status = plugins.StatusSkipped
		result.Message = "no cargo global packages in snapshot"
		return result, nil
	}

	if !executil.Available("cargo") && !opts.DryRun {
		result.Status = plugins.StatusFailed
		result.Message = "cargo not available (install rust first)"
		return result, fmt.Errorf("cargo not available")
	}

	var failed int
	for _, pkg := range payload.Cargo.Globals {
		args := []string{"install", pkg.Name}
		if pkg.Version != "" {
			args = append(args, "--version", pkg.Version)
		}

		if opts.DryRun {
			spec := pkg.Name
			if pkg.Version != "" {
				spec = fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
			}
			result.Details = append(result.Details, fmt.Sprintf("→ cargo: %s", spec))
			continue
		}

		if _, err := executil.Run(ctx, "cargo", args...); err != nil {
			failed++
			result.Details = append(result.Details, fmt.Sprintf("✗ cargo: %s — %s", pkg.Name, err.Error()))
			continue
		}
		result.Details = append(result.Details, fmt.Sprintf("✓ cargo: %s", pkg.Name))
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install %d cargo global crate(s)", len(payload.Cargo.Globals))
		return result, nil
	}

	if failed > 0 {
		result.Status = plugins.StatusFailed
		result.Message = fmt.Sprintf("%d of %d cargo crates failed", failed, len(payload.Cargo.Globals))
		return result, fmt.Errorf("%d cargo crate(s) failed", failed)
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("restored %d cargo global crate(s)", len(payload.Cargo.Globals))
	return result, nil
}

func marshalCargo(pkgs *snapshot.GlobalPackages) (json.RawMessage, error) {
	return json.Marshal(struct {
		Cargo *snapshot.GlobalPackages `json:"cargo"`
	}{Cargo: pkgs})
}

func emptyGlobals() *snapshot.GlobalPackages {
	return &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{}}
}

// parseCargoList reads `cargo install --list`, e.g.:
//
//	ripgrep v15.1.0:
//	    rg
//	cargo-edit v0.12.0:
//	    cargo-add
//
// Crate lines are NOT indented and end with ":"; the indented lines below are
// the binaries each crate installs (skipped). Crates installed from a git URL
// or local path can't be reconstructed from name alone — only crates.io ones
// restore cleanly.
func parseCargoList(output string) []snapshot.GlobalPackage {
	globals := make([]snapshot.GlobalPackage, 0)
	for _, line := range strings.Split(output, "\n") {
		// Indented lines are binaries, not crates.
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue
		}
		line = strings.TrimSpace(line)
		if !strings.HasSuffix(line, ":") {
			continue
		}
		fields := strings.Fields(strings.TrimSuffix(line, ":"))
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		version := strings.TrimPrefix(fields[1], "v")
		if plugins.IsSelf(name) {
			continue
		}
		globals = append(globals, snapshot.GlobalPackage{Name: name, Version: version})
	}
	return globals
}
