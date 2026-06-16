package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

type Engine struct {
	registry *plugins.Registry
}

func New(registry *plugins.Registry) *Engine {
	return &Engine{registry: registry}
}

func (e *Engine) Restore(ctx context.Context, snap *snapshot.Snapshot, opts plugins.RestoreOptions) ([]plugins.RestoreResult, error) {
	if opts.TargetOS == "" {
		opts.TargetOS = plugins.OS(runtime.GOOS)
	}
	if opts.TargetArch == "" {
		opts.TargetArch = runtime.GOARCH
	}

	ordered, err := e.registry.RestoreOrder()
	if err != nil {
		return nil, err
	}

	fullSnap, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}

	var results []plugins.RestoreResult
	failedDeps := make(map[string]bool)

	for _, plugin := range ordered {
		if hasFailedDependency(plugin, failedDeps) {
			results = append(results, plugins.RestoreResult{
				PluginName: plugin.Name(),
				Status:     plugins.StatusSkipped,
				Message:    "skipped due to failed dependency",
			})
			continue
		}

		payload, err := pluginPayload(plugin.Name(), fullSnap)
		if err != nil {
			return nil, err
		}

		if !hasRestoreData(plugin.Name(), payload) {
			results = append(results, plugins.RestoreResult{
				PluginName: plugin.Name(),
				Status:     plugins.StatusSkipped,
				Message:    "nothing to restore",
			})
			continue
		}

		result, err := plugin.Restore(ctx, payload, opts)
		if err != nil && result.Status == "" {
			result.Status = plugins.StatusFailed
			result.Message = err.Error()
		}
		if result.PluginName == "" {
			result.PluginName = plugin.Name()
		}

		if result.Status == plugins.StatusFailed {
			failedDeps[plugin.Name()] = true
		}

		results = append(results, result)
	}

	return results, nil
}

func hasFailedDependency(plugin plugins.Plugin, failed map[string]bool) bool {
	for _, dep := range plugin.Dependencies() {
		if failed[dep] {
			return true
		}
	}
	return false
}

func pluginPayload(name string, fullSnap []byte) (json.RawMessage, error) {
	var snap snapshot.Snapshot
	if err := json.Unmarshal(fullSnap, &snap); err != nil {
		return nil, err
	}

	switch name {
	case "node":
		return json.Marshal(struct {
			Node *snapshot.Runtime `json:"node"`
		}{Node: snap.Languages.Node})
	case "go":
		return json.Marshal(struct {
			Go *snapshot.Runtime `json:"go"`
		}{Go: snap.Languages.Go})
	case "java":
		return json.Marshal(struct {
			Java *snapshot.Java `json:"java"`
		}{Java: snap.Languages.Java})
	case "bun":
		return json.Marshal(struct {
			Bun *snapshot.Runtime `json:"bun"`
		}{Bun: snap.Languages.Bun})
	case "deno":
		return json.Marshal(struct {
			Deno *snapshot.Runtime `json:"deno"`
		}{Deno: snap.Languages.Deno})
	case "npm-globals":
		return json.Marshal(struct {
			NPM *snapshot.GlobalPackages `json:"npm"`
		}{NPM: snap.Packages.NPM})
	case "pnpm-globals":
		return json.Marshal(struct {
			PNPM *snapshot.GlobalPackages `json:"pnpm"`
		}{PNPM: snap.Packages.PNPM})
	case "bun-globals":
		return json.Marshal(struct {
			Bun *snapshot.GlobalPackages `json:"bun"`
		}{Bun: snap.Packages.Bun})
	case "deno-globals":
		return json.Marshal(struct {
			Deno *snapshot.DenoGlobals `json:"deno"`
		}{Deno: snap.Packages.Deno})
	default:
		return nil, fmt.Errorf("unknown plugin %q", name)
	}
}

func hasRestoreData(name string, payload json.RawMessage) bool {
	switch name {
	case "node":
		var data struct {
			Node *snapshot.Runtime `json:"node"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.Node != nil && data.Node.Version != ""
	case "go":
		var data struct {
			Go *snapshot.Runtime `json:"go"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.Go != nil && data.Go.Version != ""
	case "java":
		var data struct {
			Java *snapshot.Java `json:"java"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.Java != nil && data.Java.Version != ""
	case "bun":
		var data struct {
			Bun *snapshot.Runtime `json:"bun"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.Bun != nil && data.Bun.Version != ""
	case "deno":
		var data struct {
			Deno *snapshot.Runtime `json:"deno"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.Deno != nil && data.Deno.Version != ""
	case "npm-globals":
		var data struct {
			NPM *snapshot.GlobalPackages `json:"npm"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.NPM != nil && len(data.NPM.Globals) > 0
	case "pnpm-globals":
		var data struct {
			PNPM *snapshot.GlobalPackages `json:"pnpm"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.PNPM != nil && len(data.PNPM.Globals) > 0
	case "bun-globals":
		var data struct {
			Bun *snapshot.GlobalPackages `json:"bun"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.Bun != nil && len(data.Bun.Globals) > 0
	case "deno-globals":
		var data struct {
			Deno *snapshot.DenoGlobals `json:"deno"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.Deno != nil && len(data.Deno.Globals) > 0
	default:
		return false
	}
}
