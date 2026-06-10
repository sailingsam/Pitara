package cli

import (
	"fmt"
	"os"

	"github.com/sailingsam/pitara/internal/app"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/report"
	"github.com/sailingsam/pitara/internal/restore"
	"github.com/sailingsam/pitara/internal/snapshot"
	"github.com/spf13/cobra"
)

func newRestoreCmd() *cobra.Command {
	var from string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore development tools from a snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if from == "" {
				return fmt.Errorf("--from is required (cloud restore coming in Phase 3)")
			}

			data, err := os.ReadFile(from)
			if err != nil {
				return fmt.Errorf("read snapshot: %w", err)
			}

			snap, err := snapshot.Parse(data)
			if err != nil {
				return err
			}

			registry := app.DefaultRegistry()
			engine := restore.New(registry)

			results, err := engine.Restore(cmd.Context(), snap, plugins.RestoreOptions{
				DryRun: dryRun,
			})
			if err != nil {
				return err
			}

			fmt.Print(report.FormatRestore(snap, results))
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Path to local snapshot JSON file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without making changes")
	return cmd
}
