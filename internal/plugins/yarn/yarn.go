package yarn

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

func (p *Plugin) Name() string { return "yarn-globals" }

func (p *Plugin) Dependencies() []string { return []string{"node"} }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("yarn") {
		result.Warnings = append(result.Warnings, "yarn not found on PATH")
		result.Data, _ = marshalYarn(emptyGlobals())
		return result, nil
	}

	// `yarn global` only exists in Yarn classic (v1); Yarn berry removed it,
	// so this errors there — treat any failure as "nothing to capture".
	out, err := executil.Run(ctx, "yarn", "global", "list")
	if err != nil && out == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("yarn global list failed (Yarn classic only): %v", err))
		result.Data, _ = marshalYarn(emptyGlobals())
		return result, nil
	}

	globals := parseYarnList(out)
	data, err := marshalYarn(&snapshot.GlobalPackages{Globals: globals})
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Yarn *snapshot.GlobalPackages `json:"yarn"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid yarn snapshot data"
		return result, err
	}
	if payload.Yarn == nil || len(payload.Yarn.Globals) == 0 {
		result.Status = plugins.StatusSkipped
		result.Message = "no yarn global packages in snapshot"
		return result, nil
	}

	if !executil.Available("yarn") && !opts.DryRun {
		result.Status = plugins.StatusFailed
		result.Message = "yarn not available (install node/yarn first)"
		return result, fmt.Errorf("yarn not available")
	}

	var failed int
	for _, pkg := range payload.Yarn.Globals {
		spec := pkg.Name
		if pkg.Version != "" {
			spec = fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
		}

		if opts.DryRun {
			result.Details = append(result.Details, fmt.Sprintf("→ yarn: %s", spec))
			continue
		}

		if _, err := executil.Run(ctx, "yarn", "global", "add", spec); err != nil {
			failed++
			result.Details = append(result.Details, fmt.Sprintf("✗ yarn: %s — %s", spec, err.Error()))
			continue
		}
		result.Details = append(result.Details, fmt.Sprintf("✓ yarn: %s", spec))
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install %d yarn global package(s)", len(payload.Yarn.Globals))
		return result, nil
	}

	if failed > 0 {
		result.Status = plugins.StatusFailed
		result.Message = fmt.Sprintf("%d of %d yarn packages failed", failed, len(payload.Yarn.Globals))
		return result, fmt.Errorf("%d yarn package(s) failed", failed)
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("restored %d yarn global package(s)", len(payload.Yarn.Globals))
	return result, nil
}

func marshalYarn(pkgs *snapshot.GlobalPackages) (json.RawMessage, error) {
	return json.Marshal(struct {
		Yarn *snapshot.GlobalPackages `json:"yarn"`
	}{Yarn: pkgs})
}

func emptyGlobals() *snapshot.GlobalPackages {
	return &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{}}
}

// parseYarnList reads `yarn global list` output, e.g.:
//
//	yarn global v1.22.22
//	info "tiny-cli@1.1.5" has binaries:
//	   - tiny
//	Done in 0.05s.
//
// The package name/version lives in the quoted "name@version" on each
// `info "..." has binaries:` line.
func parseYarnList(output string) []snapshot.GlobalPackage {
	globals := make([]snapshot.GlobalPackage, 0)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "info \"") || !strings.Contains(line, "has binaries") {
			continue
		}
		q1 := strings.Index(line, "\"")
		q2 := strings.Index(line[q1+1:], "\"")
		if q1 == -1 || q2 == -1 {
			continue
		}
		spec := line[q1+1 : q1+1+q2]
		name, ver := splitNameVersion(spec)
		if name == "" || plugins.IsSelf(name) {
			continue
		}
		globals = append(globals, snapshot.GlobalPackage{Name: name, Version: ver})
	}
	return globals
}

// splitNameVersion splits "name@version" on the LAST '@' so scoped packages
// like "@scope/pkg@1.2.3" parse correctly.
func splitNameVersion(spec string) (string, string) {
	at := strings.LastIndex(spec, "@")
	if at <= 0 {
		return spec, ""
	}
	return spec[:at], spec[at+1:]
}
