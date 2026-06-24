package cargoglobals

import (
	"sort"
	"testing"

	"github.com/sailingsam/pitara/internal/snapshot"
)

func TestParseCargoList(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []snapshot.GlobalPackage
	}{
		{
			name:   "empty",
			output: "",
			want:   []snapshot.GlobalPackage{},
		},
		{
			name:   "single crate",
			output: "ripgrep v15.1.0:\n    rg\n",
			want:   []snapshot.GlobalPackage{{Name: "ripgrep", Version: "15.1.0"}},
		},
		{
			name: "multiple crates, multiple binaries",
			output: "cargo-edit v0.12.0:\n    cargo-add\n    cargo-rm\n" +
				"ripgrep v15.1.0:\n    rg\n",
			want: []snapshot.GlobalPackage{
				{Name: "cargo-edit", Version: "0.12.0"},
				{Name: "ripgrep", Version: "15.1.0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCargoList(tt.output)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d crates, want %d: %+v", len(got), len(tt.want), got)
			}
			sort.Slice(got, func(i, j int) bool { return got[i].Name < got[j].Name })
			sort.Slice(tt.want, func(i, j int) bool { return tt.want[i].Name < tt.want[j].Name })
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %+v, want %+v", got[i], tt.want[i])
				}
			}
		})
	}
}
