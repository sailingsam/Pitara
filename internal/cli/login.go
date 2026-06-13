package cli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/sailingsam/pitara/internal/auth"
	"github.com/sailingsam/pitara/internal/github"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in with GitHub",
		Long: "Authorizes Pitara via GitHub's device flow and stores a token locally " +
			"(~/.pitara). Snapshots are saved to a private repo in your own GitHub account; " +
			"nothing is sent to any Pitara server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			start, err := github.StartDeviceLogin(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("\nTo authorize Pitara, open:\n  %s\n\n", start.VerificationURI)
			fmt.Printf("And enter the code:  %s\n\n", start.UserCode)
			if openBrowser(start.VerificationURI) {
				fmt.Println("(opened your browser automatically)")
			}
			fmt.Println("Waiting for authorization...")

			tokenStr, err := pollUntilAuthorized(ctx, start)
			if err != nil {
				return err
			}

			// Fetch the login so we know the repo owner for later commands.
			user, err := github.NewClient(tokenStr).CurrentUser(ctx)
			if err != nil {
				return fmt.Errorf("verify token: %w", err)
			}
			if err := auth.Save(&auth.Credentials{AccessToken: tokenStr, Login: user.Login}); err != nil {
				return err
			}

			fmt.Printf("\n✓ Logged in as %s\n", user.Login)
			return nil
		},
	}
}

func pollUntilAuthorized(ctx context.Context, start *github.DeviceStart) (string, error) {
	interval := time.Duration(maxInt(start.Interval, 5)) * time.Second
	deadline := time.Now().Add(time.Duration(maxInt(start.ExpiresIn, 600)) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}

		tokenStr, err := github.PollDeviceLogin(ctx, start.DeviceCode)
		if err == nil {
			return tokenStr, nil
		}
		switch {
		case errors.Is(err, github.ErrPending):
			// keep waiting
		case errors.Is(err, github.ErrSlowDown):
			interval += 5 * time.Second
		default:
			return "", err
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("login timed out; please run `pitara login` again")
		}
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// openBrowser best-effort opens a URL. Returns false if it could not start a
// command (the URL is always printed regardless).
func openBrowser(url string) bool {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	if _, err := exec.LookPath(cmd); err != nil {
		return false
	}
	return exec.Command(cmd, args...).Start() == nil
}
