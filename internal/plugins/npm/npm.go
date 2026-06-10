package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sailingsam/pitara/internal/executil"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

var skipPackages = map[string]bool{
	"npm":  true,
	"core": true,
}

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "npm-globals" }

func (p *Plugin) Dependencies() []string { return []string{"node"} }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("npm") {
		result.Warnings = append(result.Warnings, "npm not found on PATH")
		data, _ := json.Marshal(struct {
			NPM *snapshot.GlobalPackages `json:"npm"`
		}{NPM: &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{}}})
		result.Data = data
		return result, nil
	}

	out, err := executil.Run(ctx, "npm", "list", "-g", "--depth=0", "--json")
	if err != nil && out == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("npm list failed: %v", err))
		data, _ := json.Marshal(struct {
			NPM *snapshot.GlobalPackages `json:"npm"`
		}{NPM: &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{}}})
		result.Data = data
		return result, nil
	}

	globals, parseErr := parseNPMList(out)
	if parseErr != nil {
		result.Warnings = append(result.Warnings, parseErr.Error())
		globals = []snapshot.GlobalPackage{}
	}

	data, err := json.Marshal(struct {
		NPM *snapshot.GlobalPackages `json:"npm"`
	}{NPM: &snapshot.GlobalPackages{Globals: globals}})
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		NPM *snapshot.GlobalPackages `json:"npm"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid npm snapshot data"
		return result, err
	}
	if payload.NPM == nil || len(payload.NPM.Globals) == 0 {
		result.Status = plugins.StatusSkipped
		result.Message = "no npm global packages in snapshot"
		return result, nil
	}

	if !executil.Available("npm") && !opts.DryRun {
		result.Status = plugins.StatusFailed
		result.Message = "npm not available (install node first)"
		return result, fmt.Errorf("npm not available")
	}

	var failed int
	for _, pkg := range payload.NPM.Globals {
		spec := pkg.Name
		if pkg.Version != "" {
			spec = fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
		}

		if opts.DryRun {
			result.Details = append(result.Details, fmt.Sprintf("→ npm: %s", spec))
			continue
		}

		if _, err := executil.Run(ctx, "npm", "install", "-g", spec); err != nil {
			failed++
			result.Details = append(result.Details, fmt.Sprintf("✗ npm: %s — %s", spec, err.Error()))
			continue
		}
		result.Details = append(result.Details, fmt.Sprintf("✓ npm: %s", spec))
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install %d npm global package(s)", len(payload.NPM.Globals))
		return result, nil
	}

	if failed > 0 {
		result.Status = plugins.StatusFailed
		result.Message = fmt.Sprintf("%d of %d npm packages failed", failed, len(payload.NPM.Globals))
		return result, fmt.Errorf("%d npm package(s) failed", failed)
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("restored %d npm global package(s)", len(payload.NPM.Globals))
	return result, nil
}

type npmListOutput struct {
	Dependencies map[string]npmDep `json:"dependencies"`
}

type npmDep struct {
	Version string `json:"version"`
}

func parseNPMList(output string) ([]snapshot.GlobalPackage, error) {
	var parsed npmListOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return nil, fmt.Errorf("parse npm list output: %w", err)
	}

	globals := make([]snapshot.GlobalPackage, 0, len(parsed.Dependencies))
	for name, dep := range parsed.Dependencies {
		if skipPackages[name] || plugins.IsSelf(name) {
			continue
		}
		version := strings.TrimPrefix(dep.Version, "v")
		globals = append(globals, snapshot.GlobalPackage{
			Name:    name,
			Version: version,
		})
	}
	return globals, nil
}
