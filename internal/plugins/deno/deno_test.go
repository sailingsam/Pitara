package deno

import "testing"

func TestParseDenoVersion(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
	}{
		{
			name: "deno 2.x multi-line",
			out:  "deno 2.8.3 (stable, release, x86_64-unknown-linux-gnu)\nv8 14.9.207.2-rusty\ntypescript 6.0.3",
			want: "2.8.3",
		},
		{
			name: "deno 1.x",
			out:  "deno 1.43.0 (release, x86_64-apple-darwin)\nv8 12.4.254.13\ntypescript 5.4.3",
			want: "1.43.0",
		},
		{
			name: "trailing newline",
			out:  "deno 2.0.0\n",
			want: "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDenoVersion(tt.out); got != tt.want {
				t.Errorf("parseDenoVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
