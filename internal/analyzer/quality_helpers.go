package analyzer

import (
	"os"
	"path/filepath"
)

func hasPackageJSON(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, "package.json"))
	return err == nil
}
