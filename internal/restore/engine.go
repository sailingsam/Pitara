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
	case "npm-globals":
		return json.Marshal(struct {
			NPM *snapshot.GlobalPackages `json:"npm"`
		}{NPM: snap.Packages.NPM})
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
	case "npm-globals":
		var data struct {
			NPM *snapshot.GlobalPackages `json:"npm"`
		}
		_ = json.Unmarshal(payload, &data)
		return data.NPM != nil && len(data.NPM.Globals) > 0
	default:
		return false
	}
}
