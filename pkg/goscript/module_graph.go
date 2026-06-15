package goscript

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ModuleGraph tracks module dependencies and resolves build order.
type ModuleGraph struct {
	mu       sync.RWMutex
	registry *ModuleRegistry
	edges    map[string]map[string]struct{}
}

// NewModuleGraph creates a graph backed by the provided registry.
func NewModuleGraph(registry *ModuleRegistry) *ModuleGraph {
	if registry == nil {
		registry = NewModuleRegistry()
	}

	return &ModuleGraph{
		registry: registry,
		edges:    make(map[string]map[string]struct{}),
	}
}

// Register stores a module spec and records its declared dependencies.
func (g *ModuleGraph) Register(spec ModuleSpec) error {
	spec.Normalize()
	if err := spec.Validate(); err != nil {
		return err
	}

	if err := g.registry.Register(spec); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.edges[spec.Name] == nil {
		g.edges[spec.Name] = make(map[string]struct{})
	}

	for dependency := range spec.Dependencies {
		dependency = strings.TrimSpace(dependency)
		if dependency == "" {
			continue
		}
		g.edges[spec.Name][dependency] = struct{}{}
	}

	return nil
}

// AddDependency records that module depends on dependency.
func (g *ModuleGraph) AddDependency(module, dependency string) error {
	module = strings.TrimSpace(module)
	dependency = strings.TrimSpace(dependency)
	if module == "" || dependency == "" {
		return fmt.Errorf("module and dependency names are required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	spec, ok := g.registry.Get(module)
	if !ok {
		return fmt.Errorf("module %q is not registered", module)
	}

	spec.Normalize()
	spec.Dependencies[dependency] = ""
	if err := g.registry.Register(spec); err != nil {
		return err
	}

	if g.edges[module] == nil {
		g.edges[module] = make(map[string]struct{})
	}
	g.edges[module][dependency] = struct{}{}

	return nil
}

// MissingDependencies returns registered modules whose dependencies are absent.
func (g *ModuleGraph) MissingDependencies() map[string][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	missing := make(map[string][]string)
	for _, spec := range g.registry.List() {
		for dependency := range spec.Dependencies {
			dependency = strings.TrimSpace(dependency)
			if dependency == "" {
				continue
			}

			if _, ok := g.registry.Get(dependency); !ok {
				missing[spec.Name] = append(missing[spec.Name], dependency)
			}
		}

		if deps := missing[spec.Name]; len(deps) > 1 {
			sort.Strings(deps)
			missing[spec.Name] = deps
		}
	}

	return missing
}

// Dependencies returns direct registered dependencies for a module.
func (g *ModuleGraph) Dependencies(module string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := g.edges[module]
	dependencies := make([]string, 0, len(edges))
	for dependency := range edges {
		if _, ok := g.registry.Get(dependency); ok {
			dependencies = append(dependencies, dependency)
		}
	}
	sort.Strings(dependencies)
	return dependencies
}

// BuildOrder returns a deterministic topological order for registered modules.
func (g *ModuleGraph) BuildOrder() ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	names := g.registry.Names()
	graph := make(map[string][]string, len(names))
	for _, name := range names {
		deps := g.edges[name]
		for dependency := range deps {
			if _, ok := g.registry.Get(dependency); ok {
				graph[name] = append(graph[name], dependency)
			}
		}
		sort.Strings(graph[name])
	}

	order := make([]string, 0, len(names))
	visiting := make(map[string]bool)
	visited := make(map[string]bool)
	var visit func(string) error

	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return fmt.Errorf("cycle detected involving module %q", name)
		}

		visiting[name] = true
		for _, dependency := range graph[name] {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// Validate reports missing dependencies or dependency cycles.
func (g *ModuleGraph) Validate() error {
	if missing := g.MissingDependencies(); len(missing) > 0 {
		var parts []string
		for module, deps := range missing {
			parts = append(parts, fmt.Sprintf("%s -> %s", module, strings.Join(deps, ", ")))
		}
		sort.Strings(parts)
		return fmt.Errorf("missing dependencies: %s", strings.Join(parts, "; "))
	}

	_, err := g.BuildOrder()
	return err
}
