package app

import (
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/plugins/bun"
	"github.com/sailingsam/pitara/internal/plugins/bunglobals"
	"github.com/sailingsam/pitara/internal/plugins/cargoglobals"
	"github.com/sailingsam/pitara/internal/plugins/deno"
	"github.com/sailingsam/pitara/internal/plugins/denoglobals"
	"github.com/sailingsam/pitara/internal/plugins/golang"
	"github.com/sailingsam/pitara/internal/plugins/java"
	"github.com/sailingsam/pitara/internal/plugins/node"
	"github.com/sailingsam/pitara/internal/plugins/npm"
	"github.com/sailingsam/pitara/internal/plugins/pipx"
	"github.com/sailingsam/pitara/internal/plugins/pnpm"
	"github.com/sailingsam/pitara/internal/plugins/python"
	"github.com/sailingsam/pitara/internal/plugins/rust"
	"github.com/sailingsam/pitara/internal/plugins/yarn"
)

func DefaultRegistry() *plugins.Registry {
	return plugins.NewRegistry(
		node.New(),
		golang.New(),
		java.New(),
		bun.New(),
		deno.New(),
		python.New(),
		rust.New(),
		npm.New(),
		pnpm.New(),
		bunglobals.New(),
		denoglobals.New(),
		yarn.New(),
		cargoglobals.New(),
		pipx.New(),
	)
}
