package bun

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/sailingsam/pitara/internal/executil"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "bun" }

func (p *Plugin) Dependencies() []string { return nil }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("bun") {
		result.Warnings = append(result.Warnings, "bun not found on PATH")
		result.Data, _ = marshalBun(nil)
		return result, nil
	}

	out, err := executil.Run(ctx, "bun", "-v")
	if err != nil {
		return result, err
	}

	rt := &snapshot.Runtime{
		Version: strings.TrimPrefix(strings.TrimSpace(out), "v"),
		Manager: "bun.sh",
	}
	data, err := marshalBun(rt)
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Bun *snapshot.Runtime `json:"bun"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid bun snapshot data"
		return result, err
	}
	if payload.Bun == nil || payload.Bun.Version == "" {
		result.Status = plugins.StatusSkipped
		result.Message = "no bun version in snapshot"
		return result, nil
	}

	target := payload.Bun.Version
	current := currentVersion(ctx)

	if current == target {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("Bun %s already installed", target)
		result.Details = append(result.Details, fmt.Sprintf("✓ Bun %s", target))
		return result, nil
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install Bun %s (current: %s)", target, orNone(current))
		result.Details = append(result.Details, fmt.Sprintf("→ Bun %s", target))
		return result, nil
	}

	if err := installBun(ctx, target); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = err.Error()
		result.Details = append(result.Details, fmt.Sprintf("✗ Bun %s — %s", target, err.Error()))
		return result, err
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("Bun %s installed", target)
	result.Details = append(result.Details, fmt.Sprintf("✓ Bun %s", target))
	return result, nil
}

func marshalBun(rt *snapshot.Runtime) (json.RawMessage, error) {
	return json.Marshal(struct {
		Bun *snapshot.Runtime `json:"bun"`
	}{Bun: rt})
}

func currentVersion(ctx context.Context) string {
	if !executil.Available("bun") {
		return ""
	}
	out, err := executil.Run(ctx, "bun", "-v")
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(out), "v")
}

func installBun(ctx context.Context, version string) error {
	switch runtime.GOOS {
	case "darwin", "linux":
		// Official installer; honours BUN_VERSION for a pinned release.
		script := fmt.Sprintf("BUN_VERSION=%q bash -c 'curl -fsSL https://bun.sh/install | bash'", "bun-v"+version)
		if _, err := executil.Run(ctx, "bash", "-lc", script); err == nil {
			return nil
		}
		if executil.Available("brew") {
			if _, err := executil.Run(ctx, "brew", "install", "oven-sh/bun/bun"); err == nil {
				return nil
			}
		}
	case "windows":
		if executil.Available("powershell") {
			if _, err := executil.Run(ctx, "powershell", "-c", "irm bun.sh/install.ps1 | iex"); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("could not install Bun %s automatically (see https://bun.sh)", version)
}

func orNone(v string) string {
	if v == "" {
		return "not installed"
	}
	return v
}
