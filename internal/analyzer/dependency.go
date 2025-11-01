package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// DependencyAnalyzer analyzes project dependencies
type DependencyAnalyzer struct{}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{}
}

// Analyze analyzes dependencies based on the project type
func (d *DependencyAnalyzer) Analyze(projectPath string) (map[string]string, error) {
	dependencies := make(map[string]string)

	// Try Go modules
	if deps, err := d.analyzeGoMod(projectPath); err == nil && len(deps) > 0 {
		for k, v := range deps {
			dependencies[k] = v
		}
	}

	// Try package.json
	if deps, err := d.analyzePackageJSON(projectPath); err == nil && len(deps) > 0 {
		for k, v := range deps {
			dependencies[k] = v
		}
	}

	// Try requirements.txt
	if deps, err := d.analyzeRequirementsTxt(projectPath); err == nil && len(deps) > 0 {
		for k, v := range deps {
			dependencies[k] = v
		}
	}

	return dependencies, nil
}

func (d *DependencyAnalyzer) analyzeGoMod(projectPath string) (map[string]string, error) {
	deps := make(map[string]string)
	goModPath := filepath.Join(projectPath, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return deps, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inRequire := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "require") {
			inRequire = true
			continue
		}

		if inRequire {
			if line == ")" {
				break
			}

			// Parse dependency line
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				version := parts[1]
				deps[name] = version
			}
		}
	}

	return deps, scanner.Err()
}

func (d *DependencyAnalyzer) analyzePackageJSON(projectPath string) (map[string]string, error) {
	// This is a simplified version - in production you'd parse JSON properly
	deps := make(map[string]string)
	packageJSONPath := filepath.Join(projectPath, "package.json")

	if _, err := os.Stat(packageJSONPath); err != nil {
		return deps, err
	}

	// For now, just mark that package.json exists
	deps["package.json"] = "detected"
	return deps, nil
}

func (d *DependencyAnalyzer) analyzeRequirementsTxt(projectPath string) (map[string]string, error) {
	deps := make(map[string]string)
	reqPath := filepath.Join(projectPath, "requirements.txt")

	file, err := os.Open(reqPath)
	if err != nil {
		return deps, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse dependency line (package==version or package>=version, etc.)
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return r == '=' || r == '>' || r == '<' || r == '~'
		})

		if len(parts) >= 1 {
			name := strings.TrimSpace(parts[0])
			version := "unspecified"
			if len(parts) >= 2 {
				version = strings.TrimSpace(parts[len(parts)-1])
			}
			deps[name] = version
		}
	}

	return deps, scanner.Err()
}
