package yarn

import (
	"sort"
	"testing"

	"github.com/sailingsam/pitara/internal/snapshot"
)

func TestParseYarnList(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []snapshot.GlobalPackage
	}{
		{
			name:   "empty",
			output: "yarn global v1.22.22\nDone in 0.05s.\n",
			want:   []snapshot.GlobalPackage{},
		},
		{
			name: "single tool",
			output: "yarn global v1.22.22\n" +
				"info \"tiny-cli@1.1.5\" has binaries:\n   - tiny\nDone in 0.05s.\n",
			want: []snapshot.GlobalPackage{{Name: "tiny-cli", Version: "1.1.5"}},
		},
		{
			name: "multiple + scoped package",
			output: "yarn global v1.22.22\n" +
				"info \"create-react-app@5.0.1\" has binaries:\n   - create-react-app\n" +
				"info \"@angular/cli@17.3.0\" has binaries:\n   - ng\n",
			want: []snapshot.GlobalPackage{
				{Name: "create-react-app", Version: "5.0.1"},
				{Name: "@angular/cli", Version: "17.3.0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseYarnList(tt.output)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d packages, want %d", len(got), len(tt.want))
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
