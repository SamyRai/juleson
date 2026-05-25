package automation

import (
	"fmt"
	"os"

	"github.com/SamyRai/juleson/internal/templates"
)

// generateOutputFiles generates output files based on template configuration.
func (e *Engine) generateOutputFiles(template *templates.Template, result *ExecutionResult) error {
	for _, outputFile := range template.Output.Files {
		filePath, err := e.processPrompt(outputFile.Path, make(map[string]string))
		if err != nil {
			return fmt.Errorf("failed to process output file path: %w", err)
		}

		content, err := e.generateFileContent(outputFile.Template, result)
		if err != nil {
			return fmt.Errorf("failed to generate file content: %w", err)
		}

		if err := e.writeFile(filePath, content); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		result.OutputFiles = append(result.OutputFiles, filePath)
	}

	return nil
}

// generateFileContent generates markdown report content.
func (e *Engine) generateFileContent(templateName string, result *ExecutionResult) (string, error) {
	content := fmt.Sprintf(`# %s Execution Report

## Summary
- Template: %s
- Project: %s
- Duration: %v
- Success: %t

## Tasks Executed
`, templateName, result.TemplateName, result.ProjectPath, result.Duration, result.Success)

	for _, task := range result.TasksExecuted {
		content += fmt.Sprintf("- %s (%s): %t\n", task.TaskName, task.TaskType, task.Success)
	}

	return content, nil
}

// writeFile writes content to a file.
func (e *Engine) writeFile(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), DefaultFilePermissions)
}
