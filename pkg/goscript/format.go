package goscript

import (
	"path/filepath"
	"strings"
)

// FormatScript normalizes line endings and trims trailing spaces.
func FormatScript(input string) string {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// FormatMarkup applies a light normalization pass for markup-like output.
func FormatMarkup(input string) string {
	input = FormatScript(input)
	input = strings.TrimSpace(input)
	return input
}

// FormatFile applies a best-effort formatter based on file extension.
func FormatFile(path, input string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".htm", ".gsx", ".xml", ".svg":
		return FormatMarkup(input)
	default:
		return FormatScript(input)
	}
}
