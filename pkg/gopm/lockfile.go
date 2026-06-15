package gopm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gomazing/goscript/pkg/hyper"
)

// LockedPackage represents a dependency entry in the project lockfile.
type LockedPackage struct {
	Name         string            `json:"name"`
	Version      string            `json:"version,omitempty"`
	Requested    string            `json:"requested,omitempty"`
	Scope        string            `json:"scope,omitempty"`
	Source       string            `json:"source,omitempty"`
	Integrity    string            `json:"integrity,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Lockfile records the resolved package state for a project.
type Lockfile struct {
	Name          string                    `json:"name"`
	Version       string                    `json:"version,omitempty"`
	Mode          string                    `json:"mode,omitempty"`
	PackageManager string                   `json:"packageManager,omitempty"`
	Registry      string                    `json:"registry,omitempty"`
	GeneratedAt   string                    `json:"generatedAt,omitempty"`
	Packages      map[string]LockedPackage   `json:"packages"`
}

// NewLockfileFromManifest creates a lockfile skeleton from a manifest.
func NewLockfileFromManifest(manifest ProjectManifest) Lockfile {
	lockfile := Lockfile{
		Name:          manifest.Name,
		Version:       manifest.Version,
		Mode:          manifest.Mode,
		PackageManager: manifest.PackageManager,
		Registry:      manifest.Registry,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Packages:      map[string]LockedPackage{},
	}

	for _, entry := range manifest.DependencyEntries() {
		lockfile.Packages[entry.Name] = LockedPackage{
			Name:         entry.Name,
			Version:      entry.Version,
			Requested:    entry.Version,
			Scope:        entry.Scope,
			Source:       manifest.Registry,
			Dependencies: map[string]string{},
			Metadata:     map[string]string{},
		}
	}

	return lockfile
}

// LoadLockfile reads a lockfile from disk.
func LoadLockfile(path string) (Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Lockfile{}, err
	}
	var lockfile Lockfile
	if err := hyper.Unmarshal(data, &lockfile); err != nil {
		return Lockfile{}, fmt.Errorf("decode lockfile %q: %w", path, err)
	}

	if lockfile.Packages == nil {
		lockfile.Packages = map[string]LockedPackage{}
	}
	return lockfile, nil
}

// Write serializes the lockfile to disk.
func (lf Lockfile) Write(path string) error {
	return hyper.WriteFile(path, lf)
}

// SortedPackageNames returns package names in stable order.
func (lf Lockfile) SortedPackageNames() []string {
	names := make([]string, 0, len(lf.Packages))
	for name := range lf.Packages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SetResolvedPackage updates or inserts a resolved package entry.
func (lf *Lockfile) SetResolvedPackage(pkg LockedPackage) {
	if lf.Packages == nil {
		lf.Packages = map[string]LockedPackage{}
	}
	lf.Packages[pkg.Name] = pkg
}
