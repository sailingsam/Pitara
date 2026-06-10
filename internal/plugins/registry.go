package plugins

import (
	"fmt"
	"sort"
)

type Registry struct {
	plugins map[string]Plugin
	order   []string
}

func NewRegistry(plugins ...Plugin) *Registry {
	r := &Registry{
		plugins: make(map[string]Plugin),
	}
	for _, p := range plugins {
		r.Register(p)
	}
	return r
}

func (r *Registry) Register(p Plugin) {
	name := p.Name()
	if _, exists := r.plugins[name]; exists {
		panic(fmt.Sprintf("plugin %q already registered", name))
	}
	r.plugins[name] = p
	r.order = append(r.order, name)
}

func (r *Registry) All() []Plugin {
	out := make([]Plugin, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.plugins[name])
	}
	return out
}

func (r *Registry) Get(name string) (Plugin, bool) {
	p, ok := r.plugins[name]
	return p, ok
}

// RestoreOrder returns plugins in dependency order for restore.
func (r *Registry) RestoreOrder() ([]Plugin, error) {
	names := make([]string, len(r.order))
	copy(names, r.order)

	visited := make(map[string]bool)
	temp := make(map[string]bool)
	ordered := make([]string, 0, len(names))

	var visit func(string) error
	visit = func(name string) error {
		if temp[name] {
			return fmt.Errorf("circular dependency involving %q", name)
		}
		if visited[name] {
			return nil
		}
		p, ok := r.plugins[name]
		if !ok {
			return fmt.Errorf("unknown dependency %q", name)
		}
		temp[name] = true
		for _, dep := range p.Dependencies() {
			if err := visit(dep); err != nil {
				return err
			}
		}
		temp[name] = false
		visited[name] = true
		ordered = append(ordered, name)
		return nil
	}

	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	out := make([]Plugin, len(ordered))
	for i, name := range ordered {
		out[i] = r.plugins[name]
	}
	return out, nil
}

// Names returns registered plugin names in registration order.
func (r *Registry) Names() []string {
	names := make([]string, len(r.order))
	copy(names, r.order)
	sort.Strings(names)
	return names
}
