package buildout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolvedFile represents a file included in a BO export slice.
type ResolvedFile struct {
	Path         string `json:"path"`
	RelativePath  string `json:"relativePath"`
	Kind         string `json:"kind"`
	Size         int64  `json:"size,omitempty"`
}

// ExportSlice captures the inspectable scope of an export.
type ExportSlice struct {
	ManifestPath string         `json:"manifestPath,omitempty"`
	ModuleRoot   string         `json:"moduleRoot,omitempty"`
	Name         string         `json:"name"`
	Mode         string         `json:"mode,omitempty"`
	Routes       []string       `json:"routes,omitempty"`
	Folders      []string       `json:"folders,omitempty"`
	Assets       []string       `json:"assets,omitempty"`
	Include      []string       `json:"include,omitempty"`
	Files        []ResolvedFile `json:"files,omitempty"`
}

// ResolveSlice resolves route hints and filesystem selections into an export slice.
func (m Manifest) ResolveSlice(moduleRoot string) (ExportSlice, error) {
	rootAbs, err := filepath.Abs(moduleRoot)
	if err != nil {
		return ExportSlice{}, err
	}

	slice := ExportSlice{
		Name:    m.Name,
		Mode:    m.Mode,
		ModuleRoot: rootAbs,
		Routes:  dedupeStrings(append(append([]string{}, m.Pages...), m.Paths...)),
		Folders: append([]string{}, m.Folders...),
		Assets:  append([]string{}, m.Assets...),
		Include: append([]string{}, m.Include...),
		Files:   []ResolvedFile{},
	}

	seen := map[string]struct{}{}
	for _, folder := range m.Folders {
		if err := collectExportInput(rootAbs, folder, "folder", &slice.Files, seen); err != nil {
			return ExportSlice{}, err
		}
	}
	for _, asset := range m.Assets {
		if err := collectExportInput(rootAbs, asset, "asset", &slice.Files, seen); err != nil {
			return ExportSlice{}, err
		}
	}
	for _, include := range m.Include {
		if err := collectExportPattern(rootAbs, include, "include", &slice.Files, seen); err != nil {
			return ExportSlice{}, err
		}
	}

	return slice, nil
}

func collectExportInput(rootAbs, input, kind string, files *[]ResolvedFile, seen map[string]struct{}) error {
	path := strings.TrimSpace(input)
	if path == "" {
		return nil
	}

	resolved, err := resolveUnderRoot(rootAbs, path)
	if err != nil {
		return err
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return fmt.Errorf("stat export input %q: %w", input, err)
	}

	if info.IsDir() {
		return filepath.WalkDir(resolved, func(current string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			return appendResolvedFile(rootAbs, current, kind, files, seen)
		})
	}

	return appendResolvedFile(rootAbs, resolved, kind, files, seen)
}

func collectExportPattern(rootAbs, pattern, kind string, files *[]ResolvedFile, seen map[string]struct{}) error {
	raw := strings.TrimSpace(pattern)
	if raw == "" {
		return nil
	}

	globPattern, err := resolveUnderRoot(rootAbs, raw)
	if err != nil {
		return err
	}

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return fmt.Errorf("glob export pattern %q: %w", pattern, err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("export pattern %q did not match any files", pattern)
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return fmt.Errorf("stat export match %q: %w", match, err)
		}

		if info.IsDir() {
			if err := filepath.WalkDir(match, func(current string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() {
					return nil
				}
				return appendResolvedFile(rootAbs, current, kind, files, seen)
			}); err != nil {
				return err
			}
			continue
		}

		if err := appendResolvedFile(rootAbs, match, kind, files, seen); err != nil {
			return err
		}
	}

	return nil
}

func appendResolvedFile(rootAbs, path, kind string, files *[]ResolvedFile, seen map[string]struct{}) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(rootAbs, abs)
	if err != nil {
		return err
	}
	if rel == "." {
		rel = filepath.Base(abs)
	}
	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path %q escapes export root %q", path, rootAbs)
	}

	rel = filepath.ToSlash(rel)
	if _, ok := seen[abs]; ok {
		return nil
	}
	seen[abs] = struct{}{}

	info, err := os.Stat(abs)
	if err != nil {
		return err
	}

	*files = append(*files, ResolvedFile{
		Path:        abs,
		RelativePath: rel,
		Kind:        kind,
		Size:        info.Size(),
	})
	return nil
}

func resolveUnderRoot(rootAbs, input string) (string, error) {
	path := strings.TrimSpace(input)
	if path == "" {
		return "", nil
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(rootAbs, path)
	}
	path = filepath.Clean(path)

	rel, err := filepath.Rel(rootAbs, path)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside export root %q", input, rootAbs)
	}

	return path, nil
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
