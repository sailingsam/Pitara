package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const apiBase = "https://api.github.com"

// ErrNotFound is returned when a repo or file does not exist (HTTP 404).
var ErrNotFound = fmt.Errorf("not found")

// Client is an authenticated GitHub REST client.
type Client struct {
	token string
	http  *http.Client
}

func NewClient(token string) *Client {
	return &Client{token: token, http: &http.Client{Timeout: 30 * time.Second}}
}

// User is the authenticated GitHub user.
type User struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CurrentUser returns the authenticated user (used for whoami and repo owner).
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var u User
	if err := c.do(ctx, http.MethodGet, "/user", nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// RepoExists reports whether owner/repo exists and is accessible.
func (c *Client) RepoExists(ctx context.Context, owner, repo string) (bool, error) {
	err := c.do(ctx, http.MethodGet, "/repos/"+owner+"/"+repo, nil, nil)
	if err == nil {
		return true, nil
	}
	if err == ErrNotFound {
		return false, nil
	}
	return false, err
}

// CreateRepo creates a repository for the authenticated user.
func (c *Client) CreateRepo(ctx context.Context, name string, private bool, description string) error {
	body := map[string]any{
		"name":        name,
		"private":     private,
		"description": description,
		"auto_init":   false,
	}
	return c.do(ctx, http.MethodPost, "/user/repos", body, nil)
}

// FileEntry is one item from a repo directory listing.
type FileEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

// ListDir lists the contents of a repo directory (path "" = root).
func (c *Client) ListDir(ctx context.Context, owner, repo, path string) ([]FileEntry, error) {
	var entries []FileEntry
	if err := c.do(ctx, http.MethodGet, contentsPath(owner, repo, path), nil, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetFile returns a file's decoded content and its blob SHA (needed to update).
func (c *Client) GetFile(ctx context.Context, owner, repo, path string) (content []byte, sha string, err error) {
	var out struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
		SHA      string `json:"sha"`
	}
	if err := c.do(ctx, http.MethodGet, contentsPath(owner, repo, path), nil, &out); err != nil {
		return nil, "", err
	}
	if out.Encoding != "base64" {
		return nil, "", fmt.Errorf("unexpected encoding %q", out.Encoding)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(out.Content, "\n", ""))
	if err != nil {
		return nil, "", fmt.Errorf("decode file: %w", err)
	}
	return decoded, out.SHA, nil
}

// PutFile creates or updates a file with a commit. Pass the existing sha to
// update; leave it empty to create.
func (c *Client) PutFile(ctx context.Context, owner, repo, path string, content []byte, message, sha string) error {
	body := map[string]any{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
	}
	if sha != "" {
		body["sha"] = sha
	}
	return c.do(ctx, http.MethodPut, contentsPath(owner, repo, path), body, nil)
}

func contentsPath(owner, repo, path string) string {
	return fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "pitara-cli")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("github request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return githubError(resp.StatusCode, raw)
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode github response: %w", err)
		}
	}
	return nil
}

func githubError(status int, raw []byte) error {
	var e struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &e) == nil && e.Message != "" {
		return fmt.Errorf("github error (%d): %s", status, e.Message)
	}
	return fmt.Errorf("github error (%d)", status)
}
