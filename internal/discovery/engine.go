package discovery

import (
	"context"
	"fmt"

	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

type Engine struct {
	registry *plugins.Registry
}

func New(registry *plugins.Registry) *Engine {
	return &Engine{registry: registry}
}

func (e *Engine) Scan(ctx context.Context, label string) (*snapshot.Snapshot, []string, error) {
	var allWarnings []string
	results := make([]plugins.ScanResult, 0, len(e.registry.All()))

	for _, plugin := range e.registry.All() {
		result, err := plugin.Scan(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("scan %s: %w", plugin.Name(), err)
		}
		allWarnings = append(allWarnings, result.Warnings...)
		results = append(results, result)
	}

	snap, err := snapshot.BuildFromScan(label, results)
	if err != nil {
		return nil, nil, err
	}
	return snap, allWarnings, nil
}
