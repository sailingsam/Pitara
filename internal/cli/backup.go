package cli

import (
	"fmt"
	"os"

	"github.com/sailingsam/pitara/internal/app"
	"github.com/sailingsam/pitara/internal/auth"
	"github.com/sailingsam/pitara/internal/discovery"
	"github.com/sailingsam/pitara/internal/github"
	"github.com/spf13/cobra"
)

func newBackupCmd() *cobra.Command {
	var label string

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Scan this machine and back it up to your GitHub",
		Long: "Requires login. Scans the machine and commits the snapshot to a private " +
			"`pitara-snapshots` repo in your own GitHub account (created automatically on " +
			"first use). With no --label it updates default.json; most people never need a label.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			creds, err := auth.Load()
			if err != nil {
				return err
			}
			if creds == nil {
				return fmt.Errorf("not logged in (run `pitara login`)")
			}
			if label == "" {
				label = github.DefaultLabel
			}

			// Scan (label is also recorded inside the snapshot).
			registry := app.DefaultRegistry()
			snap, warnings, err := discovery.New(registry).Scan(ctx, label)
			if err != nil {
				return err
			}
			for _, w := range warnings {
				fmt.Fprintf(os.Stderr, "warning: %s\n", w)
			}
			data, err := snap.JSON()
			if err != nil {
				return err
			}

			store := github.NewStore(creds.AccessToken, creds.Login)
			if err := store.EnsureRepo(ctx); err != nil {
				return fmt.Errorf("prepare snapshots repo: %w", err)
			}
			machine := fmt.Sprintf("%s/%s", snap.Machine.OS, snap.Machine.Arch)
			if err := store.Save(ctx, label, data, machine); err != nil {
				return fmt.Errorf("save snapshot: %w", err)
			}

			fmt.Printf("✓ Backed up %q to %s/%s\n", label, creds.Login, github.RepoName)
			return nil
		},
	}

	cmd.Flags().StringVar(&label, "label", "", "Machine label (only needed for multiple machines)")
	return cmd
}
