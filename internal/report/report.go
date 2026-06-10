package report

import (
	"fmt"
	"strings"

	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/snapshot"
)

func FormatRestore(snap *snapshot.Snapshot, results []plugins.RestoreResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Pitara Restore Report\n")
	fmt.Fprintf(&b, "───────────────────────\n")
	if snap.Machine.Label != "" {
		fmt.Fprintf(&b, "Snapshot:  %s (%s, %s)\n", snap.Machine.Label, snap.Machine.OS, snap.Machine.Arch)
	} else {
		fmt.Fprintf(&b, "Snapshot:  local (%s, %s)\n", snap.Machine.OS, snap.Machine.Arch)
	}
	fmt.Fprintf(&b, "Created:   %s\n\n", snap.CreatedAt.Format("2006-01-02 15:04 MST"))

	warnings := 0
	failures := 0

	for _, result := range results {
		if result.Status == plugins.StatusSkipped && result.Message == "nothing to restore" {
			continue
		}

		fmt.Fprintf(&b, "%s\n", titleFor(result.PluginName))
		if len(result.Details) > 0 {
			for _, line := range result.Details {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		} else {
			icon := iconFor(result.Status)
			fmt.Fprintf(&b, "  %s %s\n", icon, result.Message)
		}
		fmt.Fprintln(&b)

		if result.Status == plugins.StatusFailed {
			failures++
		}
		if result.Status == plugins.StatusSkipped && result.Message != "nothing to restore" {
			warnings++
		}
	}

	switch {
	case failures > 0:
		fmt.Fprintf(&b, "Restore completed with %d failure(s).\n", failures)
	case warnings > 0:
		fmt.Fprintf(&b, "Restore completed with %d warning(s).\n", warnings)
	default:
		b.WriteString("Restore completed.\n")
	}

	return b.String()
}

func titleFor(pluginName string) string {
	switch pluginName {
	case "node":
		return "Runtimes"
	case "npm-globals":
		return "Global Packages (npm)"
	default:
		return pluginName
	}
}

func iconFor(status plugins.Status) string {
	switch status {
	case plugins.StatusSuccess:
		return "✓"
	case plugins.StatusSkipped:
		return "⚠"
	case plugins.StatusFailed:
		return "✗"
	default:
		return "·"
	}
}
