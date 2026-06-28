package pipx

import (
	"sort"
	"testing"

	"github.com/sailingsam/pitara/internal/snapshot"
)

func TestParsePipxList(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    []snapshot.GlobalPackage
		wantErr bool
	}{
		{
			name:   "empty venvs",
			output: `{"venvs":{}}`,
			want:   []snapshot.GlobalPackage{},
		},
		{
			name: "single tool",
			output: `{"venvs":{
				"black":{"metadata":{"main_package":{"package":"black","package_version":"24.4.2"}}}
			}}`,
			want: []snapshot.GlobalPackage{
				{Name: "black", Version: "24.4.2"},
			},
		},
		{
			name: "multiple tools",
			output: `{"venvs":{
				"black":{"metadata":{"main_package":{"package":"black","package_version":"24.4.2"}}},
				"httpie":{"metadata":{"main_package":{"package":"httpie","package_version":"3.2.2"}}},
				"poetry":{"metadata":{"main_package":{"package":"poetry","package_version":"1.8.3"}}}
			}}`,
			want: []snapshot.GlobalPackage{
				{Name: "black", Version: "24.4.2"},
				{Name: "httpie", Version: "3.2.2"},
				{Name: "poetry", Version: "1.8.3"},
			},
		},
		{
			name: "self and pipx are skipped",
			output: `{"venvs":{
				"pipx":{"metadata":{"main_package":{"package":"pipx","package_version":"1.5.0"}}},
				"pitara":{"metadata":{"main_package":{"package":"pitara","package_version":"0.5.0"}}},
				"black":{"metadata":{"main_package":{"package":"black","package_version":"24.4.2"}}}
			}}`,
			want: []snapshot.GlobalPackage{
				{Name: "black", Version: "24.4.2"},
			},
		},
		{
			name:    "invalid json",
			output:  `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePipxList(tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: got %v, wanted %v as error", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d packages, wanted %d", len(got), len(tt.want))
			}

			// parsePipxList ranges over a map (the "venvs" object), so output
			// order is not deterministic — sort both sides before comparing.
			sort.Slice(got, func(i, j int) bool { return got[i].Name < got[j].Name })
			sort.Slice(tt.want, func(i, j int) bool { return tt.want[i].Name < tt.want[j].Name })

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %+v, wanted %+v", got[i], tt.want[i])
				}
			}
		})
	}
}
