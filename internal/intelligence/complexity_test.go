package intelligence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeComplexity(t *testing.T) {
	// Create a temporary directory for test files
	dir, err := os.MkdirTemp("", "complexity_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Write a test Go file with varying complexity
	testCode := `
package testpkg

func Simple() {
	// Complexity 1
}

func Complex() {
	if true {
		for i := 0; i < 10; i++ {
			switch i {
			case 1:
			case 2:
			}
		}
	}
	// Base (1) + if (1) + for (1) + case 1 (1) + case 2 (1) = 5
}

func (t *MyType) Method() {
	if true && false || true {
		// Base (1) + if (1) + && (1) + || (1) = 4
	}
}

type MyType struct{}
`
	err = os.WriteFile(filepath.Join(dir, "main.go"), []byte(testCode), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Create a go.mod so packages.Load works properly
	modCode := "module testpkg\ngo 1.25\n"
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(modCode), 0600)

	results, err := AnalyzeComplexity(context.Background(), dir)
	if err != nil {
		t.Fatalf("AnalyzeComplexity failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(results))
	}

	// Verify sorting (highest first) and values
	if results[0].FuncName != "Complex" || results[0].Complexity != 5 {
		t.Errorf("expected Complex with 5, got %s with %d", results[0].FuncName, results[0].Complexity)
	}
	if results[1].FuncName != "(*MyType).Method" || results[1].Complexity != 4 {
		t.Errorf("expected (*MyType).Method with 4, got %s with %d", results[1].FuncName, results[1].Complexity)
	}
	if results[2].FuncName != "Simple" || results[2].Complexity != 1 {
		t.Errorf("expected Simple with 1, got %s with %d", results[2].FuncName, results[2].Complexity)
	}
}
