package gopm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectManifestRoundTrip(t *testing.T) {
	manifest := NewProjectManifest("demo-app")
	manifest.Mode = "sw"
	manifest.Type = "erp"
	manifest.Dependencies["goscript-ui"] = "^1.0.0"
	manifest.DevDependencies["goscript-test"] = "^0.2.0"

	root := t.TempDir()
	path := filepath.Join(root, "gopm.hyper")
	if err := manifest.Write(path); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	loaded, err := LoadProjectManifest(path)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	if loaded.Name != "demo-app" {
		t.Fatalf("expected name demo-app, got %q", loaded.Name)
	}
	if loaded.Mode != "sw" {
		t.Fatalf("expected mode sw, got %q", loaded.Mode)
	}
	if len(loaded.DependencyEntries()) != 2 {
		t.Fatalf("expected 2 dependency entries, got %d", len(loaded.DependencyEntries()))
	}
}

func TestLockfileFromManifest(t *testing.T) {
	manifest := NewProjectManifest("demo-app")
	manifest.Dependencies["goscript-ui"] = "^1.0.0"

	lockfile := NewLockfileFromManifest(manifest)
	if len(lockfile.Packages) != 1 {
		t.Fatalf("expected 1 locked package, got %d", len(lockfile.Packages))
	}

	lockPath := filepath.Join(t.TempDir(), manifest.Lockfile)
	if err := lockfile.Write(lockPath); err != nil {
		t.Fatalf("write lockfile: %v", err)
	}

	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read lockfile: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected lockfile to contain data")
	}
}
