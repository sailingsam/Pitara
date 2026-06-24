package rust

import "testing"

func TestParseRustVersion(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
	}{
		{"standard", "rustc 1.78.0 (9b00956e5 2024-04-29)", "1.78.0"},
		{"newer", "rustc 1.96.0 (ac68faa20 2026-05-25)", "1.96.0"},
		{"trailing newline", "rustc 1.80.1\n", "1.80.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRustVersion(tt.out); got != tt.want {
				t.Errorf("parseRustVersion(%q) = %q, want %q", tt.out, got, tt.want)
			}
		})
	}
}
