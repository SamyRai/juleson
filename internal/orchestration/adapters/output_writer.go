package adapters

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

const defaultOutputFilePermissions = 0644

type MarkdownOutputWriter struct{}

func NewMarkdownOutputWriter() *MarkdownOutputWriter {
	return &MarkdownOutputWriter{}
}

func (w *MarkdownOutputWriter) WriteOutputs(ctx context.Context, template domain.Template, result domain.Result) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	outputs := make([]string, 0, len(template.OutputFiles))
	for _, output := range template.OutputFiles {
		path := renderOutputPath(output.Path, result)
		content := reportContent(firstNonEmpty(output.Template, path), template, result)
		if err := os.WriteFile(path, []byte(content), defaultOutputFilePermissions); err != nil {
			return outputs, fmt.Errorf("write output file %q: %w", path, err)
		}
		outputs = append(outputs, path)
	}
	return outputs, nil
}

func renderOutputPath(path string, result domain.Result) string {
	values := map[string]string{
		"template": result.Goal.ID,
		"project":  result.Goal.Context.ProjectPath,
	}
	rendered := path
	for key, value := range values {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
		rendered = strings.ReplaceAll(rendered, "{{."+key+"}}", value)
	}
	return rendered
}

func reportContent(name string, template domain.Template, result domain.Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s Execution Report\n\n", name))
	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("- Template: %s\n", template.Name))
	sb.WriteString(fmt.Sprintf("- Project: %s\n", result.Goal.Context.ProjectPath))
	sb.WriteString(fmt.Sprintf("- Duration: %v\n", result.Duration))
	sb.WriteString(fmt.Sprintf("- Success: %t\n\n", result.Success))
	sb.WriteString("## Tasks Executed\n")
	for _, task := range result.Tasks {
		sb.WriteString(fmt.Sprintf("- %s (%s): %t\n", task.TaskName, task.TaskType, task.Success))
	}
	return sb.String()
}
