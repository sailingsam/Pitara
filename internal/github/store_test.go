package github

import "testing"

func TestFileFor(t *testing.T) {
	cases := map[string]string{
		"":             "default.json", // no label → default
		"default":      "default.json",
		"work-laptop":  "work-laptop.json",
		"office-pc":    "office-pc.json",
	}
	for label, want := range cases {
		if got := fileFor(label); got != want {
			t.Errorf("fileFor(%q) = %q, want %q", label, got, want)
		}
	}
}

func TestReadmeMentionsSafety(t *testing.T) {
	// The auto-generated README must reassure users no secrets are stored.
	if want := "No secrets"; !contains(readmeContent, want) {
		t.Errorf("README missing %q reassurance", want)
	}
	if want := "pitara restore"; !contains(readmeContent, want) {
		t.Errorf("README missing restore instructions")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
