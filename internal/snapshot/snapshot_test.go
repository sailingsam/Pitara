package snapshot

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	s := New("test", "linux", "amd64")
	if err := s.Validate(); err != nil {
		t.Fatalf("expected valid snapshot: %v", err)
	}

	s.Machine.OS = ""
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for missing os")
	}
}

func TestParse(t *testing.T) {
	raw := []byte(`{
		"schemaVersion": 1,
		"createdAt": "2026-06-05T10:00:00Z",
		"machine": {"os": "linux", "arch": "amd64"},
		"languages": {},
		"packages": {}
	}`)

	s, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !s.CreatedAt.Equal(time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected createdAt: %v", s.CreatedAt)
	}
}
