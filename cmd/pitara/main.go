package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sailingsam/pitara/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cli.NewRoot()
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
