package pnpm

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

func (p *Plugin) Name() string { return "pnpm-globals" }

func (p *Plugin) Dependencies() []string { return []string{"node"} }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("pnpm") {
		result.Warnings = append(result.Warnings, "pnpm not found on PATH")
		result.Data, _ = marshalPNPM(nil)
		return result, nil
	}

	out, err := executil.Run(ctx, "pnpm", "ls", "-g", "--json")
	if err != nil && out == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("pnpm ls failed: %v", err))
		result.Data, _ = marshalPNPM(emptyGlobals())
		return result, nil
	}

	globals, parseErr := parsePNPMList(out)
	if parseErr != nil {
		result.Warnings = append(result.Warnings, parseErr.Error())
		globals = []snapshot.GlobalPackage{}
	}

	data, err := marshalPNPM(&snapshot.GlobalPackages{Globals: globals})
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		PNPM *snapshot.GlobalPackages `json:"pnpm"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid pnpm snapshot data"
		return result, err
	}
	if payload.PNPM == nil || len(payload.PNPM.Globals) == 0 {
		result.Status = plugins.StatusSkipped
		result.Message = "no pnpm global packages in snapshot"
		return result, nil
	}

	if !executil.Available("pnpm") && !opts.DryRun {
		result.Status = plugins.StatusFailed
		result.Message = "pnpm not available"
		return result, fmt.Errorf("pnpm not available")
	}

	var failed int
	for _, pkg := range payload.PNPM.Globals {
		spec := pkg.Name
		if pkg.Version != "" {
			spec = fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
		}

		if opts.DryRun {
			result.Details = append(result.Details, fmt.Sprintf("→ pnpm: %s", spec))
			continue
		}

		if _, err := executil.Run(ctx, "pnpm", "add", "-g", spec); err != nil {
			failed++
			result.Details = append(result.Details, fmt.Sprintf("✗ pnpm: %s — %s", spec, err.Error()))
			continue
		}
		result.Details = append(result.Details, fmt.Sprintf("✓ pnpm: %s", spec))
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install %d pnpm global package(s)", len(payload.PNPM.Globals))
		return result, nil
	}

	if failed > 0 {
		result.Status = plugins.StatusFailed
		result.Message = fmt.Sprintf("%d of %d pnpm packages failed", failed, len(payload.PNPM.Globals))
		return result, fmt.Errorf("%d pnpm package(s) failed", failed)
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("restored %d pnpm global package(s)", len(payload.PNPM.Globals))
	return result, nil
}

func marshalPNPM(pkgs *snapshot.GlobalPackages) (json.RawMessage, error) {
	return json.Marshal(struct {
		PNPM *snapshot.GlobalPackages `json:"pnpm"`
	}{PNPM: pkgs})
}

func emptyGlobals() *snapshot.GlobalPackages {
	return &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{}}
}

// pnpm ls -g --json returns an array of project objects, each optionally
// carrying a "dependencies" map of name -> { version }.
type pnpmProject struct {
	Dependencies map[string]pnpmDep `json:"dependencies"`
}

type pnpmDep struct {
	Version string `json:"version"`
}

func parsePNPMList(output string) ([]snapshot.GlobalPackage, error) {
	var projects []pnpmProject
	if err := json.Unmarshal([]byte(output), &projects); err != nil {
		return nil, fmt.Errorf("parse pnpm ls output: %w", err)
	}

	globals := make([]snapshot.GlobalPackage, 0)
	for _, proj := range projects {
		for name, dep := range proj.Dependencies {
			if plugins.IsSelf(name) {
				continue
			}
			globals = append(globals, snapshot.GlobalPackage{
				Name:    name,
				Version: strings.TrimPrefix(dep.Version, "v"),
			})
		}
	}
	return globals, nil
}
