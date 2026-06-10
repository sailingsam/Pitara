package plugins

import "testing"

func TestIsSelf(t *testing.T) {
	cases := map[string]bool{
		"pitara":             true,
		"@sailingsam/pitara": true,
		"typescript":         false,
		"npm":                false,
		"":                   false,
	}
	for name, want := range cases {
		if got := IsSelf(name); got != want {
			t.Errorf("IsSelf(%q) = %v, want %v", name, got, want)
		}
	}
}
