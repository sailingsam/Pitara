package plugins

import (
	"context"
	"encoding/json"
	"testing"
)

type stubPlugin struct {
	name string
	deps []string
}

func (s stubPlugin) Name() string { return s.name }
func (s stubPlugin) Scan(ctx context.Context) (ScanResult, error) {
	return ScanResult{PluginName: s.name, Data: json.RawMessage(`{}`)}, nil
}
func (s stubPlugin) Restore(ctx context.Context, snap json.RawMessage, opts RestoreOptions) (RestoreResult, error) {
	return RestoreResult{PluginName: s.name, Status: StatusSuccess}, nil
}
func (s stubPlugin) Dependencies() []string { return s.deps }
func (s stubPlugin) SupportedOS() []OS       { return []OS{OSLinux} }

func TestRestoreOrder(t *testing.T) {
	r := NewRegistry(
		stubPlugin{name: "npm-globals", deps: []string{"node"}},
		stubPlugin{name: "node"},
	)

	ordered, err := r.RestoreOrder()
	if err != nil {
		t.Fatalf("restore order: %v", err)
	}
	if len(ordered) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(ordered))
	}
	if ordered[0].Name() != "node" {
		t.Fatalf("expected node first, got %s", ordered[0].Name())
	}
	if ordered[1].Name() != "npm-globals" {
		t.Fatalf("expected npm-globals second, got %s", ordered[1].Name())
	}
}
