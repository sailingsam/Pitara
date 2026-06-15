package python

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sailingsam/pitara/internal/executil"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

type Plugin struct{}

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "python"
}

func (p *Plugin) Scan(
	ctx context.Context,
) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("python3") && !executil.Available("python") {
		result.Warnings = append(result.Warnings, "python not found on PATH")
		result.Data, _ = marshalPython(nil)
		return result, nil
	}

	out, err := runPythonVersion(ctx)
	if err != nil {
		return result, err
	}

	rt := &snapshot.Runtime{
		Version: parsePythonVersion(out),
		Manager: detectManager(),
	}

	data, err := marshalPython(rt)
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil

}

func (p *Plugin) Restore(
	ctx context.Context,
	snap json.RawMessage,
	opts plugins.RestoreOptions,
) (plugins.RestoreResult, error) {

	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Python *snapshot.Runtime `json:"python"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid python snapshot data"
		return result, err
	}

	if payload.Python == nil || payload.Python.Version == "" {
		result.Status = plugins.StatusSkipped
		result.Message = "no python version in snapshot"
		return result, nil
	}

	target := payload.Python.Version
	current := currentVersion(ctx)

	if current == target {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("Python %s already installed", target)
		result.Details = append(result.Details, fmt.Sprintf("✓ Python %s", target))
		return result, nil
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install Python %s (current: %s)", target, orNone(current))
		result.Details = append(result.Details, fmt.Sprintf("-> Python %s", target))
		return result, nil
	}

	if err := installPython(ctx, target); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = err.Error()
		result.Details = append(result.Details, fmt.Sprintf("✗ Python %s - %s", target, err.Error()))
		return result, err
	}
	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("Python %s installed", target)
	result.Details = append(result.Details, fmt.Sprintf("✓ Python %s", target))

	return result, nil
}

func (p *Plugin) Dependencies() []string { return nil }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{
		plugins.OSDarwin,
		plugins.OSLinux,
		plugins.OSWindows,
	}
}

func marshalPython(rt *snapshot.Runtime) (json.RawMessage, error) {
	return json.Marshal(struct {
		Python *snapshot.Runtime `json:"python"`
	}{Python: rt})
}

func runPythonVersion(ctx context.Context) (string, error) {
	if executil.Available("python3") {
		return executil.Run(ctx, "python3", "--version")
	}
	return executil.Run(ctx, "python", "--version")
}

// parsePythonVersion turns "Python 3.13.5" to "3.13.5"
func parsePythonVersion(str string) string {
	str = strings.TrimSpace(str)

	if strings.HasPrefix(str, "Python ") {
		return strings.TrimPrefix(str, "Python ")
	}
	// fallback for other formats (takes last field)
	fields := strings.Fields(str)
	if len(fields) > 0 {
		return fields[len(fields)-1]
	}

	return str
}

func detectManager() string {
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		pythonPath, _ = exec.LookPath("python")
	}
	if pythonPath == "" {
		return "unknown"
	}
	pythonPath = filepath.Clean(pythonPath)
	if strings.Contains(pythonPath, "homebrew") || strings.Contains(pythonPath, "linuxbrew") {
		return "brew"
	}
	if strings.Contains(pythonPath, "pyenv") {
		return "pyenv"
	}
	if strings.Contains(pythonPath, "mise") || strings.Contains(pythonPath, ".local/share/mise") {
		return "mise"
	}

	return "system"
}

func currentVersion(ctx context.Context) string {
	if !executil.Available("python3") && !executil.Available("python") {
		return ""
	}
	out, err := runPythonVersion(ctx)
	if err != nil {
		return ""
	}
	return parsePythonVersion(out)
}
func orNone(v string) string {
	if v == "" {
		return "not installed"
	}
	return v
}
func installPython(
	ctx context.Context,
	version string,
) error {
	switch runtime.GOOS {
	case "darwin":
		if executil.Available("brew") {
			if _, err := executil.Run(ctx,
				"brew",
				"install",
				fmt.Sprintf("python@%s", majorMinor(version)),
			); err == nil {
				return nil
			}
		}
	case "linux":
		if executil.Available("apt-get") {
			if _, err := executil.Run(
				ctx,
				"sudo",
				"apt-get",
				"install",
				"-y",
				fmt.Sprintf("python%s", majorMinor(version)),
			); err == nil {
				return nil
			}
		}
		if executil.Available("dnf") {
			if _, err := executil.Run(
				ctx,
				"sudo",
				"dnf",
				"install",
				"-y",
				fmt.Sprintf("python%s", majorMinor(version)),
			); err == nil {
				return nil
			}
		}
		// other package managers could be added later

	case "windows":
		if executil.Available("winget") {
			if _, err := executil.Run(
				ctx,
				"winget",
				"install",
				"--id",
				"Python.Python."+strings.ReplaceAll(majorMinor(version), ".", ""),
				"-e",
			); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("could not install Python %s automatically (install from https://www.python.org/downloads/ or via your package manager)", version)
}
func majorMinor(v string) string {
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}
