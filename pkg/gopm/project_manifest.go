package gopm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gomazing/goscript/pkg/hyper"
)

// ProjectManifest describes the package-level contract for a GoScript project.
type ProjectManifest struct {
	Name            string            `json:"name"`
	Version         string            `json:"version,omitempty"`
	Mode            string            `json:"mode,omitempty"`
	Type            string            `json:"type,omitempty"`
	Description     string            `json:"description,omitempty"`
	PackageManager   string            `json:"packageManager,omitempty"`
	Main            string            `json:"main,omitempty"`
	Registry        string            `json:"registry,omitempty"`
	Private         bool              `json:"private,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies,omitempty"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	Workspaces      []string          `json:"workspaces,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Lockfile        string            `json:"lockfile,omitempty"`
}

// DependencyEntry identifies a single dependency scope entry.
type DependencyEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Scope   string `json:"scope"`
}

// NewProjectManifest creates a sane default manifest for a new project.
func NewProjectManifest(name string) ProjectManifest {
	manifest := ProjectManifest{
		Name:          name,
		Version:       "0.1.0",
		Mode:          "cs",
		Type:          "app",
		PackageManager: "gopm",
		Main:          "./cmd/server",
		Registry:      "https://registry.gopm.dev",
		Dependencies:  map[string]string{},
		DevDependencies: map[string]string{},
		PeerDependencies: map[string]string{},
		Scripts: map[string]string{
			"dev":   "go run ./cmd/server",
			"build": "bo export packs/" + name + ".pack -exe",
			"pack":  "bo export packs/" + name + ".pack -goe",
			"test":  "go test ./...",
		},
		Workspaces: []string{},
		Metadata:   map[string]string{},
		Lockfile:   "goscript.lock.hyper",
	}
	return manifest
}

// LoadProjectManifest loads and normalizes a package manifest from disk.
func LoadProjectManifest(path string) (ProjectManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectManifest{}, err
	}
	var manifest ProjectManifest
	if err := hyper.Unmarshal(data, &manifest); err != nil {
		return ProjectManifest{}, fmt.Errorf("decode project manifest %q: %w", path, err)
	}

	manifest.Normalize(path)
	if err := manifest.Validate(); err != nil {
		return ProjectManifest{}, err
	}

	return manifest, nil
}

// Normalize fills inferable defaults.
func (m *ProjectManifest) Normalize(sourcePath string) {
	if m.Name == "" {
		base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
		if base == "" || base == "." || base == string(filepath.Separator) {
			base = "goscript-app"
		}
		m.Name = base
	}

	if m.Version == "" {
		m.Version = "0.1.0"
	}
	if m.Mode == "" {
		m.Mode = "cs"
	}
	if m.Type == "" {
		m.Type = "app"
	}
	if m.PackageManager == "" {
		m.PackageManager = "gopm"
	}
	if m.Main == "" {
		m.Main = "./cmd/server"
	}
	if m.Registry == "" {
		m.Registry = "https://registry.gopm.dev"
	}
	if m.Lockfile == "" {
		m.Lockfile = "goscript.lock.hyper"
	}
	if m.Dependencies == nil {
		m.Dependencies = map[string]string{}
	}
	if m.DevDependencies == nil {
		m.DevDependencies = map[string]string{}
	}
	if m.PeerDependencies == nil {
		m.PeerDependencies = map[string]string{}
	}
	if m.Scripts == nil {
		m.Scripts = map[string]string{}
	}
	if m.Workspaces == nil {
		m.Workspaces = []string{}
	}
	if m.Metadata == nil {
		m.Metadata = map[string]string{}
	}
}

// Validate checks the minimal package contract.
func (m ProjectManifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("project manifest name is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("project manifest version is required")
	}
	if m.Mode != "" && m.Mode != "cs" && m.Mode != "sw" {
		return fmt.Errorf("project manifest mode must be either \"cs\" or \"sw\"")
	}
	return nil
}

// Write serializes the manifest to disk.
func (m ProjectManifest) Write(path string) error {
	return hyper.WriteFile(path, m)
}

// DependencyEntries returns a stable list of dependencies across all scopes.
func (m ProjectManifest) DependencyEntries() []DependencyEntry {
	entries := make([]DependencyEntry, 0, len(m.Dependencies)+len(m.DevDependencies)+len(m.PeerDependencies))
	seen := map[string]struct{}{}
	for _, name := range sortedMapKeys(m.Dependencies) {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		entries = append(entries, DependencyEntry{Name: name, Version: m.Dependencies[name], Scope: "runtime"})
	}
	for _, name := range sortedMapKeys(m.DevDependencies) {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		entries = append(entries, DependencyEntry{Name: name, Version: m.DevDependencies[name], Scope: "dev"})
	}
	for _, name := range sortedMapKeys(m.PeerDependencies) {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		entries = append(entries, DependencyEntry{Name: name, Version: m.PeerDependencies[name], Scope: "peer"})
	}
	return entries
}

// DependencyNames returns dependency names in deterministic order.
func (m ProjectManifest) DependencyNames() []string {
	names := make([]string, 0, len(m.Dependencies)+len(m.DevDependencies)+len(m.PeerDependencies))
	for _, entry := range m.DependencyEntries() {
		names = append(names, entry.Name)
	}
	return names
}

// LockfilePath resolves the lockfile path relative to a base directory.
func (m ProjectManifest) LockfilePath(baseDir string) string {
	if strings.TrimSpace(m.Lockfile) == "" {
		m.Lockfile = "goscript.lock.hyper"
	}
	if filepath.IsAbs(m.Lockfile) {
		return m.Lockfile
	}
	if baseDir == "" {
		return m.Lockfile
	}
	return filepath.Join(baseDir, m.Lockfile)
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
