package snapshot

import (
	"encoding/json"
	"fmt"
	"time"
)

const CurrentSchemaVersion = 1

type Snapshot struct {
	SchemaVersion int       `json:"schemaVersion"`
	ID            string    `json:"id,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	Machine       Machine   `json:"machine"`
	Languages     Languages `json:"languages"`
	Packages      Packages  `json:"packages"`
}

type Machine struct {
	Label string `json:"label,omitempty"`
	OS    string `json:"os"`
	Arch  string `json:"arch"`
}

type Languages struct {
	Node   *Runtime `json:"node,omitempty"`
	Go     *Runtime `json:"go,omitempty"`
	Java   *Java    `json:"java,omitempty"`
	Bun    *Runtime `json:"bun,omitempty"`
	Deno   *Runtime `json:"deno,omitempty"`
	Python *Runtime `json:"python,omitempty"`
}

type Runtime struct {
	Version string `json:"version"`
	Manager string `json:"manager,omitempty"`
}

type Java struct {
	Version      string `json:"version"`
	Distribution string `json:"distribution,omitempty"`
}

type Packages struct {
	NPM  *GlobalPackages `json:"npm,omitempty"`
	PNPM *GlobalPackages `json:"pnpm,omitempty"`
	Bun  *GlobalPackages `json:"bun,omitempty"`
	Deno *DenoGlobals    `json:"deno,omitempty"`
}

type GlobalPackages struct {
	Globals []GlobalPackage `json:"globals"`
}

type GlobalPackage struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// DenoGlobals holds CLIs installed with `deno install -g`. Unlike npm/bun
// globals, a Deno global isn't a versioned package name — it's a module
// specifier plus the permission flags it was granted, so it needs its own type.
type DenoGlobals struct {
	Globals []DenoGlobal `json:"globals"`
}

type DenoGlobal struct {
	Name      string   `json:"name"`
	Specifier string   `json:"specifier"`
	Flags     []string `json:"flags,omitempty"`
}

func New(label, os, arch string) *Snapshot {
	return &Snapshot{
		SchemaVersion: CurrentSchemaVersion,
		CreatedAt:     time.Now().UTC(),
		Machine: Machine{
			Label: label,
			OS:    os,
			Arch:  arch,
		},
		Languages: Languages{},
		Packages:  Packages{},
	}
}

func (s *Snapshot) Validate() error {
	if s.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema version %d (expected %d)", s.SchemaVersion, CurrentSchemaVersion)
	}
	if s.Machine.OS == "" {
		return fmt.Errorf("machine.os is required")
	}
	if s.Machine.Arch == "" {
		return fmt.Errorf("machine.arch is required")
	}
	return nil
}

func Parse(data []byte) (*Snapshot, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse snapshot: %w", err)
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Snapshot) JSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
