package buildout

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// PackSlice describes a slice declared inside a pack file.
type PackSlice struct {
	Path    string `json:"path"`
	Default bool   `json:"default,omitempty"`
	Label   string `json:"label,omitempty"`
}

// PackBundle describes a bundle directive declared inside a pack file.
type PackBundle struct {
	Slice  string `json:"slice,omitempty"`
	Target Target `json:"target"`
}

// LoadManifest reads and parses a BO pack file from disk.
func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}

	manifest, err := ParsePack(string(data))
	if err != nil {
		return Manifest{}, fmt.Errorf("parse pack %q: %w", path, err)
	}

	manifest.Normalize(path)
	if err := manifest.Validate(); err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

// ParsePack parses the textual BO pack format into a Manifest.
func ParsePack(source string) (Manifest, error) {
	manifest := Manifest{
		Environment: map[string]string{},
		Metadata:    map[string]string{},
	}

	lines := strings.Split(source, "\n")
	for _, rawLine := range lines {
		line := stripPackComment(strings.TrimSpace(rawLine))
		if line == "" {
			continue
		}

		tokens, err := tokenizePackLine(line)
		if err != nil {
			return Manifest{}, err
		}
		if len(tokens) == 0 {
			continue
		}

		switch strings.ToLower(tokens[0]) {
		case "pack":
			if len(tokens) > 1 {
				manifest.Name = tokens[1]
				if manifest.Output == "" {
					manifest.Output = tokens[1]
				}
			}
		case "mode":
			if len(tokens) > 1 {
				manifest.Mode = strings.ToLower(tokens[1])
			}
		case "output":
			if len(tokens) > 1 {
				manifest.Output = tokens[1]
			}
		case "module":
			if len(tokens) > 1 {
				manifest.Module = tokens[1]
			}
		case "entrypoint":
			if len(tokens) > 1 {
				manifest.Entrypoint = tokens[1]
			}
		case "base":
			if len(tokens) > 1 {
				manifest.BaseDir = tokens[1]
			}
		case "agents":
			if len(tokens) > 1 {
				manifest.AgentsDir = tokens[1]
			}
		case "runtime":
			if len(tokens) > 1 {
				manifest.Runtime = tokens[1]
			}
		case "description":
			if len(tokens) > 1 {
				manifest.Description = strings.Join(tokens[1:], " ")
			}
		case "page":
			if len(tokens) > 1 {
				manifest.Pages = append(manifest.Pages, tokens[1])
			}
		case "path":
			if len(tokens) > 1 {
				manifest.Paths = append(manifest.Paths, tokens[1])
			}
		case "folder":
			if len(tokens) > 1 {
				manifest.Folders = append(manifest.Folders, tokens[1])
			}
		case "include":
			if len(tokens) > 1 {
				manifest.Include = append(manifest.Include, tokens[1])
			}
		case "asset":
			if len(tokens) > 1 {
				manifest.Assets = append(manifest.Assets, tokens[1])
			}
		case "env":
			key, value := parsePackKeyValue(tokens[1:])
			if key != "" {
				manifest.Environment[key] = value
			}
		case "meta":
			key, value := parsePackKeyValue(tokens[1:])
			if key != "" {
				manifest.Metadata[key] = value
			}
		case "slice":
			slice := PackSlice{}
			if len(tokens) > 1 {
				slice.Path = tokens[1]
			}
			if len(tokens) > 3 && tokens[2] == "-" {
				slice.Label = tokens[3]
				slice.Default = strings.EqualFold(slice.Label, "default")
			}
			if slice.Path != "" {
				manifest.Slices = append(manifest.Slices, slice)
				if strings.HasPrefix(slice.Path, "/") {
					manifest.Pages = append(manifest.Pages, slice.Path)
				} else {
					manifest.Paths = append(manifest.Paths, slice.Path)
				}
			}
		case "bundle":
			bundle := PackBundle{}
			switch {
			case len(tokens) == 2 && strings.HasPrefix(tokens[1], "-"):
				bundle.Target = parseBundleTarget(tokens[1][1:])
			case len(tokens) >= 4 && strings.EqualFold(tokens[1], "slice"):
				bundle.Slice = tokens[2]
				bundle.Target = parseBundleTarget(strings.TrimPrefix(tokens[3], "-"))
			case len(tokens) >= 3 && strings.EqualFold(tokens[1], "slice") && strings.HasPrefix(tokens[2], "-"):
				bundle.Target = parseBundleTarget(strings.TrimPrefix(tokens[2], "-"))
			default:
				if len(tokens) > 1 {
					bundle.Target = parseBundleTarget(strings.TrimPrefix(tokens[1], "-"))
				}
			}
			if bundle.Target != "" {
				manifest.Bundles = append(manifest.Bundles, bundle)
				manifest.Bundle = true
			}
		}
	}

	return manifest, nil
}

