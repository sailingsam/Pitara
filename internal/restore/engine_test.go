package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

// fakePlugin is a controllable Plugin for exercising the restore engine.
// It uses a real plugin name (e.g. "node", "npm-globals") so that the engine's
// pluginPayload / hasRestoreData switches resolve it.
type fakePlugin struct {
	name    string
	deps    []string
	fail    bool // Restore returns StatusFailed
	called  bool // set true when Restore is invoked
	gotOpts plugins.RestoreOptions
}

func (f *fakePlugin) Name() string           { return f.name }
func (f *fakePlugin) Dependencies() []string { return f.deps }
func (f *fakePlugin) SupportedOS() []plugins.OS {
	return []plugins.OS{plugins.OSDarwin, plugins.OSLinux, plugins.OSWindows}
}
func (f *fakePlugin) Scan(ctx context.Context) (plugins.ScanResult, error) {
	return plugins.ScanResult{PluginName: f.name}, nil
}
func (f *fakePlugin) Restore(ctx context.Context, snap json.RawMessage, opts plugins.RestoreOptions) (plugins.RestoreResult, error) {
	f.called = true
	f.gotOpts = opts
	if f.fail {
		return plugins.RestoreResult{PluginName: f.name, Status: plugins.StatusFailed, Message: "boom"}, fmt.Errorf("boom")
	}
	return plugins.RestoreResult{PluginName: f.name, Status: plugins.StatusSuccess}, nil
}

// resultFor finds a plugin's result by name.
func resultFor(results []plugins.RestoreResult, name string) (plugins.RestoreResult, bool) {
	for _, r := range results {
		if r.PluginName == name {
			return r, true
		}
	}
	return plugins.RestoreResult{}, false
}

func snapWithNodeAndNPM() *snapshot.Snapshot {
	return &snapshot.Snapshot{
		Languages: snapshot.Languages{Node: &snapshot.Runtime{Version: "20.0.0"}},
		Packages: snapshot.Packages{
			NPM: &snapshot.GlobalPackages{Globals: []snapshot.GlobalPackage{{Name: "typescript", Version: "5.4.5"}}},
		},
	}
}

func TestHasFailedDependency(t *testing.T) {
	p := &fakePlugin{name: "npm-globals", deps: []string{"node"}}
	if !hasFailedDependency(p, map[string]bool{"node": true}) {
		t.Error("expected true when a dependency is in the failed set")
	}
	if hasFailedDependency(p, map[string]bool{}) {
		t.Error("expected false when no dependency has failed")
	}
}

func TestPluginPayloadRoutesToCorrectField(t *testing.T) {
	snap := snapWithNodeAndNPM()
	full, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}

	// node + npm-globals have data → hasRestoreData true; go has none → false.
	for _, tc := range []struct {
		name string
		want bool
	}{
		{"node", true},
		{"npm-globals", true},
		{"go", false},
		{"pipx-globals", false},
	} {
		payload, err := pluginPayload(tc.name, full)
		if err != nil {
			t.Fatalf("pluginPayload(%q) error: %v", tc.name, err)
		}
		if got := hasRestoreData(tc.name, payload); got != tc.want {
			t.Errorf("hasRestoreData(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestPluginPayloadUnknownPlugin(t *testing.T) {
	if _, err := pluginPayload("does-not-exist", []byte(`{}`)); err == nil {
		t.Error("expected an error for an unknown plugin name")
	}
}

func TestHasRestoreData(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    bool
	}{
		{"node", `{"node":{"version":"20.0.0"}}`, true},
		{"node", `{"node":{"version":""}}`, false},
		{"node", `{"node":null}`, false},
		{"npm-globals", `{"npm":{"globals":[{"name":"x","version":"1"}]}}`, true},
		{"npm-globals", `{"npm":{"globals":[]}}`, false},
	}
	for _, tt := range tests {
		if got := hasRestoreData(tt.name, json.RawMessage(tt.payload)); got != tt.want {
			t.Errorf("hasRestoreData(%q, %s) = %v, want %v", tt.name, tt.payload, got, tt.want)
		}
	}
}

func TestRestoreSkipsDependentWhenDependencyFails(t *testing.T) {
	node := &fakePlugin{name: "node", fail: true}
	npm := &fakePlugin{name: "npm-globals", deps: []string{"node"}}
	eng := New(plugins.NewRegistry(node, npm))

	results, err := eng.Restore(context.Background(), snapWithNodeAndNPM(), plugins.RestoreOptions{})
	if err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	if r, _ := resultFor(results, "node"); r.Status != plugins.StatusFailed {
		t.Errorf("node status = %q, want failed", r.Status)
	}
	r, _ := resultFor(results, "npm-globals")
	if r.Status != plugins.StatusSkipped || r.Message != "skipped due to failed dependency" {
		t.Errorf("npm-globals = %+v, want skipped due to failed dependency", r)
	}
	if npm.called {
		t.Error("npm-globals.Restore should NOT be called when its dependency failed")
	}
}

func TestRestoreSkipsWhenNothingToRestore(t *testing.T) {
	node := &fakePlugin{name: "node"}
	eng := New(plugins.NewRegistry(node))

	// Empty snapshot → no node data → should skip without calling Restore.
	results, err := eng.Restore(context.Background(), &snapshot.Snapshot{}, plugins.RestoreOptions{})
	if err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}
	r, _ := resultFor(results, "node")
	if r.Status != plugins.StatusSkipped || r.Message != "nothing to restore" {
		t.Errorf("node = %+v, want skipped/nothing to restore", r)
	}
	if node.called {
		t.Error("node.Restore should NOT be called when there is no data")
	}
}

func TestRestorePassesDryRunAndDefaultsTargetOS(t *testing.T) {
	node := &fakePlugin{name: "node"}
	eng := New(plugins.NewRegistry(node))

	_, err := eng.Restore(context.Background(), snapWithNodeAndNPM(), plugins.RestoreOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}
	if !node.called {
		t.Fatal("node.Restore should have been called")
	}
	if !node.gotOpts.DryRun {
		t.Error("DryRun flag was not passed through to the plugin")
	}
	if string(node.gotOpts.TargetOS) != runtime.GOOS {
		t.Errorf("TargetOS = %q, want default %q", node.gotOpts.TargetOS, runtime.GOOS)
	}
}
