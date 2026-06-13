// Package github is the CLI's direct client for the GitHub REST API and the
// OAuth device-login flow. It uses only the standard library, so the shipped
// `pitara` binary stays dependency-light and never routes through any server
// of ours — the token lives only on the user's machine and talks to GitHub.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// defaultClientID is the PUBLIC OAuth App client id for Pitara. The device flow
// needs no secret, so shipping this in the binary is safe — and it routes every
// user's login through the project's OAuth App (useful as a usage signal).
// Override with PITARA_GITHUB_CLIENT_ID to point at your own app.
const defaultClientID = "Ov23liRj6GK3vHRdCm8G"

// scope: full repo access is required to create the private snapshots repo and
// write file contents. Mitigated by being open source + no server + easy revoke.
const scope = "repo"

const (
	deviceCodeURL  = "https://github.com/login/device/code"
	accessTokenURL = "https://github.com/login/oauth/access_token"
)

func clientID() string {
	if v := os.Getenv("PITARA_GITHUB_CLIENT_ID"); v != "" {
		return v
	}
	return defaultClientID
}

// DeviceStart is what the user must approve in their browser.
type DeviceStart struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// ErrPending means the user has not yet approved the device code.
var ErrPending = fmt.Errorf("authorization pending")

// ErrSlowDown means GitHub asked us to poll less often.
var ErrSlowDown = fmt.Errorf("slow down")

// StartDeviceLogin begins the GitHub device-authorization flow.
func StartDeviceLogin(ctx context.Context) (*DeviceStart, error) {
	form := url.Values{"client_id": {clientID()}, "scope": {scope}}
	var out DeviceStart
	if err := postForm(ctx, deviceCodeURL, form, &out); err != nil {
		return nil, err
	}
	if out.DeviceCode == "" {
		return nil, fmt.Errorf("github did not return a device code (is PITARA_GITHUB_CLIENT_ID set to a valid OAuth App?)")
	}
	return &out, nil
}

// PollDeviceLogin checks once whether the user has approved. Returns the access
// token on success, ErrPending/ErrSlowDown while waiting.
func PollDeviceLogin(ctx context.Context, deviceCode string) (string, error) {
	form := url.Values{
		"client_id":   {clientID()},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}
	var out struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := postForm(ctx, accessTokenURL, form, &out); err != nil {
		return "", err
	}

	switch out.Error {
	case "":
		if out.AccessToken == "" {
			return "", fmt.Errorf("github returned no token")
		}
		return out.AccessToken, nil
	case "authorization_pending":
		return "", ErrPending
	case "slow_down":
		return "", ErrSlowDown
	case "expired_token":
		return "", fmt.Errorf("the code expired; please run `pitara login` again")
	case "access_denied":
		return "", fmt.Errorf("authorization was denied")
	default:
		if out.ErrorDescription != "" {
			return "", fmt.Errorf("github: %s", out.ErrorDescription)
		}
		return "", fmt.Errorf("github: %s", out.Error)
	}
}

func postForm(ctx context.Context, endpoint string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("contact github: %w", err)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
