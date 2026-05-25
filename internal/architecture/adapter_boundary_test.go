package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevAndDockerAdaptersDoNotOwnCommandExecution(t *testing.T) {
	root := repoRoot(t)
	files := []string{
		"internal/cli/commands/dev.go",
		"internal/mcp/tools/dev.go",
		"internal/mcp/tools/docker.go",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, file))
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			text := string(content)
			forbidden := []string{
				"\"os/exec\"",
				"exec.Command",
				"exec.CommandContext",
				"github.com/SamyRai/juleson/internal/build",
			}
			for _, pattern := range forbidden {
				if strings.Contains(text, pattern) {
					t.Fatalf("%s still contains adapter-owned execution dependency %q", file, pattern)
				}
			}
			if !strings.Contains(text, "internal/orchestrator") {
				t.Fatalf("%s must call internal/orchestrator as the owner package", file)
			}
		})
	}
}

func TestCLIMCPAdaptersDoNotImportExec(t *testing.T) {
	root := repoRoot(t)
	adapterDirs := []string{
		"internal/cli/commands",
		"internal/mcp/tools",
	}

	for _, adapterDir := range adapterDirs {
		t.Run(adapterDir, func(t *testing.T) {
			entries, err := os.ReadDir(filepath.Join(root, adapterDir))
			if err != nil {
				t.Fatalf("read adapter dir: %v", err)
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
					continue
				}
				file := filepath.Join(root, adapterDir, entry.Name())
				content, err := os.ReadFile(file)
				if err != nil {
					t.Fatalf("read file: %v", err)
				}
				text := string(content)
				if strings.Contains(text, "\"os/exec\"") || strings.Contains(text, "exec.Command") {
					t.Fatalf("%s still owns command execution", filepath.Join(adapterDir, entry.Name()))
				}
			}
		})
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			t.Fatal("go.mod not found")
		}
		dir = next
	}
}
