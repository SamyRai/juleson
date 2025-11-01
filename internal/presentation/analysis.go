package presentation

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/analyzer"
)

// ProjectAnalysisFormatter formats project analysis results
type ProjectAnalysisFormatter struct{}

// NewProjectAnalysisFormatter creates a new project analysis formatter
func NewProjectAnalysisFormatter() *ProjectAnalysisFormatter {
	return &ProjectAnalysisFormatter{}
}

// Format displays project analysis results in a user-friendly format
func (f *ProjectAnalysisFormatter) Format(context *analyzer.ProjectContext) string {
	var sb strings.Builder

	sb.WriteString("📊 Project Analysis Results\n")
	sb.WriteString("==========================\n")
	sb.WriteString(fmt.Sprintf("Project Name: %s\n", context.ProjectName))
	sb.WriteString(fmt.Sprintf("Project Type: %s\n", context.ProjectType))
	sb.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(context.Languages, ", ")))
	sb.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(context.Frameworks, ", ")))
	sb.WriteString(fmt.Sprintf("Architecture: %s\n", context.Architecture))
	sb.WriteString(fmt.Sprintf("Complexity: %s\n", context.Complexity))
	sb.WriteString(fmt.Sprintf("Git Status: %s\n", context.GitStatus))
	sb.WriteString(fmt.Sprintf("Dependencies: %d\n", len(context.Dependencies)))
	sb.WriteString(fmt.Sprintf("File Types: %d\n", len(context.FileStructure)))

	return sb.String()
}
