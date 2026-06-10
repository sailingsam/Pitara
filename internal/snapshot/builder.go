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
	case "npm-globals":
		var data struct {
			NPM *GlobalPackages `json:"npm"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return err
		}
		snap.Packages.NPM = data.NPM
	default:
		return fmt.Errorf("unknown plugin %q", result.PluginName)
	}
	return nil
}
