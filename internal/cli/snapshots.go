package cli

import (
	"fmt"

	"github.com/sailingsam/pitara/internal/auth"
	"github.com/sailingsam/pitara/internal/github"
	"github.com/spf13/cobra"
)

func newSnapshotsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Browse your backups on GitHub",
	}
	cmd.AddCommand(newSnapshotsListCmd())
	return cmd
}

func newSnapshotsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your saved machine labels",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := auth.Load()
			if err != nil {
				return err
			}
			if creds == nil {
				return fmt.Errorf("not logged in (run `pitara login`)")
			}

			labels, err := github.NewStore(creds.AccessToken, creds.Login).List(cmd.Context())
			if err != nil {
				return err
			}
			if len(labels) == 0 {
				fmt.Println("No backups yet. Run `pitara backup` to create one.")
				return nil
			}

			fmt.Printf("Backups in %s/%s:\n", creds.Login, github.RepoName)
			for _, label := range labels {
				marker := ""
				if label == github.DefaultLabel {
					marker = "  (this machine, default)"
				}
				fmt.Printf("  • %s%s\n", label, marker)
			}
			return nil
		},
	}
}
