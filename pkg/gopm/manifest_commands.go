package gopm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest prints or scaffolds a project manifest.
func (pm *PackageManager) Manifest(args []string) {
	path := "gopm.hyper"
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		path = strings.TrimSpace(args[0])
	}

	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			path = filepath.Join(path, "gopm.hyper")
			info, err = os.Stat(path)
		}
		if err == nil && !info.IsDir() {
			manifest, err := LoadProjectManifest(path)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			printProjectManifestSummary(path, manifest)
			return
		}
	}

	projectName := projectNameFromPath(path)
	manifest := NewProjectManifest(projectName)
	if err := manifest.Write(path); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Created project manifest: %s\n", path)
	printProjectManifestSummary(path, manifest)
}

// Lock resolves the current project manifest into a lockfile.
func (pm *PackageManager) Lock(args []string) {
	path := "gopm.hyper"
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		path = strings.TrimSpace(args[0])
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, "gopm.hyper")
	}

	manifest, err := LoadProjectManifest(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	lockfile := NewLockfileFromManifest(manifest)
	lockPath := manifest.LockfilePath(filepath.Dir(path))
	if err := lockfile.Write(lockPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Lockfile written: %s\n", lockPath)
	fmt.Printf("Packages: %d\n", len(lockfile.Packages))
}

func printProjectManifestSummary(path string, manifest ProjectManifest) {
	fmt.Printf("Project manifest: %s\n", path)
	fmt.Printf("Name: %s\n", manifest.Name)
	fmt.Printf("Version: %s\n", manifest.Version)
	fmt.Printf("Mode: %s\n", manifest.Mode)
	fmt.Printf("Type: %s\n", manifest.Type)
	fmt.Printf("Main: %s\n", manifest.Main)
	fmt.Printf("Dependencies: %d\n", len(manifest.Dependencies))
	fmt.Printf("Dev dependencies: %d\n", len(manifest.DevDependencies))
	fmt.Printf("Peer dependencies: %d\n", len(manifest.PeerDependencies))
	fmt.Printf("Lockfile: %s\n", manifest.LockfilePath(filepath.Dir(path)))
}

func projectNameFromPath(path string) string {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		if cwd, err := os.Getwd(); err == nil {
			dir = cwd
		}
	}

	name := sanitizeName(filepath.Base(dir))
	if name == "" {
		return "goscript-app"
	}

	return name
}
