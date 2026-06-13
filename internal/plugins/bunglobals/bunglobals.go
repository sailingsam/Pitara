package bunglobals

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

func (p *Plugin) Name() string { return "bun-globals" }

func (p *Plugin) Dependencies() []string { return []string{"bun"} }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("bun") {
		result.Warnings = append(result.Warnings, "bun not found on PATH")
		result.Data, _ = marshalBun(emptyGlobals())
		return result, nil
	}

	out, err := executil.RunCombined(ctx, "bun", "pm", "ls", "-g")
	if err != nil && out == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("bun pm ls failed: %v", err))
		result.Data, _ = marshalBun(emptyGlobals())
		return result, nil
	}

	globals := parseBunList(out)
	data, err := marshalBun(&snapshot.GlobalPackages{Globals: globals})
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Bun *snapshot.GlobalPackages `json:"bun"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid bun snapshot data"
		return result, err
	}
	if payload.Bun == nil || len(payload.Bun.Globals) == 0 {
		result.Status = plugins.StatusSkipped
		result.Message = "no bun global packages in snapshot"
		return result, nil
	}

	if !executil.Available("bun") && !opts.DryRun {
		result.Status = plugins.StatusFailed
		result.Message = "bun not available"
		return result, fmt.Errorf("bun not available")
	}

	var failed int
	for _, pkg := range payload.Bun.Globals {
		spec := pkg.Name
		if pkg.Version != "" {
			spec = fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
		}

		if opts.DryRun {
			result.Details = append(result.Details, fmt.Sprintf("→ bun: %s", spec))
			continue
		}

		if _, err := executil.Run(ctx, "bun", "install", "-g", spec); err != nil {
			failed++
			result.Details = append(result.Details, fmt.Sprintf("✗ bun: %s — %s", spec, err.Error()))
			continue
		}
		result.Details = append(result.Details, fmt.Sprintf("✓ bun: %s", spec))
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install %d bun global package(s)", len(payload.Bun.Globals))
		return result, nil
	}

	if failed > 0 {
		result.Status = plugins.StatusFailed
		result.Message = fmt.Sprintf("%d of %d bun packages failed", failed, len(payload.Bun.Globals))
		return result, fmt.Errorf("%d bun package(s) failed", failed)
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("restored %d bun global package(s)", len(payload.Bun.Globals))
	return result, nil
}

func marshalBun(pkgs *snapshot.GlobalPackages) (json.RawMessage, error) {
	return json.Marshal(struct {
		Bun *snapshot.GlobalPackages `json:"bun"`
	}{Bun: pkgs})
}

func emptyGlobals() *snapshot.GlobalPackages {
	return &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{}}
}

// bun pm ls -g prints a text tree, e.g.:
//
//	/home/user/.bun/install/global node_modules (220)
//	├── @dotenvx/dotenvx@1.51.4
//	└── repomix@1.11.1
func parseBunList(output string) []snapshot.GlobalPackage {
	globals := make([]snapshot.GlobalPackage, 0)
	for _, line := range strings.Split(output, "\n") {
		idx := strings.Index(line, "── ")
		if idx == -1 {
			continue
		}
		spec := strings.TrimSpace(line[idx+len("── "):])
		if spec == "" {
			continue
		}
		name, version := splitNameVersion(spec)
		if name == "" || plugins.IsSelf(name) {
			continue
		}
		globals = append(globals, snapshot.GlobalPackage{Name: name, Version: version})
	}
	return globals
}

// splitNameVersion splits "name@version" on the LAST '@', so scoped packages
// like "@scope/pkg@1.2.3" parse correctly (name="@scope/pkg", version="1.2.3").
func splitNameVersion(spec string) (string, string) {
	at := strings.LastIndex(spec, "@")
	if at <= 0 {
		return spec, ""
	}
	return spec[:at], spec[at+1:]
}
