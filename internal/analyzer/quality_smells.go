package analyzer

import (
	"fmt"
	"os/exec"
)

// analyzeCodeSmells analyzes code quality issues.
func (c *smellAnalyzer) analyzeCodeSmells(projectPath string, languages []string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)

	for _, lang := range languages {
		switch lang {
		case "go":
			if goSmells, err := c.analyzeGoCodeSmells(projectPath); err == nil {
				smells = append(smells, goSmells...)
			}
		case "python":
			if pySmells, err := c.analyzePythonCodeSmells(projectPath); err == nil {
				smells = append(smells, pySmells...)
			}
		case "javascript", "typescript":
			if jsSmells, err := c.analyzeJavaScriptCodeSmells(projectPath); err == nil {
				smells = append(smells, jsSmells...)
			}
		}
	}

	return smells, nil
}

func (c *smellAnalyzer) analyzeGoCodeSmells(projectPath string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)
	if _, err := exec.LookPath("revive"); err == nil {
		cmd := exec.Command("revive", "-formatter", "json", "./...")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if reviveSmells, err := c.parseReviveOutput(string(output)); err == nil {
				smells = append(smells, reviveSmells...)
			}
		}
	}
	return smells, nil
}

func (c *smellAnalyzer) analyzePythonCodeSmells(projectPath string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)
	if _, err := exec.LookPath("pylint"); err == nil {
		cmd := exec.Command("pylint", "--output-format=json", ".")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if pylintSmells, err := c.parsePylintOutput(string(output)); err == nil {
				smells = append(smells, pylintSmells...)
			}
		}
	}
	return smells, nil
}

func (c *smellAnalyzer) analyzeJavaScriptCodeSmells(projectPath string) ([]CodeSmell, error) {
	smells := make([]CodeSmell, 0)
	if hasPackageJSON(projectPath) {
		cmd := exec.Command("npx", "eslint", ".", "--format", "json")
		cmd.Dir = projectPath
		output, err := cmd.Output()
		if err == nil {
			if eslintSmells, err := c.parseESLintOutput(string(output)); err == nil {
				smells = append(smells, eslintSmells...)
			}
		}
	}
	return smells, nil
}

func (c *smellAnalyzer) parseReviveOutput(output string) ([]CodeSmell, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *smellAnalyzer) parsePylintOutput(output string) ([]CodeSmell, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *smellAnalyzer) parseESLintOutput(output string) ([]CodeSmell, error) {
	return nil, fmt.Errorf("not implemented")
}
