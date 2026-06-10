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

func Available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
