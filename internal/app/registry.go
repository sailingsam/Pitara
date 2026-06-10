package app

import (
	"github.com/sailingsam/pitara/internal/plugins"
	"github.com/sailingsam/pitara/internal/plugins/node"
	"github.com/sailingsam/pitara/internal/plugins/npm"
)

func DefaultRegistry() *plugins.Registry {
	return plugins.NewRegistry(
		node.New(),
		npm.New(),
	)
}
