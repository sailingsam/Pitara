package java

import "testing"

func TestParseVersion(t *testing.T) {
	// `java -version` banner (printed to stderr).
	out := `openjdk version "17.0.19" 2026-04-21
OpenJDK Runtime Environment (build 17.0.19+10-1-24.04.2-Ubuntu)`
	if got := parseVersion(out); got != "17.0.19" {
		t.Errorf("parseVersion = %q, want 17.0.19", got)
	}

	legacy := `java version "1.8.0_392"`
	if got := parseVersion(legacy); got != "1.8.0_392" {
		t.Errorf("parseVersion legacy = %q, want 1.8.0_392", got)
	}
}

func TestMajorVersion(t *testing.T) {
	cases := map[string]string{
		"17.0.19":   "17",
		"21.0.2":    "21",
		"1.8.0_392": "8",
		"11":        "11",
	}
	for in, want := range cases {
		if got := majorVersion(in); got != want {
			t.Errorf("majorVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDetectDistribution(t *testing.T) {
	if got := detectDistribution(`openjdk version "17.0.19"`); got != "openjdk" {
		t.Errorf("detectDistribution = %q, want openjdk", got)
	}
	if got := detectDistribution(`Temurin-17`); got != "temurin" {
		t.Errorf("detectDistribution = %q, want temurin", got)
	}
}
