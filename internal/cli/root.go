package cli

import (
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "pitara",
		Short: "Backup and restore your development environment",
		Long:  "Pitara scans your machine for language runtimes and global CLI tools, then restores them on a new machine.",
	}

	root.AddCommand(newScanCmd())
	root.AddCommand(newRestoreCmd())

	return root
}
