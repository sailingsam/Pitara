package node

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

func (p *Plugin) Name() string { return "node" }

func (p *Plugin) Dependencies() []string { return nil }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("node") {
		result.Warnings = append(result.Warnings, "node not found on PATH")
		data, _ := json.Marshal(struct {
			Node *snapshot.Runtime `json:"node"`
		}{})
		result.Data = data
		return result, nil
	}

	versionOut, err := executil.Run(ctx, "node", "-v")
	if err != nil {
		return result, err
	}

	rt := &snapshot.Runtime{
		Version: strings.TrimPrefix(versionOut, "v"),
		Manager: detectManager(),
	}

	data, err := json.Marshal(struct {
		Node *snapshot.Runtime `json:"node"`
	}{Node: rt})
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Node *snapshot.Runtime `json:"node"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid node snapshot data"
		return result, err
	}
	if payload.Node == nil || payload.Node.Version == "" {
		result.Status = plugins.StatusSkipped
		result.Message = "no node version in snapshot"
		return result, nil
	}

	target := payload.Node.Version
	current := currentVersion(ctx)

	if current == target {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("Node.js %s already installed", target)
		result.Details = append(result.Details, fmt.Sprintf("✓ Node.js %s", target))
		return result, nil
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install Node.js %s (current: %s)", target, orNone(current))
		result.Details = append(result.Details, fmt.Sprintf("→ Node.js %s", target))
		return result, nil
	}

	if err := installNode(ctx, target); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = err.Error()
		result.Details = append(result.Details, fmt.Sprintf("✗ Node.js %s — %s", target, err.Error()))
		return result, err
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("Node.js %s installed", target)
	result.Details = append(result.Details, fmt.Sprintf("✓ Node.js %s", target))
	return result, nil
}

func currentVersion(ctx context.Context) string {
	if !executil.Available("node") {
		return ""
	}
	out, err := executil.Run(ctx, "node", "-v")
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(out, "v")
}

func installNode(ctx context.Context, version string) error {
	if executil.Available("fnm") {
		if _, err := executil.Run(ctx, "fnm", "install", version); err == nil {
			_, err = executil.Run(ctx, "fnm", "use", version)
			return err
		}
	}

	// nvm-windows is a separate, native tool (not the Unix nvm.sh bash script):
	// it's driven directly as `nvm install <v>` / `nvm use <v>`.
	if runtime.GOOS == "windows" && executil.Available("nvm") {
		if _, err := executil.Run(ctx, "nvm", "install", version); err == nil {
			_, err = executil.Run(ctx, "nvm", "use", version)
			return err
		}
	}

	nvmSh := filepath.Join(nvmDir(), "nvm.sh")
	if _, err := os.Stat(nvmSh); err == nil {
		script := fmt.Sprintf("source %q && nvm install %s && nvm use %s", nvmSh, version, version)
		_, err := executil.Run(ctx, "bash", "-lc", script)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("could not install Node.js %s automatically (install nvm/fnm or install manually)", version)
}

func detectManager() string {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		return "unknown"
	}
	nodePath = filepath.Clean(nodePath)

	if os.Getenv("NVM_DIR") != "" || strings.Contains(nodePath, ".nvm") {
		return "nvm"
	}
	if strings.Contains(nodePath, "fnm") || os.Getenv("FNM_DIR") != "" {
		return "fnm"
	}
	if strings.Contains(nodePath, "mise") || strings.Contains(nodePath, ".local/share/mise") {
		return "mise"
	}
	if strings.Contains(nodePath, "homebrew") || strings.Contains(nodePath, "linuxbrew") {
		return "brew"
	}
	return "system"
}

func nvmDir() string {
	if dir := os.Getenv("NVM_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".nvm")
}

func orNone(v string) string {
	if v == "" {
		return "not installed"
	}
	return v
}
