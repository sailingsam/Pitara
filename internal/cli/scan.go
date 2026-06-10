package cli

import (
	"fmt"
	"os"

	"github.com/sailingsam/pitara/internal/app"
	"github.com/sailingsam/pitara/internal/discovery"
	"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
	var output string
	var label string

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan this machine and output a snapshot",
		Long:  "Discover installed runtimes and global packages. Prints JSON to stdout or saves to a file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := app.DefaultRegistry()
			engine := discovery.New(registry)

			snap, warnings, err := engine.Scan(cmd.Context(), label)
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

			if output != "" {
				if err := os.WriteFile(output, data, 0o644); err != nil {
					return fmt.Errorf("write snapshot: %w", err)
				}
				fmt.Fprintf(os.Stderr, "snapshot written to %s\n", output)
				return nil
			}

			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Write snapshot to file instead of stdout")
	cmd.Flags().StringVar(&label, "label", "", "Machine label to include in snapshot")
	return cmd
}
