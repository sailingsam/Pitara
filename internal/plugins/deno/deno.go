package deno

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

func (p *Plugin) Name() string { return "deno" }

func (p *Plugin) Dependencies() []string { return nil }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("deno") {
		result.Warnings = append(result.Warnings, "deno not found on PATH")
		result.Data, _ = marshalDeno(nil)
		return result, nil
	}

	out, err := executil.Run(ctx, "deno", "--version")
	if err != nil {
		return result, err
	}

	rt := &snapshot.Runtime{
		Version: parseDenoVersion(out),
		Manager: "deno.land",
	}
	data, err := marshalDeno(rt)
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Deno *snapshot.Runtime `json:"deno"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid deno snapshot data"
		return result, err
	}
	if payload.Deno == nil || payload.Deno.Version == "" {
		result.Status = plugins.StatusSkipped
		result.Message = "no deno version in snapshot"
		return result, nil
	}

	target := payload.Deno.Version
	current := currentVersion(ctx)

	if current == target {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("Deno %s already installed", target)
		result.Details = append(result.Details, fmt.Sprintf("✓ Deno %s", target))
		return result, nil
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install Deno %s (current: %s)", target, orNone(current))
		result.Details = append(result.Details, fmt.Sprintf("→ Deno %s", target))
		return result, nil
	}

	if err := installDeno(ctx, target); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = err.Error()
		result.Details = append(result.Details, fmt.Sprintf("✗ Deno %s — %s", target, err.Error()))
		return result, err
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("Deno %s installed", target)
	result.Details = append(result.Details, fmt.Sprintf("✓ Deno %s", target))
	return result, nil
}

func marshalDeno(rt *snapshot.Runtime) (json.RawMessage, error) {
	return json.Marshal(struct {
		Deno *snapshot.Runtime `json:"deno"`
	}{Deno: rt})
}

// parseDenoVersion turns the multi-line `deno --version` output into "2.8.3".
// The first line looks like: "deno 2.8.3 (stable, release, x86_64-...)".
func parseDenoVersion(out string) string {
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "deno" {
			return fields[1]
		}
	}
	return strings.TrimSpace(out)
}

func currentVersion(ctx context.Context) string {
	if !executil.Available("deno") {
		return ""
	}
	out, err := executil.Run(ctx, "deno", "--version")
	if err != nil {
		return ""
	}
	return parseDenoVersion(out)
}

func installDeno(ctx context.Context, version string) error {
	switch runtime.GOOS {
	case "darwin", "linux":
		// Official installer takes the version as a positional arg.
		script := fmt.Sprintf("curl -fsSL https://deno.land/install.sh | sh -s %q", "v"+version)
		if _, err := executil.Run(ctx, "bash", "-lc", script); err == nil {
			return nil
		}
		if executil.Available("brew") {
			if _, err := executil.Run(ctx, "brew", "install", "deno"); err == nil {
				return nil
			}
		}
	case "windows":
		if executil.Available("winget") {
			if _, err := executil.Run(ctx, "winget", "install", "--id", "DenoLand.Deno", "-e"); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("could not install Deno %s automatically (see https://deno.land)", version)
}

func orNone(v string) string {
	if v == "" {
		return "not installed"
	}
	return v
}
