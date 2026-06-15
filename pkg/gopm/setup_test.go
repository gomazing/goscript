package gopm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomazing/goscript/pkg/buildout"
)

func TestParseSetupArgsDefaults(t *testing.T) {
	opts, err := parseSetupArgs([]string{"demo-app"})
	if err != nil {
		t.Fatalf("parseSetupArgs returned error: %v", err)
	}

	if opts.Mode != "cs" {
		t.Fatalf("expected mode cs, got %q", opts.Mode)
	}

	if opts.Type != "app" {
		t.Fatalf("expected type app, got %q", opts.Type)
	}

	if filepath.Base(opts.ProjectDir) != "demo-app" {
		t.Fatalf("expected project dir to end with demo-app, got %q", opts.ProjectDir)
	}
}

func TestParseSetupArgsSwarmERP(t *testing.T) {
	opts, err := parseSetupArgs([]string{"--sw", "--type", "erp", "mesh-suite"})
	if err != nil {
		t.Fatalf("parseSetupArgs returned error: %v", err)
	}

	if opts.Mode != "sw" {
		t.Fatalf("expected mode sw, got %q", opts.Mode)
	}

	if opts.Type != "erp" {
		t.Fatalf("expected type erp, got %q", opts.Type)
	}
}

func TestSetupProjectWritesManifest(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "erp-suite")

	pm := NewPackageManager()
	manifestPath, err := pm.setupProject(SetupOptions{
		ProjectDir:   projectDir,
		ProjectName:  "erp-suite",
		Mode:         "sw",
		Type:         "erp",
		Entrypoint:   "./cmd/server",
		ManifestName: "erp-suite",
	})
	if err != nil {
		t.Fatalf("setupProject returned error: %v", err)
	}

	manifest, err := buildout.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("failed to load pack: %v", err)
	}

	projectManifestPath := filepath.Join(projectDir, "gopm.hyper")
	if _, err := os.Stat(projectManifestPath); err != nil {
		t.Fatalf("expected project manifest %s to exist: %v", projectManifestPath, err)
	}

	projectManifest, err := LoadProjectManifest(projectManifestPath)
	if err != nil {
		t.Fatalf("failed to load project manifest: %v", err)
	}
	if projectManifest.Mode != "sw" {
		t.Fatalf("expected project manifest mode sw, got %q", projectManifest.Mode)
	}
	if projectManifest.PackageManager != "gopm" {
		t.Fatalf("expected package manager gopm, got %q", projectManifest.PackageManager)
	}

	if manifest.Mode != "sw" {
		t.Fatalf("expected mode sw, got %q", manifest.Mode)
	}

	if manifest.BaseDir != "base" {
		t.Fatalf("expected base dir base, got %q", manifest.BaseDir)
	}

	if manifest.AgentsDir != "agents" {
		t.Fatalf("expected agents dir agents, got %q", manifest.AgentsDir)
	}

	requiredDirs := []string{
		filepath.Join(projectDir, "base"),
		filepath.Join(projectDir, "agents"),
		filepath.Join(projectDir, "app", "pages"),
		filepath.Join(projectDir, "cmd", "server"),
		filepath.Join(projectDir, "app", "swarm-policies"),
		filepath.Join(projectDir, "packs"),
	}

	for _, dir := range requiredDirs {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s to exist", dir)
		}
	}

	requiredFiles := []string{
		filepath.Join(projectDir, "go.mod"),
		filepath.Join(projectDir, "README.md"),
		filepath.Join(projectDir, "cmd", "server", "main.go"),
		filepath.Join(projectDir, "app", "pages", "home.gsx"),
		filepath.Join(projectDir, "app", "pages", "home.go"),
	}

	for _, file := range requiredFiles {
		if info, err := os.Stat(file); err != nil || info.IsDir() {
			t.Fatalf("expected file %s to exist", file)
		}
	}

	starterHomeGo, err := os.ReadFile(filepath.Join(projectDir, "app", "pages", "home.go"))
	if err != nil {
		t.Fatalf("failed to read starter home.go: %v", err)
	}
	if !strings.Contains(string(starterHomeGo), "func Home") || !strings.Contains(string(starterHomeGo), "goscript.CreateElement") {
		t.Fatalf("starter home.go does not contain the expected GoScript runtime page")
	}

	starterMain, err := os.ReadFile(filepath.Join(projectDir, "cmd", "server", "main.go"))
	if err != nil {
		t.Fatalf("failed to read starter main.go: %v", err)
	}
	if !strings.Contains(string(starterMain), "RegisterTalkEndpoint") || !strings.Contains(string(starterMain), "/api/hello") {
		t.Fatalf("starter main.go is missing the expected batteries-included routes")
	}

	starterGoMod, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		t.Fatalf("failed to read starter go.mod: %v", err)
	}
	goMod := string(starterGoMod)
	if !strings.Contains(goMod, "module example.com/erp-suite") {
		t.Fatalf("starter go.mod does not contain the expected module path")
	}
	if !strings.Contains(goMod, "require github.com/gomazing/goscript v0.0.0") {
		t.Fatalf("starter go.mod is missing the goscript dependency requirement")
	}
	if !strings.Contains(goMod, "replace github.com/gomazing/goscript =>") {
		t.Fatalf("starter go.mod is missing the local replace entry")
	}

	starterReadme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	if err != nil {
		t.Fatalf("failed to read starter README.md: %v", err)
	}
	readme := string(starterReadme)
	if !strings.Contains(readme, "gopm.hyper") || !strings.Contains(readme, "go run ./cmd/server") {
		t.Fatalf("starter README.md does not describe the portable manifest and run command")
	}
}
