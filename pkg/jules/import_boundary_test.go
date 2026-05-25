package jules

import (
	"os/exec"
	"strings"
	"testing"
)

func TestPackageHasNoAppImports(t *testing.T) {
	output, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", ".").Output()
	if err != nil {
		t.Fatalf("go list imports: %v", err)
	}

	forbidden := []string{
		"/internal/",
		"github.com/spf13/cobra",
		"github.com/spf13/viper",
		"github.com/google/go-github",
		"github.com/modelcontextprotocol/",
		"google.golang.org/genai",
		"os/exec",
		"path/filepath",
	}

	imports := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, importPath := range imports {
		for _, forbiddenImport := range forbidden {
			if strings.Contains(importPath, forbiddenImport) {
				t.Fatalf("pkg/jules imports forbidden dependency %q via %q", forbiddenImport, importPath)
			}
		}
	}
}
