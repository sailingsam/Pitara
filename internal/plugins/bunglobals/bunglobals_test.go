package bunglobals

import "testing"

func TestParseBunList(t *testing.T) {
	out := `/home/sam/.bun/install/global node_modules (220)
├── @dotenvx/dotenvx@1.51.4
└── repomix@1.11.1`

	got := parseBunList(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 packages, got %d: %+v", len(got), got)
	}
	if got[0].Name != "@dotenvx/dotenvx" || got[0].Version != "1.51.4" {
		t.Errorf("scoped package parsed wrong: %+v", got[0])
	}
	if got[1].Name != "repomix" || got[1].Version != "1.11.1" {
		t.Errorf("plain package parsed wrong: %+v", got[1])
	}
}

func TestParseBunListEmpty(t *testing.T) {
	if got := parseBunList("/path node_modules (0)"); len(got) != 0 {
		t.Errorf("expected 0 packages, got %+v", got)
	}
}

func TestSplitNameVersion(t *testing.T) {
	cases := []struct{ in, name, version string }{
		{"repomix@1.11.1", "repomix", "1.11.1"},
		{"@scope/pkg@2.0.0", "@scope/pkg", "2.0.0"},
		{"noversion", "noversion", ""},
	}
	for _, c := range cases {
		name, version := splitNameVersion(c.in)
		if name != c.name || version != c.version {
			t.Errorf("splitNameVersion(%q) = (%q,%q), want (%q,%q)", c.in, name, version, c.name, c.version)
		}
	}
}
