package plugins

import (
	"context"
	"encoding/json"
)

type OS string

const (
	OSDarwin  OS = "darwin"
	OSLinux   OS = "linux"
	OSWindows OS = "windows"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusSkipped Status = "skipped"
	StatusFailed  Status = "failed"
)

type ScanResult struct {
	PluginName string
	Data       json.RawMessage
	Warnings   []string
}

type RestoreResult struct {
	PluginName string
	Status     Status
	Message    string
	Details    []string
}

type RestoreOptions struct {
	DryRun     bool
	TargetOS   OS
	TargetArch string
}

type Plugin interface {
	Name() string
	Scan(ctx context.Context) (ScanResult, error)
	Restore(ctx context.Context, snap json.RawMessage, opts RestoreOptions) (RestoreResult, error)
	Dependencies() []string
	SupportedOS() []OS
}
