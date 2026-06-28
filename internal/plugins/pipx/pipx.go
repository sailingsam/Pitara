package pipx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sailingsam/pitara/internal/executil"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

var skipPackages = map[string]bool{
	"pipx": true,
}

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "pipx-globals" }

func (p *Plugin) Dependencies() []string { return []string{"python"} }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("pipx") {
		result.Warnings = append(result.Warnings, "pipx not found on PATH")
		result.Data = emptyData()
		return result, nil
	}

	out, err := executil.Run(ctx, "pipx", "list", "--json")
	if err != nil && out == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("pipx list failed: %v", err))
		result.Data = emptyData()
		return result, nil
	}

	globals, parseErr := parsePipxList(out)
	if parseErr != nil {
		result.Warnings = append(result.Warnings, parseErr.Error())
		globals = []snapshot.GlobalPackage{}
	}

	data, err := json.Marshal(struct {
		Pipx *snapshot.GlobalPackages `json:"pipx"`
	}{Pipx: &snapshot.GlobalPackages{Globals: globals}})
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Pipx *snapshot.GlobalPackages `json:"pipx"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid pipx snapshot data"
		return result, err
	}
	if payload.Pipx == nil || len(payload.Pipx.Globals) == 0 {
		result.Status = plugins.StatusSkipped
		result.Message = "no pipx global packages in snapshot"
		return result, nil
	}

	if !executil.Available("pipx") && !opts.DryRun {
		result.Status = plugins.StatusFailed
		result.Message = "pipx not available (install python/pipx first)"
		return result, fmt.Errorf("pipx not available")
	}

	var failed int
	for _, pkg := range payload.Pipx.Globals {
		// pipx uses pip-style specs: name==version
		spec := pkg.Name
		if pkg.Version != "" {
			spec = fmt.Sprintf("%s==%s", pkg.Name, pkg.Version)
		}

		if opts.DryRun {
			result.Details = append(result.Details, fmt.Sprintf("→ pipx: %s", spec))
			continue
		}

		if _, err := executil.Run(ctx, "pipx", "install", spec); err != nil {
			failed++
			result.Details = append(result.Details, fmt.Sprintf("✗ pipx: %s — %s", spec, err.Error()))
			continue
		}
		result.Details = append(result.Details, fmt.Sprintf("✓ pipx: %s", spec))
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install %d pipx global package(s)", len(payload.Pipx.Globals))
		return result, nil
	}

	if failed > 0 {
		result.Status = plugins.StatusFailed
		result.Message = fmt.Sprintf("%d of %d pipx packages failed", failed, len(payload.Pipx.Globals))
		return result, fmt.Errorf("%d pipx package(s) failed", failed)
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("restored %d pipx global package(s)", len(payload.Pipx.Globals))
	return result, nil
}

// pipxListOutput mirrors `pipx list --json`. The top-level "venvs" is an object
// keyed by package name, so iterating it is NOT order-deterministic — callers
// that compare results must sort first.
type pipxListOutput struct {
	Venvs map[string]pipxVenv `json:"venvs"`
}

type pipxVenv struct {
	Metadata struct {
		MainPackage struct {
			Package        string `json:"package"`
			PackageVersion string `json:"package_version"`
		} `json:"main_package"`
	} `json:"metadata"`
}

func parsePipxList(output string) ([]snapshot.GlobalPackage, error) {
	var parsed pipxListOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return nil, fmt.Errorf("parse pipx list output: %w", err)
	}

	globals := make([]snapshot.GlobalPackage, 0, len(parsed.Venvs))
	for _, venv := range parsed.Venvs {
		name := venv.Metadata.MainPackage.Package
		if name == "" || skipPackages[name] || plugins.IsSelf(name) {
			continue
		}
		globals = append(globals, snapshot.GlobalPackage{
			Name:    name,
			Version: venv.Metadata.MainPackage.PackageVersion,
		})
	}
	return globals, nil
}

func emptyData() json.RawMessage {
	data, _ := json.Marshal(struct {
		Pipx *snapshot.GlobalPackages `json:"pipx"`
	}{Pipx: &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{}}})
	return data
}
