package goscript

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// SourceKind identifies the broad role of a GoScript source unit.
type SourceKind string

const (
	SourceKindUnknown  SourceKind = "unknown"
	SourceKindSurface   SourceKind = "surface"
	SourceKindModule    SourceKind = "module"
	SourceKindManifest  SourceKind = "manifest"
	SourceKindRuntime   SourceKind = "runtime"
	SourceKindTemplate  SourceKind = "template"
)

// RuntimePlane describes where a source unit primarily runs.
type RuntimePlane string

const (
	RuntimePlaneFrontend RuntimePlane = "frontend"
	RuntimePlaneBackend  RuntimePlane = "backend"
	RuntimePlaneRealtime  RuntimePlane = "realtime"
	RuntimePlaneHybrid    RuntimePlane = "hybrid"
)

// RuntimeSemantics captures the execution profile that a source unit expects.
type RuntimeSemantics struct {
	Plane        RuntimePlane
	Realtime     bool
	Streaming    bool
	Stateful     bool
	Deterministic bool
}

// SourceUnit describes a single source file or logical language unit.
type SourceUnit struct {
	Path         string
	Kind         SourceKind
	Module       string
	Content      string
	EntryPoint   bool
	Experimental bool
	Semantics    RuntimeSemantics
}

// Normalize fills default values for a source unit.
func (u SourceUnit) Normalize() SourceUnit {
	u.Path = normalizeSourcePath(u.Path)
	if u.Kind == "" {
		u.Kind = sourceKindFromPath(u.Path)
	}
	if u.Module == "" {
		u.Module = sourceModuleFromPath(u.Path)
	}
	if u.Semantics == (RuntimeSemantics{}) {
		u.Semantics = inferRuntimeSemantics(u.Kind, u.EntryPoint, u.Experimental)
	}
	return u
}

// Validate ensures the source unit is structurally usable.
func (u SourceUnit) Validate() error {
	if strings.TrimSpace(u.Path) == "" {
		return fmt.Errorf("source path is required")
	}
	if strings.TrimSpace(u.Module) == "" {
		return fmt.Errorf("source module is required")
	}
	return nil
}

// Lowerer converts a source unit into a structural UI graph.
type Lowerer func(SourceUnit) (GraftNode, []Diagnostic, error)

// LoweredUnit captures the result of lowering a source unit.
type LoweredUnit struct {
	Source      SourceUnit
	Surface     GraftNode
	Run         RunNode
	HTML        string
	Semantics   RuntimeSemantics
	Diagnostics []Diagnostic
}

// LanguageSnapshot summarizes the registered language state.
type LanguageSnapshot struct {
	Modules     []ModuleSpec
	Sources     []SourceUnit
	BuildOrder  []string
	Diagnostics []Diagnostic
}

// LanguageCore is the explicit source -> lowering -> runtime contract for GoScript.
type LanguageCore struct {
	mu      sync.RWMutex
	modules *ModuleGraph
	sources map[string]SourceUnit
}

// NewLanguageCore creates a new language core with an isolated module graph.
func NewLanguageCore() *LanguageCore {
	return &LanguageCore{
		modules: NewModuleGraph(nil),
		sources: make(map[string]SourceUnit),
	}
}

func (c *LanguageCore) ensure() {
	if c.modules == nil {
		c.modules = NewModuleGraph(nil)
	}
	if c.sources == nil {
		c.sources = make(map[string]SourceUnit)
	}
}

// RegisterModule stores a module in the core language graph.
func (c *LanguageCore) RegisterModule(spec ModuleSpec) error {
	if c == nil {
		return fmt.Errorf("language core is nil")
	}
	c.ensure()
	return c.modules.Register(spec)
}

// AddSource stores a normalized source unit.
func (c *LanguageCore) AddSource(unit SourceUnit) error {
	if c == nil {
		return fmt.Errorf("language core is nil")
	}
	c.ensure()

	unit = unit.Normalize()
	if err := unit.Validate(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.sources[unit.Path] = unit
	return nil
}

// Source returns a registered source unit by path.
func (c *LanguageCore) Source(path string) (SourceUnit, bool) {
	if c == nil {
		return SourceUnit{}, false
	}
	c.ensure()

	path = normalizeSourcePath(path)
	c.mu.RLock()
	defer c.mu.RUnlock()

	unit, ok := c.sources[path]
	return unit, ok
}

// Sources returns all registered source units in path order.
func (c *LanguageCore) Sources() []SourceUnit {
	if c == nil {
		return nil
	}
	c.ensure()

	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]SourceUnit, 0, len(c.sources))
	for _, unit := range c.sources {
		out = append(out, unit)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

// Modules returns the registered module specs in deterministic order.
func (c *LanguageCore) Modules() []ModuleSpec {
	if c == nil {
		return nil
	}
	c.ensure()
	return c.modules.List()
}

// BuildOrder returns the module build order from the graph.
func (c *LanguageCore) BuildOrder() ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("language core is nil")
	}
	c.ensure()
	return c.modules.BuildOrder()
}

