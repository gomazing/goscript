package cli

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gomazing/goscript/pkg/goscript"
)

// FormatTargets formats files or directories in place.
func FormatTargets(targets []string) error {
	files, err := collectFiles(targets)
	if err != nil {
		return err
	}

	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		formatted := goscript.FormatFile(file, string(contents))
		if formatted == string(contents) {
			continue
		}

		if err := os.WriteFile(file, []byte(formatted), 0644); err != nil {
			return err
		}
		fmt.Printf("formatted %s\n", file)
	}

	return nil
}

// CheckTargets reports diagnostics for files or directories.
func CheckTargets(targets []string) error {
	if len(targets) == 0 {
		targets = []string{"."}
	}

	anyDiagnostics := false
	for _, target := range targets {
		diagnostics, err := goscript.AnalyzePath(target)
		if err != nil {
			return err
		}

		for _, diagnostic := range diagnostics {
			anyDiagnostics = true
			fmt.Printf("%s:%s: %s [%s]\n", diagnostic.File, diagnostic.Severity, diagnostic.Message, diagnostic.Code)
			if diagnostic.Suggestion != "" {
				fmt.Printf("  -> %s\n", diagnostic.Suggestion)
			}
		}
	}

	if !anyDiagnostics {
		fmt.Println("no diagnostics found")
	}

	return nil
}

// IndexTargets indexes source files and prints discovered symbols.
func IndexTargets(targets []string) error {
	files, err := collectFiles(targets)
	if err != nil {
		return err
	}

	index := goscript.NewDocumentIndex()
	total := 0

	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		symbols := index.IndexSource(file, string(contents))
		total += len(symbols)
		for _, symbol := range symbols {
			fmt.Printf("%s:%d %s %s\n", symbol.File, symbol.Line, symbol.Kind, symbol.Name)
		}
	}

	if total == 0 {
		fmt.Println("no symbols found")
	}

	return nil
}

// WatchTargets watches files or directories and prints change batches.
func WatchTargets(targets []string) error {
	if len(targets) == 0 {
		targets = []string{"."}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher := goscript.NewWatcher(targets...)
	watcher.SetInterval(1 * time.Second)

	fmt.Println("watching for changes, press Ctrl+C to stop")

	for batch := range watcher.Watch(ctx) {
		if len(batch) == 0 {
			continue
		}
		for _, change := range batch {
			fmt.Println(goscript.FormatChange(change))
		}
	}

	if err := ctx.Err(); err != nil && err != context.Canceled {
		return err
	}
	return nil
}

func collectFiles(targets []string) ([]string, error) {
	if len(targets) == 0 {
		targets = []string{"."}
	}

	files := make([]string, 0)
	seen := make(map[string]struct{})

	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			clean := filepath.Clean(target)
			if _, ok := seen[clean]; !ok {
				seen[clean] = struct{}{}
				files = append(files, clean)
			}
			continue
		}

		err = filepath.WalkDir(target, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if d.IsDir() {
				if shouldSkipDir(path, d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}

			if !isTextLikeFile(path) {
				return nil
			}

			clean := filepath.Clean(path)
			if _, ok := seen[clean]; ok {
				return nil
			}
			seen[clean] = struct{}{}
			files = append(files, clean)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(files)
	return files, nil
}

func shouldSkipDir(path, name string) bool {
	if path == "." {
		return false
	}

	switch name {
	case ".git", "vendor", "node_modules", "dist", "build", ".idea", ".vscode":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}

func isTextLikeFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".gsx", ".html", ".htm", ".md", ".css", ".hyper", ".yaml", ".yml", ".txt":
		return true
	default:
		return false
	}
}

