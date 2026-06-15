package buildout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSliceCollectsRoutesAndFiles(t *testing.T) {
	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "app", "components"), 0o755); err != nil {
		t.Fatalf("mkdir app/components: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "static"), 0o755); err != nil {
		t.Fatalf("mkdir static: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app", "components", "card.go"), []byte("package components"), 0o644); err != nil {
		t.Fatalf("write component: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "static", "logo.svg"), []byte("<svg />"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	manifest := Manifest{
		Name:    "admin",
		Mode:    "sw",
		Pages:   []string{"/admin"},
		Paths:   []string{"/admin/users"},
		Folders: []string{"app/components"},
		Assets:  []string{"static"},
		Include: []string{"app/components/*.go"},
	}

	slice, err := manifest.ResolveSlice(root)
	if err != nil {
		t.Fatalf("ResolveSlice returned error: %v", err)
	}

	if len(slice.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(slice.Routes))
	}

	if len(slice.Files) != 2 {
		t.Fatalf("expected 2 resolved files, got %d", len(slice.Files))
	}

	if slice.Files[0].RelativePath == "" {
		t.Fatalf("expected resolved relative path to be populated")
	}
}