// Validate checks the module graph and the source bindings.
func (c *LanguageCore) Validate() error {
	if c == nil {
		return fmt.Errorf("language core is nil")
	}
	c.ensure()

	if err := c.modules.Validate(); err != nil {
		return err
	}

	for _, source := range c.Sources() {
		if _, ok := c.modules.registry.Get(source.Module); !ok {
			return fmt.Errorf("source %q references unknown module %q", source.Path, source.Module)
		}
	}

	return nil
}

// Diagnostics returns source and module diagnostics for the current snapshot.
func (c *LanguageCore) Diagnostics() []Diagnostic {
	if c == nil {
		return nil
	}
	c.ensure()

	diagnostics := make([]Diagnostic, 0)
	for _, source := range c.Sources() {
		diagnostics = append(diagnostics, AnalyzeSource(source.Path, source.Content)...)
	}

	if missing := c.modules.MissingDependencies(); len(missing) > 0 {
		for module, deps := range missing {
			diagnostics = append(diagnostics, Diagnostic{
				File:       module,
				Code:       "missing-dependency",
				Severity:   "warning",
				Message:    fmt.Sprintf("module %q depends on missing modules: %s", module, strings.Join(deps, ", ")),
				Suggestion: "register the dependency module or remove the dependency from the pack",
			})
		}
	}

	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].File == diagnostics[j].File {
			return diagnostics[i].Code < diagnostics[j].Code
		}
		return diagnostics[i].File < diagnostics[j].File
	})

	return diagnostics
}

// Snapshot returns a stable view of the core state.
func (c *LanguageCore) Snapshot() (LanguageSnapshot, error) {
	if c == nil {
		return LanguageSnapshot{}, fmt.Errorf("language core is nil")
	}
	c.ensure()

	buildOrder, err := c.BuildOrder()
	if err != nil {
		return LanguageSnapshot{}, err
	}

	return LanguageSnapshot{
		Modules:     c.Modules(),
		Sources:     c.Sources(),
		BuildOrder:  buildOrder,
		Diagnostics: c.Diagnostics(),
	}, nil
}

// LowerSource lowers a stored source unit through the supplied lowerer.
func (c *LanguageCore) LowerSource(path string, lower Lowerer) (LoweredUnit, error) {
	if c == nil {
		return LoweredUnit{}, fmt.Errorf("language core is nil")
	}
	if lower == nil {
		return LoweredUnit{}, fmt.Errorf("lowerer is required")
	}
	c.ensure()

	source, ok := c.Source(path)
	if !ok {
		return LoweredUnit{}, fmt.Errorf("source not found: %s", path)
	}

	surface, extraDiagnostics, err := lower(source)
	if err != nil {
		return LoweredUnit{}, err
	}

	lowered := LoweredUnit{
		Source:      source,
		Surface:     surface,
		Run:         surface.Lower(),
		HTML:        surface.Render(),
		Semantics:   source.Semantics,
		Diagnostics: AnalyzeSource(source.Path, source.Content),
	}
	lowered.Diagnostics = append(lowered.Diagnostics, extraDiagnostics...)
	return lowered, nil
}

func normalizeSourcePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	path = filepath.Clean(path)
	if path == "." {
		return ""
	}

	return filepath.ToSlash(path)
}

func sourceKindFromPath(path string) SourceKind {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".gsx", ".wrt", ".html", ".htm", ".svg":
		return SourceKindSurface
	case ".go":
		return SourceKindModule
	case ".pack", ".hyper":
		return SourceKindManifest
	case ".tmpl", ".template":
		return SourceKindTemplate
	default:
		return SourceKindUnknown
	}
}

func sourceModuleFromPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "module"
	}

	cleaned := filepath.Clean(path)
	dir := filepath.Dir(cleaned)
	if dir != "." && dir != string(filepath.Separator) {
		module := sanitizeModuleName(filepath.Base(dir))
		if module != "" {
			return module
		}
	}

	module := sanitizeModuleName(strings.TrimSuffix(filepath.Base(cleaned), filepath.Ext(cleaned)))
	if module == "" {
		return "module"
	}
	return module
}

func sanitizeModuleName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"\\", "-",
		"/", "-",
		" ", "-",
		".", "-",
		"_", "-",
		":", "-",
	)
	raw = replacer.Replace(raw)
	raw = strings.Trim(raw, "-")
	return raw
}

func inferRuntimeSemantics(kind SourceKind, entryPoint, experimental bool) RuntimeSemantics {
	semantics := RuntimeSemantics{
		Plane:         RuntimePlaneHybrid,
		Realtime:      false,
		Streaming:     false,
		Stateful:      false,
		Deterministic: true,
	}

	switch kind {
	case SourceKindSurface:
		semantics.Plane = RuntimePlaneFrontend
		semantics.Streaming = true
	case SourceKindModule:
		semantics.Plane = RuntimePlaneBackend
		semantics.Realtime = true
		semantics.Streaming = true
		semantics.Stateful = true
	case SourceKindManifest:
		semantics.Plane = RuntimePlaneHybrid
		semantics.Stateful = true
	case SourceKindRuntime:
		semantics.Plane = RuntimePlaneHybrid
		semantics.Realtime = true
		semantics.Streaming = true
		semantics.Stateful = true
	}

	if entryPoint {
		semantics.Stateful = true
	}
	if experimental {
		semantics.Deterministic = false
	}

	return semantics
}
