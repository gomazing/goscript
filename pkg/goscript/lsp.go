package goscript

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	funcSymbolPattern  = regexp.MustCompile(`^\s*func\s+([A-Za-z_][A-Za-z0-9_]*)`)
	typeSymbolPattern  = regexp.MustCompile(`^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)`)
	varSymbolPattern   = regexp.MustCompile(`^\s*var\s+([A-Za-z_][A-Za-z0-9_]*)`)
	constSymbolPattern = regexp.MustCompile(`^\s*const\s+([A-Za-z_][A-Za-z0-9_]*)`)
)

// Symbol describes an indexed language symbol.
type Symbol struct {
	Name     string            `json:"name"`
	Kind     string            `json:"kind"`
	File     string            `json:"file,omitempty"`
	Line     int               `json:"line,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// DocumentIndex stores symbols for completion and navigation.
type DocumentIndex struct {
	mu          sync.RWMutex
	symbols     map[string]Symbol
	names       []string
	sortedNames []string
	dirty       bool
}

// NewDocumentIndex creates an empty index.
func NewDocumentIndex() *DocumentIndex {
	return &DocumentIndex{
		symbols: make(map[string]Symbol),
	}
}

// IndexSource scans source and records simple symbols.
func (i *DocumentIndex) IndexSource(fileName, source string) []Symbol {
	lines := strings.Split(source, "\n")
	results := make([]Symbol, 0)

	for idx, line := range lines {
		if match := funcSymbolPattern.FindStringSubmatch(line); len(match) == 2 {
			results = append(results, Symbol{Name: match[1], Kind: "func", File: fileName, Line: idx + 1})
		}
		if match := typeSymbolPattern.FindStringSubmatch(line); len(match) == 2 {
			results = append(results, Symbol{Name: match[1], Kind: "type", File: fileName, Line: idx + 1})
		}
		if match := varSymbolPattern.FindStringSubmatch(line); len(match) == 2 {
			results = append(results, Symbol{Name: match[1], Kind: "var", File: fileName, Line: idx + 1})
		}
		if match := constSymbolPattern.FindStringSubmatch(line); len(match) == 2 {
			results = append(results, Symbol{Name: match[1], Kind: "const", File: fileName, Line: idx + 1})
		}
	}

	i.mu.Lock()
	for _, symbol := range results {
		if _, exists := i.symbols[symbol.Name]; !exists {
			i.names = append(i.names, symbol.Name)
			i.dirty = true
		}
		i.symbols[symbol.Name] = symbol
	}
	i.mu.Unlock()

	return results
}

// Lookup returns a symbol by name.
func (i *DocumentIndex) Lookup(name string) (Symbol, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	symbol, ok := i.symbols[name]
	return symbol, ok
}

// Complete returns symbol names that share a prefix.
func (i *DocumentIndex) Complete(prefix string) []string {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.dirty {
		i.sortedNames = append(i.sortedNames[:0], i.names...)
		sort.Strings(i.sortedNames)
		i.dirty = false
	}

	names := i.sortedNames
	if len(names) == 0 {
		return nil
	}

	lower := sort.Search(len(names), func(idx int) bool {
		return names[idx] >= prefix
	})

	results := make([]string, 0, len(names)-lower)
	for idx := lower; idx < len(names); idx++ {
		name := names[idx]
		if !strings.HasPrefix(name, prefix) {
			break
		}
		results = append(results, name)
	}
	return results
}
