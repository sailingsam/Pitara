package executil

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s %v: %s", name, args, msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RunCombined runs a command and returns its combined stdout+stderr, trimmed.
// Some tools (notably `java -version`) print their version banner to stderr even
// on success, so callers that need that output use this instead of Run.
func RunCombined(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	if err := cmd.Run(); err != nil {
		out := strings.TrimSpace(combined.String())
		if out == "" {
			out = err.Error()
		}
		return "", fmt.Errorf("%s %v: %s", name, args, out)
	}
	return strings.TrimSpace(combined.String()), nil
}

func Available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
