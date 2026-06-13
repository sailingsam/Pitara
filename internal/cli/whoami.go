package cli

import (
	"fmt"

	"github.com/sailingsam/pitara/internal/auth"
	"github.com/sailingsam/pitara/internal/github"
	"github.com/spf13/cobra"
)

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the logged-in GitHub user",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := auth.Load()
			if err != nil {
				return err
			}
			if creds == nil {
				return fmt.Errorf("not logged in (run `pitara login`)")
			}

			user, err := github.NewClient(creds.AccessToken).CurrentUser(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Print(user.Login)
			if user.Name != "" {
				fmt.Printf(" (%s)", user.Name)
			}
			fmt.Println()
			return nil
		},
	}
}
