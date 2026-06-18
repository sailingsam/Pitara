package python

import "testing"

func TestParsePythonVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "regular python3",
			input: "python 3.12.3\n",
			want:  "3.12.3",
		},
		{
			name:  "python version 2",
			input: "python 2.7",
			want:  "2.7",
		},
		{
			name:  "version only",
			input: "3.10.2",
			want:  "3.10.2",
		},
		{
			name:  "white space",
			input: "python 3.12.3        \n ",
			want:  "3.12.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePythonVersion(tt.input)
			if got != tt.want {
				t.Errorf(
					"parsePythonVersion(%q) = %q but wanted %q",
					tt.input,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestMajorMinor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"full version", "3.12.3", "3.12"},
		{"already major.minor", "3.12", "3.12"},
		{"single num", "3", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := majorMinor(tt.input)

			if got != tt.want {
				t.Errorf(
					"majorMinor(%q) = %q but wanted %q",
					tt.input,
					got,
					tt.want,
				)
			}
		})
	}

}
