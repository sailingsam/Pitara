package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// RepoName is the private repository Pitara stores snapshots in.
const RepoName = "pitara-snapshots"

// DefaultLabel is used when the user runs `pitara backup` / `restore` with no
// --label. Most people have one machine and never need a label.
const DefaultLabel = "default"

const readmeContent = "# Pitara Snapshots\n\n" +
	"This repository holds backups of my **development environment** — language\n" +
	"runtimes (Node, Go, Java, Bun) and global CLI packages — captured by\n" +
	"[Pitara](https://github.com/sailingsam/pitara).\n\n" +
	"It is **auto-managed**. Each file is a snapshot; every backup is a new commit,\n" +
	"so the full history is the git log.\n\n" +
	"## Files\n\n" +
	"- `default.json` — this machine's latest snapshot\n" +
	"- `<label>.json` — a named machine (`pitara backup --label <name>`)\n\n" +
	"## Restore on a new machine\n\n" +
	"```bash\n" +
	"pitara restore                  # restore the latest backup\n" +
	"pitara restore --label <name>   # restore a specific machine\n" +
	"```\n\n" +
	"## What's inside\n\n" +
	"Plain JSON: installed runtimes + global packages, with versions.\n" +
	"**No secrets** — no SSH keys, tokens, or environment variables are ever stored.\n\n" +
	"---\n" +
	"Generated and maintained by Pitara. Edit by hand at your own risk.\n"

// Store is the high-level snapshot store backed by the user's GitHub repo.
type Store struct {
	c     *Client
	owner string
}

// NewStore builds a store for the given token and repo owner (GitHub login).
func NewStore(token, owner string) *Store {
	return &Store{c: NewClient(token), owner: owner}
}

// EnsureRepo creates the private snapshots repo (with a README) if it is missing.
func (s *Store) EnsureRepo(ctx context.Context) error {
	exists, err := s.c.RepoExists(ctx, s.owner, RepoName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if err := s.c.CreateRepo(ctx, RepoName, true, "My Pitara dev-environment backups"); err != nil {
		return err
	}
	// Best-effort README so the repo isn't empty and explains itself.
	_ = s.c.PutFile(ctx, s.owner, RepoName, "README.md", []byte(readmeContent), "Initialize Pitara snapshots", "")
	return nil
}

// Save writes a snapshot for a label, creating or updating <label>.json.
func (s *Store) Save(ctx context.Context, label string, data []byte, machine string) error {
	path := fileFor(label)

	// An update needs the existing blob SHA; a create does not.
	_, sha, err := s.c.GetFile(ctx, s.owner, RepoName, path)
	if err != nil && err != ErrNotFound {
		return err
	}

	msg := fmt.Sprintf("pitara backup: %s (%s)", label, machine)
	return s.c.PutFile(ctx, s.owner, RepoName, path, data, msg, sha)
}

// Load returns the latest snapshot JSON for a label.
func (s *Store) Load(ctx context.Context, label string) ([]byte, error) {
	data, _, err := s.c.GetFile(ctx, s.owner, RepoName, fileFor(label))
	if err == ErrNotFound {
		return nil, fmt.Errorf("no backup found for %q (run `pitara backup`)", label)
	}
	return data, err
}

// List returns the labels (snapshot file names without .json) in the repo.
func (s *Store) List(ctx context.Context) ([]string, error) {
	entries, err := s.c.ListDir(ctx, s.owner, RepoName, "")
	if err == ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var labels []string
	for _, e := range entries {
		if e.Type == "file" && strings.HasSuffix(e.Name, ".json") {
			labels = append(labels, strings.TrimSuffix(e.Name, ".json"))
		}
	}
	sort.Strings(labels)
	return labels, nil
}

// fileFor maps a label to its snapshot file path.
func fileFor(label string) string {
	if label == "" {
		label = DefaultLabel
	}
	return label + ".json"
}

// CommitStamp is a small helper for human-readable commit context.
func CommitStamp(t time.Time) string { return t.UTC().Format("2006-01-02 15:04 MST") }
