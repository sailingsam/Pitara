package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sailingsam/pitara/internal/app"
	"github.com/sailingsam/pitara/internal/auth"
	"github.com/sailingsam/pitara/internal/github"
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/report"
	"github.com/sailingsam/pitara/internal/restore"
	"github.com/sailingsam/pitara/internal/snapshot"
	"github.com/spf13/cobra"
)

func newRestoreCmd() *cobra.Command {
	var from string
	var label string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore development tools from a backup",
		Long: "Restore from your GitHub backup (default: this machine's latest), a named " +
			"machine (--label), or a local file (--from).",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			data, err := loadSnapshotData(ctx, from, label)
			if err != nil {
				return err
			}

			snap, err := snapshot.Parse(data)
			if err != nil {
				return err
			}

			results, err := restore.New(app.DefaultRegistry()).
				Restore(ctx, snap, plugins.RestoreOptions{DryRun: dryRun})
			if err != nil {
				return err
			}

			fmt.Print(report.FormatRestore(snap, results))
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Restore from a local snapshot JSON file")
	cmd.Flags().StringVar(&label, "label", "", "Restore a specific machine's backup")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without making changes")
	return cmd
}

// loadSnapshotData resolves the snapshot bytes: a local file (--from) or the
// GitHub backup for a label (default label when none given; requires login).
func loadSnapshotData(ctx context.Context, from, label string) (json.RawMessage, error) {
	if from != "" {
		data, err := os.ReadFile(from)
		if err != nil {
			return nil, fmt.Errorf("read snapshot: %w", err)
		}
		return data, nil
	}

	creds, err := auth.Load()
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return nil, fmt.Errorf("not logged in: use --from <file> for a local snapshot, or run `pitara login`")
	}
	if label == "" {
		label = github.DefaultLabel
	}
	return github.NewStore(creds.AccessToken, creds.Login).Load(ctx, label)
}
