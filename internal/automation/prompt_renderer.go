package automation

import (
	"fmt"
	"strings"
	"time"
)

// processPrompt processes a Jules prompt with context variables.
func (e *Engine) processPrompt(prompt string, contextVars map[string]string) (string, error) {
	processed := prompt

	for key := range contextVars {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		actualValue := e.getContextValue(key)
		processed = strings.ReplaceAll(processed, placeholder, actualValue)
	}

	builtins := map[string]string{
		"ProjectPath": e.projectPath,
		"ProjectName": e.context.ProjectName,
		"ProjectType": e.context.ProjectType,
		"Timestamp":   time.Now().Format(time.RFC3339),
	}

	for key, value := range builtins {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		processed = strings.ReplaceAll(processed, placeholder, value)
	}

	return processed, nil
}

// getContextValue gets a value from the project context.
func (e *Engine) getContextValue(key string) string {
	if e.context == nil {
		return ""
	}

	switch key {
	case "ProjectPath":
		return e.context.ProjectPath
	case "ProjectName":
		return e.context.ProjectName
	case "ProjectType":
		return e.context.ProjectType
	case "Languages":
		return strings.Join(e.context.Languages, ", ")
	case "Frameworks":
		return strings.Join(e.context.Frameworks, ", ")
	case "Architecture":
		return e.context.Architecture
	case "Complexity":
		return e.context.Complexity
	case "GitStatus":
		return e.context.GitStatus
	default:
		if value, exists := e.context.CustomParams[key]; exists {
			return value
		}
		return ""
	}
}
