package goscript

import (
	"fmt"
	"strings"
	"testing"
)

func TestDocumentIndexCompletionPrefixOrder(t *testing.T) {
	index := NewDocumentIndex()
	index.IndexSource("main.go", "package main\nfunc Zebra() {}\nfunc Alpha() {}\nfunc Zed() {}\n")

	completions := index.Complete("Z")
	if len(completions) != 2 {
		t.Fatalf("expected 2 completions, got %d", len(completions))
	}

	if completions[0] != "Zebra" || completions[1] != "Zed" {
		t.Fatalf("unexpected completions: %#v", completions)
	}
}

func BenchmarkDocumentIndexComplete(b *testing.B) {
	index := NewDocumentIndex()
	var source strings.Builder
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&source, "func Symbol%04d() {}\n", i)
	}
	index.IndexSource("symbols.go", source.String())

	_ = index.Complete("Symbol9")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.Complete("Symbol9")
	}
}
