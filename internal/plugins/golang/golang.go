package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sailingsam/pitara/internal/executil"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "go" }

func (p *Plugin) Dependencies() []string { return nil }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("go") {
		result.Warnings = append(result.Warnings, "go not found on PATH")
		result.Data, _ = marshalGo(nil)
		return result, nil
	}

	out, err := executil.Run(ctx, "go", "version")
	if err != nil {
		return result, err
	}

	rt := &snapshot.Runtime{
		Version: parseGoVersion(out),
		Manager: detectManager(),
	}
	data, err := marshalGo(rt)
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Go *snapshot.Runtime `json:"go"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid go snapshot data"
		return result, err
	}
	if payload.Go == nil || payload.Go.Version == "" {
		result.Status = plugins.StatusSkipped
		result.Message = "no go version in snapshot"
		return result, nil
	}

	target := payload.Go.Version
	current := currentVersion(ctx)

	if current == target {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("Go %s already installed", target)
		result.Details = append(result.Details, fmt.Sprintf("✓ Go %s", target))
		return result, nil
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install Go %s (current: %s)", target, orNone(current))
		result.Details = append(result.Details, fmt.Sprintf("→ Go %s", target))
		return result, nil
	}

	if err := installGo(ctx, target); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = err.Error()
		result.Details = append(result.Details, fmt.Sprintf("✗ Go %s — %s", target, err.Error()))
		return result, err
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("Go %s installed", target)
	result.Details = append(result.Details, fmt.Sprintf("✓ Go %s", target))
	return result, nil
}

func marshalGo(rt *snapshot.Runtime) (json.RawMessage, error) {
	return json.Marshal(struct {
		Go *snapshot.Runtime `json:"go"`
	}{Go: rt})
}

// parseGoVersion turns "go version go1.23.5 linux/amd64" into "1.23.5".
func parseGoVersion(out string) string {
	fields := strings.Fields(out)
	for _, f := range fields {
		if strings.HasPrefix(f, "go") && len(f) > 2 && (f[2] >= '0' && f[2] <= '9') {
			return strings.TrimPrefix(f, "go")
		}
	}
	return strings.TrimPrefix(out, "go version ")
}

func currentVersion(ctx context.Context) string {
	if !executil.Available("go") {
		return ""
	}
	out, err := executil.Run(ctx, "go", "version")
	if err != nil {
		return ""
	}
	return parseGoVersion(out)
}

func installGo(ctx context.Context, version string) error {
	switch runtime.GOOS {
	case "darwin", "linux":
		if executil.Available("brew") {
			if _, err := executil.Run(ctx, "brew", "install", "go"); err == nil {
				return nil
			}
		}
	case "windows":
		if executil.Available("winget") {
			if _, err := executil.Run(ctx, "winget", "install", "--id", "GoLang.Go", "-e"); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("could not install Go %s automatically (install from https://go.dev/dl/ or via your package manager)", version)
}

func detectManager() string {
	goRoot, err := exec.LookPath("go")
	if err != nil {
		return "unknown"
	}
	goRoot = filepath.Clean(goRoot)
	if strings.Contains(goRoot, "homebrew") || strings.Contains(goRoot, "linuxbrew") {
		return "brew"
	}
	if strings.Contains(goRoot, "mise") || strings.Contains(goRoot, ".local/share/mise") {
		return "mise"
	}
	if dir := os.Getenv("GOROOT"); dir != "" {
		return "system"
	}
	return "system"
}

func orNone(v string) string {
	if v == "" {
		return "not installed"
	}
	return v
}
