package snapshot

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/sailingsam/pitara/internal/plugins"
)

func BuildFromScan(label string, results []plugins.ScanResult) (*Snapshot, error) {
	snap := New(label, runtime.GOOS, runtime.GOARCH)
	snap.CreatedAt = time.Now().UTC()

	for _, result := range results {
		if err := applyScanResult(snap, result); err != nil {
			return nil, fmt.Errorf("plugin %s: %w", result.PluginName, err)
		}
	}

	if err := snap.Validate(); err != nil {
		return nil, err
	}
	return snap, nil
}

func applyScanResult(snap *Snapshot, result plugins.ScanResult) error {
	switch result.PluginName {
	case "node":
		var data struct {
			Node *Runtime `json:"node"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Languages.Node = data.Node
	case "go":
		var data struct {
			Go *Runtime `json:"go"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Languages.Go = data.Go
	case "java":
		var data struct {
			Java *Java `json:"java"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Languages.Java = data.Java
	case "bun":
		var data struct {
			Bun *Runtime `json:"bun"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Languages.Bun = data.Bun
	case "npm-globals":
		var data struct {
			NPM *GlobalPackages `json:"npm"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Packages.NPM = data.NPM
	case "pnpm-globals":
		var data struct {
			PNPM *GlobalPackages `json:"pnpm"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Packages.PNPM = data.PNPM
	case "bun-globals":
		var data struct {
			Bun *GlobalPackages `json:"bun"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Packages.Bun = data.Bun
	default:
		return fmt.Errorf("unknown plugin %q", result.PluginName)
	}
	return nil
}
