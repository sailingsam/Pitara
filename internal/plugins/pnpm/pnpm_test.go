package pnpm

import (
	"testing"

	"github.com/sailingsam/pitara/internal/snapshot"
)

func TestParsePNPMList(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    []snapshot.GlobalPackage
		wantErr bool
	}{
		{ // empty json array
			name:   "empty json array",
			output: `[]`,
			want:   []snapshot.GlobalPackage{},
		},
		{ // single project with dependencies
			name: "single project single dependencies",
			output: `[
				{
					"dependencies":{
						"typescript": {"version":"5.4.5"},
						"prettier": {"version":"3.2.1"}
					}	
				}	
			]`,
			want: []snapshot.GlobalPackage{
				{Name: "typescript", Version: "5.4.5"},
				{Name: "prettier", Version: "3.2.1"},
			},
		},
		{ // multiple projects with dependencies
			name: "multiple projects with multiple dependencies",
			output: `[
				{
					"dependencies":{
					"eslint": {"version":"8.57.0"}
					}
				},
				{
					"dependencies":{
						"typescript": {"version":"5.4.3"},
						"pnpm": {"version":"9.0.0"}
					}	
				}	
			]`,
			want: []snapshot.GlobalPackage{
				{Name: "eslint", Version: "8.57.0"},
				{Name: "typescript", Version: "5.4.3"},
				{Name: "pnpm", Version: "9.0.0"},
			},
		},
		{ // project with no dependencies
			name: "single project with no dependencies",
			output: `[
				{
					"name": "prettier"	
				}	
			]`,
			want: []snapshot.GlobalPackage{},
		},
		{ // version with v prefix is stripped
			name: "single project with v prefix versioning",
			output: `[
				{
					"dependencies":{
						"node-sass": {"version":"v4.14.2"}	
					}	
				}	
			]`,
			want: []snapshot.GlobalPackage{
				{Name: "node-sass", Version: "4.14.2"},
			},
		},
		{ //self packages are excluded
			name: "self packages project",
			output: `[
				{
					"dependencies":{
						"pitara":	{"version":"0.5.0"},
						"@sailingsam/pitara": {"version": "0.5.0"},
						"typescript": {"version": "5.4.5"}
					}	
				}	
			]`,
			want: []snapshot.GlobalPackage{

				{Name: "typescript", Version: "5.4.5"},
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
			got, err := parsePNPMList(tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: got %v, wanted %v as error", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d packages, wanted %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("got %s as package name, wanted %s", got[i].Name, tt.want[i].Name)
				}
				if got[i].Version != tt.want[i].Version {
					t.Errorf("got %s as package name, wanted %s", got[i].Version, tt.want[i].Version)
				}
			}
		})
	}
}
