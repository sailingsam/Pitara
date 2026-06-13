package java

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/sailingsam/pitara/internal/executil"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "java" }

func (p *Plugin) Dependencies() []string { return nil }

func (p *Plugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}

// `java -version` writes its banner to stderr, e.g.:
//   openjdk version "17.0.19" 2026-04-21
var versionRe = regexp.MustCompile(`version "([^"]+)"`)

func (p *Plugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	result := plugins.ScanResult{PluginName: p.Name()}

	if !executil.Available("java") {
		result.Warnings = append(result.Warnings, "java not found on PATH")
		result.Data, _ = marshalJava(nil)
		return result, nil
	}

	out, err := executil.RunCombined(ctx, "java", "-version")
	if err != nil {
		return result, err
	}

	version := parseVersion(out)
	if version == "" {
		result.Warnings = append(result.Warnings, "could not parse java version")
		result.Data, _ = marshalJava(nil)
		return result, nil
	}

	jv := &snapshot.Java{
		Version:      version,
		Distribution: detectDistribution(out),
	}
	data, err := marshalJava(jv)
	if err != nil {
		return result, err
	}
	result.Data = data
	return result, nil
}

func (p *Plugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	result := plugins.RestoreResult{PluginName: p.Name()}

	var payload struct {
		Java *snapshot.Java `json:"java"`
	}
	if err := json.Unmarshal(snap, &payload); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = "invalid java snapshot data"
		return result, err
	}
	if payload.Java == nil || payload.Java.Version == "" {
		result.Status = plugins.StatusSkipped
		result.Message = "no java version in snapshot"
		return result, nil
	}

	target := payload.Java.Version
	current := currentVersion(ctx)
	if current != "" && majorVersion(current) == majorVersion(target) {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("Java %s already installed", current)
		result.Details = append(result.Details, fmt.Sprintf("✓ Java %s", current))
		return result, nil
	}

	if opts.DryRun {
		result.Status = plugins.StatusSuccess
		result.Message = fmt.Sprintf("would install Java %s (current: %s)", target, orNone(current))
		result.Details = append(result.Details, fmt.Sprintf("→ Java %s", target))
		return result, nil
	}

	if err := installJava(ctx, target); err != nil {
		result.Status = plugins.StatusFailed
		result.Message = err.Error()
		result.Details = append(result.Details, fmt.Sprintf("✗ Java %s — %s", target, err.Error()))
		return result, err
	}

	result.Status = plugins.StatusSuccess
	result.Message = fmt.Sprintf("Java %s installed", target)
	result.Details = append(result.Details, fmt.Sprintf("✓ Java %s", target))
	return result, nil
}

func marshalJava(jv *snapshot.Java) (json.RawMessage, error) {
	return json.Marshal(struct {
		Java *snapshot.Java `json:"java"`
	}{Java: jv})
}

func parseVersion(out string) string {
	if m := versionRe.FindStringSubmatch(out); len(m) == 2 {
		return m[1]
	}
	return ""
}

// majorVersion returns the feature version: "17.0.19" -> "17", "1.8.0_392" -> "8".
func majorVersion(v string) string {
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return v
	}
	if parts[0] == "1" && len(parts) > 1 {
		return parts[1]
	}
	return parts[0]
}

func detectDistribution(out string) string {
	lower := strings.ToLower(out)
	switch {
	case strings.Contains(lower, "temurin"), strings.Contains(lower, "adoptium"):
		return "temurin"
	case strings.Contains(lower, "openjdk"):
		return "openjdk"
	case strings.Contains(lower, "zulu"):
		return "zulu"
	case strings.Contains(lower, "corretto"):
		return "corretto"
	case strings.Contains(lower, "graalvm"):
		return "graalvm"
	}
	return ""
}

func currentVersion(ctx context.Context) string {
	if !executil.Available("java") {
		return ""
	}
	out, err := executil.RunCombined(ctx, "java", "-version")
	if err != nil {
		return ""
	}
	return parseVersion(out)
}

func installJava(ctx context.Context, version string) error {
	major := majorVersion(version)
	switch runtime.GOOS {
	case "darwin":
		if executil.Available("brew") {
			if _, err := executil.Run(ctx, "brew", "install", "temurin@"+major); err == nil {
				return nil
			}
		}
	case "linux":
		if executil.Available("apt-get") {
			if _, err := executil.Run(ctx, "sudo", "apt-get", "install", "-y", "openjdk-"+major+"-jdk"); err == nil {
				return nil
			}
		}
		if executil.Available("brew") {
			if _, err := executil.Run(ctx, "brew", "install", "temurin@"+major); err == nil {
				return nil
			}
		}
	case "windows":
		if executil.Available("winget") {
			if _, err := executil.Run(ctx, "winget", "install", "--id", "EclipseAdoptium.Temurin."+major+".JDK", "-e"); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("could not install Java %s automatically (install Temurin %s from https://adoptium.net/)", version, major)
}

func orNone(v string) string {
	if v == "" {
		return "not installed"
	}
	return v
}
