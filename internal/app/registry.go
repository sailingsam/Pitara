package app

import (
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/plugins/bun"
	"github.com/sailingsam/pitara/internal/plugins/bunglobals"
	"github.com/sailingsam/pitara/internal/plugins/golang"
	"github.com/sailingsam/pitara/internal/plugins/java"
	"github.com/sailingsam/pitara/internal/plugins/node"
	"github.com/sailingsam/pitara/internal/plugins/npm"
	"github.com/sailingsam/pitara/internal/plugins/pnpm"
)

func DefaultRegistry() *plugins.Registry {
	return plugins.NewRegistry(
		node.New(),
		golang.New(),
		java.New(),
		bun.New(),
		npm.New(),
		pnpm.New(),
		bunglobals.New(),
	)
}
