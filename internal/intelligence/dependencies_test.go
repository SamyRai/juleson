package intelligence

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeDependencies(t *testing.T) {
	// Create a temporary directory for test packages
	dir, err := os.MkdirTemp("", "deps_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	// Create a go.mod so packages.Load works properly
	modCode := "module testdeps\ngo 1.25\n"
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(modCode), 0600)

	// pkg a
	os.MkdirAll(filepath.Join(dir, "a"), 0755)
	os.WriteFile(filepath.Join(dir, "a", "a.go"), []byte("package a\nimport \"testdeps/b\"\nfunc A() { b.B() }"), 0600)

	// pkg b
	os.MkdirAll(filepath.Join(dir, "b"), 0755)
	os.WriteFile(filepath.Join(dir, "b", "b.go"), []byte("package b\nfunc B() {}"), 0600)

	// main pkg
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nimport \"testdeps/a\"\nfunc main() { a.A() }"), 0600)

	graph, err := AnalyzeDependencies(context.Background(), dir)
	if err != nil {
		t.Fatalf("AnalyzeDependencies failed: %v", err)
	}

	// Verify edges
	if len(graph.Edges["root"]) != 1 || graph.Edges["root"][0] != "a" {
		t.Errorf("expected root -> a, got %v", graph.Edges["root"])
	}
	if len(graph.Edges["a"]) != 1 || graph.Edges["a"][0] != "b" {
		t.Errorf("expected a -> b, got %v", graph.Edges["a"])
	}

	// Verify Mermaid render
	mermaid := RenderMermaid(graph)
	if !strings.Contains(mermaid, "\"root\" --> \"a\"") {
		t.Errorf("mermaid missing root->a: %s", mermaid)
	}
	if !strings.Contains(mermaid, "\"a\" --> \"b\"") {
		t.Errorf("mermaid missing a->b: %s", mermaid)
	}
}
