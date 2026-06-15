package goscript

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWatcherDiff(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")

	if err := os.WriteFile(file, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	watcher := NewWatcher(dir)
	first, err := watcher.Scan()
	if err != nil {
		t.Fatalf("first scan: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected one file in first snapshot, got %d", len(first))
	}

	if err := os.WriteFile(file, []byte("package main\nvar Version = 1\n"), 0644); err != nil {
		t.Fatalf("rewrite file: %v", err)
	}

	second, err := watcher.Scan()
	if err != nil {
		t.Fatalf("second scan: %v", err)
	}

	changes := watcher.Diff(second)
	if len(changes) == 0 {
		t.Fatalf("expected a change batch")
	}
}