// Write serializes the manifest back to a pack file.
func (m Manifest) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(m.PackString()), 0o644)
}

// PackString renders the pack as a stable textual manifest.
func (m Manifest) PackString() string {
	var b strings.Builder
	if m.Name != "" {
		fmt.Fprintf(&b, "pack %s\n", m.Name)
	}
	if m.Mode != "" {
		fmt.Fprintf(&b, "mode %s\n", m.Mode)
	}
	if m.Output != "" && m.Output != m.Name {
		fmt.Fprintf(&b, "output %s\n", m.Output)
	}
	if m.Module != "" {
		fmt.Fprintf(&b, "module %s\n", m.Module)
	}
	if m.Entrypoint != "" {
		fmt.Fprintf(&b, "entrypoint %s\n", m.Entrypoint)
	}
	if m.BaseDir != "" {
		fmt.Fprintf(&b, "base %s\n", m.BaseDir)
	}
	if m.AgentsDir != "" {
		fmt.Fprintf(&b, "agents %s\n", m.AgentsDir)
	}
	if m.Runtime != "" {
		fmt.Fprintf(&b, "runtime %s\n", m.Runtime)
	}
	if m.Description != "" {
		fmt.Fprintf(&b, "description %s\n", quotePackValue(m.Description))
	}

	for _, page := range uniquePackValues(m.Pages) {
		fmt.Fprintf(&b, "page %s\n", page)
	}
	for _, path := range uniquePackValues(m.Paths) {
		fmt.Fprintf(&b, "path %s\n", path)
	}
	for _, folder := range uniquePackValues(m.Folders) {
		fmt.Fprintf(&b, "folder %s\n", folder)
	}
	for _, asset := range uniquePackValues(m.Assets) {
		fmt.Fprintf(&b, "asset %s\n", asset)
	}
	for _, include := range uniquePackValues(m.Include) {
		fmt.Fprintf(&b, "include %s\n", include)
	}

	for _, slice := range m.Slices {
		if slice.Path == "" {
			continue
		}
		switch {
		case slice.Default:
			fmt.Fprintf(&b, "slice %s - default\n", slice.Path)
		case slice.Label != "":
			fmt.Fprintf(&b, "slice %s - %s\n", slice.Path, slice.Label)
		default:
			fmt.Fprintf(&b, "slice %s\n", slice.Path)
		}
	}

	for _, bundle := range m.Bundles {
		switch {
		case bundle.Slice != "":
			fmt.Fprintf(&b, "bundle slice %s -%s\n", bundle.Slice, bundle.Target)
		case bundle.Target != "":
			fmt.Fprintf(&b, "bundle -%s\n", bundle.Target)
		}
	}

	return b.String()
}

func stripPackComment(line string) string {
	if line == "" {
		return ""
	}

	inQuotes := false
	escaped := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inQuotes {
			escaped = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if !inQuotes && ch == '#' {
			return strings.TrimSpace(line[:i])
		}
		if !inQuotes && ch == '/' && i+1 < len(line) && line[i+1] == '/' {
			return strings.TrimSpace(line[:i])
		}
	}
	return strings.TrimSpace(line)
}

func tokenizePackLine(line string) ([]string, error) {
	tokens := make([]string, 0, 4)
	var current strings.Builder
	inQuotes := false
	escaped := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' && inQuotes {
			escaped = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if !inQuotes && (ch == ' ' || ch == '\t') {
			flush()
			continue
		}
		current.WriteByte(ch)
	}

	if inQuotes {
		return nil, fmt.Errorf("unterminated quote in pack line %q", line)
	}
	flush()
	return tokens, nil
}

func parsePackKeyValue(tokens []string) (string, string) {
	if len(tokens) == 0 {
		return "", ""
	}

	if len(tokens) == 1 {
		parts := strings.SplitN(tokens[0], "=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
		return strings.TrimSpace(tokens[0]), ""
	}

	key := strings.TrimSpace(tokens[0])
	if len(tokens) == 2 {
		return key, strings.TrimSpace(tokens[1])
	}

	return key, strings.Join(tokens[1:], " ")
}

func parseBundleTarget(raw string) Target {
	target, err := ParseTarget(raw)
	if err != nil {
		return Target(strings.TrimSpace(raw))
	}
	return target
}

func quotePackValue(value string) string {
	if value == "" {
		return `""`
	}
	if strings.ContainsAny(value, " \t\"") {
		return strconv.Quote(value)
	}
	return value
}

func uniquePackValues(values []string) []string {
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
