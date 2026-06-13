package cli

import (
	"fmt"

	"github.com/sailingsam/pitara/internal/auth"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear the local Pitara session",
		Long: "Removes the stored token from ~/.pitara. To fully revoke access, also " +
			"remove Pitara from https://github.com/settings/applications.",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := auth.Load()
			if err != nil {
				return err
			}
			if creds == nil {
				fmt.Println("Not logged in.")
				return nil
			}
			if err := auth.Clear(); err != nil {
				return err
			}
			fmt.Println("✓ Logged out. (Revoke fully at github.com/settings/applications)")
			return nil
		},
	}
}
