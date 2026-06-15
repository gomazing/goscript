package goscript

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Diagnostic describes a source or manifest issue.
type Diagnostic struct {
	File       string `json:"file,omitempty"`
	Code       string `json:"code,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
}

// AnalyzeSource produces lightweight diagnostics for a GoScript source file.
func AnalyzeSource(fileName, source string) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)

	if strings.Contains(source, "TODO") {
		diagnostics = append(diagnostics, Diagnostic{
			File:     fileName,
			Code:     "todo-comment",
			Severity: "info",
			Message:  "source contains TODO markers",
		})
	}

	if strings.Contains(source, "not fully implemented") {
		diagnostics = append(diagnostics, Diagnostic{
			File:       fileName,
			Code:       "stubbed-implementation",
			Severity:   "warning",
			Message:    "source contains a stubbed implementation",
			Suggestion: "replace the stub with a complete implementation or mark it as experimental",
		})
	}

	return diagnostics
}

// AnalyzeFile reads a source file and returns its diagnostics.
func AnalyzeFile(fileName string) ([]Diagnostic, error) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return AnalyzeSource(fileName, string(contents)), nil
}

// AnalyzePath walks a file or directory and aggregates diagnostics.
func AnalyzePath(root string) ([]Diagnostic, error) {
	diagnostics := make([]Diagnostic, 0)

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return AnalyzeFile(root)
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == "dist" || name == "build" {
				if path != root {
					return filepath.SkipDir
				}
			}
			return nil
		}

		switch strings.ToLower(filepath.Ext(path)) {
		case ".go", ".gsx", ".html", ".htm", ".md", ".css", ".hyper", ".yaml", ".yml":
		default:
			return nil
		}

		fileDiagnostics, err := AnalyzeFile(path)
		if err != nil {
			return err
		}
		diagnostics = append(diagnostics, fileDiagnostics...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].File == diagnostics[j].File {
			return diagnostics[i].Code < diagnostics[j].Code
		}
		return diagnostics[i].File < diagnostics[j].File
	})

	return diagnostics, nil
}
