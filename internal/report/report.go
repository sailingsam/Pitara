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
	lastTitle := ""

	for _, result := range results {
		if result.Status == plugins.StatusSkipped && result.Message == "nothing to restore" {
			continue
		}

		title := titleFor(result.PluginName)
		if title != lastTitle {
			if lastTitle != "" {
				fmt.Fprintln(&b)
			}
			fmt.Fprintf(&b, "%s\n", title)
			lastTitle = title
		}
		if len(result.Details) > 0 {
			for _, line := range result.Details {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		} else {
			icon := iconFor(result.Status)
			fmt.Fprintf(&b, "  %s %s\n", icon, result.Message)
		}

		if result.Status == plugins.StatusFailed {
			failures++
		}
		if result.Status == plugins.StatusSkipped && result.Message != "nothing to restore" {
			warnings++
		}
	}

	fmt.Fprintln(&b)
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
	case "node", "go", "java", "bun", "deno", "python":
		return "Runtimes"
	case "npm-globals":
		return "Global Packages (npm)"
	case "pnpm-globals":
		return "Global Packages (pnpm)"
	case "bun-globals":
		return "Global Packages (bun)"
	case "deno-globals":
		return "Global Packages (deno)"
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
