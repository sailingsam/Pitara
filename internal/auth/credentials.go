// Package auth stores and loads the local Pitara session.
// The session is a GitHub access token (from the device-login flow) plus the
// user's login, kept in ~/.pitara/credentials.json with 0600 permissions.
// Nothing from scanned machines is ever written here.
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials is the locally stored GitHub session.
type Credentials struct {
	AccessToken string `json:"accessToken"`
	Login       string `json:"login,omitempty"` // GitHub username (repo owner)
}

// Dir returns ~/.pitara, honouring PITARA_HOME for tests/overrides.
func Dir() (string, error) {
	if override := os.Getenv("PITARA_HOME"); override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, ".pitara"), nil
}

func credentialsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// Save writes credentials to disk with restrictive permissions.
func Save(c *Credentials) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	return nil
}

// Load reads credentials from disk. Returns (nil, nil) when not logged in.
func Load() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read credentials: %w", err)
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	return &c, nil
}

// Clear removes the stored credentials. Missing file is not an error.
func Clear() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials: %w", err)
	}
	return nil
}
