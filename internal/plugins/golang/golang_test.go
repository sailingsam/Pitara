package golang

import "testing"

func TestParseGoVersion(t *testing.T) {
	cases := map[string]string{
		"go version go1.23.5 linux/amd64":  "1.23.5",
		"go version go1.21.0 darwin/arm64": "1.21.0",
	}
	for in, want := range cases {
		if got := parseGoVersion(in); got != want {
			t.Errorf("parseGoVersion(%q) = %q, want %q", in, got, want)
		}
	}
}
