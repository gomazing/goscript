package goscript

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

type surfaceCacheEntry struct {
	hash    string
	lowered LoweredUnit
}

// SurfaceCompiler caches lowering results for a specific lowering strategy.
type SurfaceCompiler struct {
	core  *LanguageCore
	lower Lowerer

	mu    sync.RWMutex
	cache map[string]surfaceCacheEntry
}

// NewSurfaceCompiler creates a cached compiler around a language core.
func NewSurfaceCompiler(core *LanguageCore, lower Lowerer) *SurfaceCompiler {
	return &SurfaceCompiler{
		core:  core,
		lower: lower,
		cache: make(map[string]surfaceCacheEntry),
	}
}

// Compile lowers a source unit and caches the result until the source changes.
func (c *SurfaceCompiler) Compile(path string) (LoweredUnit, error) {
	if c == nil {
		return LoweredUnit{}, fmt.Errorf("surface compiler is nil")
	}
	if c.core == nil {
		return LoweredUnit{}, fmt.Errorf("language core is nil")
	}
	if c.lower == nil {
		return LoweredUnit{}, fmt.Errorf("lowerer is required")
	}

	source, ok := c.core.Source(path)
	if !ok {
		return LoweredUnit{}, fmt.Errorf("source not found: %s", path)
	}

	hash := digestSourceUnit(source)

	c.mu.RLock()
	entry, ok := c.cache[source.Path]
	c.mu.RUnlock()
	if ok && entry.hash == hash {
		return cloneLoweredUnit(entry.lowered), nil
	}

	lowered, err := c.core.LowerSource(path, c.lower)
	if err != nil {
		return LoweredUnit{}, err
	}

	c.mu.Lock()
	c.cache[source.Path] = surfaceCacheEntry{
		hash:    hash,
		lowered: cloneLoweredUnit(lowered),
	}
	c.mu.Unlock()

	return lowered, nil
}

// CompileSurfaceSource lowers a single source string into a complete surface artifact.
func CompileSurfaceSource(path, source string) (LoweredUnit, error) {
	unit := SourceUnit{
		Path:    path,
		Content: source,
		Kind:    sourceKindFromPath(path),
		Module:  sourceModuleFromPath(path),
	}
	unit = unit.Normalize()
	if err := unit.Validate(); err != nil {
		return LoweredUnit{}, err
	}

	parser := NewJSXParser(nil)
	surface, err := parser.ParseSurface(source)
	if err != nil {
		return LoweredUnit{}, err
	}

	return LoweredUnit{
		Source:      unit,
		Surface:     surface,
		Run:         surface.Lower(),
		HTML:        surface.Render(),
		Semantics:   unit.Semantics,
		Diagnostics: AnalyzeSource(unit.Path, unit.Content),
	}, nil
}

// CompileAll lowers every known source in deterministic order.
func (c *SurfaceCompiler) CompileAll() ([]LoweredUnit, error) {
	if c == nil {
		return nil, fmt.Errorf("surface compiler is nil")
	}
	if c.core == nil {
		return nil, fmt.Errorf("language core is nil")
	}

	sources := c.core.Sources()
	out := make([]LoweredUnit, 0, len(sources))
	for _, source := range sources {
		lowered, err := c.Compile(source.Path)
		if err != nil {
			return nil, err
		}
		out = append(out, lowered)
	}

	return out, nil
}

// Invalidate removes one source from the cache.
func (c *SurfaceCompiler) Invalidate(path string) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, normalizeSourcePath(path))
}

// Reset clears the full lowering cache.
func (c *SurfaceCompiler) Reset() {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]surfaceCacheEntry)
}

func digestSourceUnit(source SourceUnit) string {
	h := sha256.New()
	_, _ = h.Write([]byte(source.Path))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(string(source.Kind)))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(source.Module))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(source.Content))
	_, _ = h.Write([]byte{0})
	if source.EntryPoint {
		_, _ = h.Write([]byte("entrypoint"))
	}
	_, _ = h.Write([]byte{0})
	if source.Experimental {
		_, _ = h.Write([]byte("experimental"))
	}
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(string(source.Semantics.Plane)))
	_, _ = h.Write([]byte{0})
	if source.Semantics.Realtime {
		_, _ = h.Write([]byte("realtime"))
	}
	_, _ = h.Write([]byte{0})
	if source.Semantics.Streaming {
		_, _ = h.Write([]byte("streaming"))
	}
	_, _ = h.Write([]byte{0})
	if source.Semantics.Stateful {
		_, _ = h.Write([]byte("stateful"))
	}
	_, _ = h.Write([]byte{0})
	if !source.Semantics.Deterministic {
		_, _ = h.Write([]byte("nondeterministic"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func cloneLoweredUnit(unit LoweredUnit) LoweredUnit {
	unit.Diagnostics = append([]Diagnostic(nil), unit.Diagnostics...)
	unit.Surface = cloneGraftNode(unit.Surface)
	unit.Run = cloneRunNode(unit.Run)
	return unit
}

func cloneGraftNode(node GraftNode) GraftNode {
	return GraftNode{
		Kind:     node.Kind,
		Tag:      node.Tag,
		Props:    cloneProps(node.Props),
		Value:    node.Value,
		Children: cloneGraftNodes(node.Children),
	}
}

func cloneRunNode(node RunNode) RunNode {
	return RunNode{
		Kind:     node.Kind,
		Tag:      node.Tag,
		Props:    cloneProps(node.Props),
		Value:    node.Value,
		Children: cloneRunNodes(node.Children),
	}
}
