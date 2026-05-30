package conflict

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildContextPayload gathers the requested context and formats it as a Markdown payload
func BuildContextPayload(ctx context.Context, workingDir, filePath, patchContent string, opts *ResolutionOptions) (string, error) {
	var b strings.Builder

	b.WriteString("# Merge Conflict Resolution Request\n\n")

	if opts.Guidance != "" {
		b.WriteString("## Developer Instructions\n\n")
		b.WriteString(opts.Guidance)
		b.WriteString("\n\n")
	}

	if opts.IncludeLocalFile {
		b.WriteString(fmt.Sprintf("## Current State of `%s`\n\n", filePath))
		fullPath := filepath.Join(workingDir, filePath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				b.WriteString("*(File does not exist locally)*\n\n")
			} else {
				b.WriteString(fmt.Sprintf("*(Error reading file: %v)*\n\n", err))
			}
		} else {
			b.WriteString("```\n")
			b.WriteString(string(content))
			if !strings.HasSuffix(string(content), "\n") {
				b.WriteString("\n")
			}
			b.WriteString("```\n\n")
		}
	}

	if opts.IncludePatchDiff {
		b.WriteString("## Failing Patch Diff\n\n")
		b.WriteString("```diff\n")
		b.WriteString(patchContent)
		if !strings.HasSuffix(patchContent, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}

	if opts.IncludeCompilerOut {
		b.WriteString("## Compiler / Linter Errors\n\n")
		out, err := getCompilerOutput(ctx, workingDir)
		if err != nil {
			b.WriteString(fmt.Sprintf("*(Error running compiler: %v)*\n\n", err))
		}

		if out != "" {
			b.WriteString("```\n")
			b.WriteString(out)
			b.WriteString("\n```\n\n")
		} else {
			b.WriteString("*(No compiler errors found)*\n\n")
		}
	}

	return b.String(), nil
}

func getCompilerOutput(ctx context.Context, dir string) (string, error) {
	// Attempt a go build or similar depending on the project type.
	// As juleson is a Go tool, we default to go build.
	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	// It's fine if the build fails, we just want the output
	if err != nil && out.Len() == 0 {
		return "", fmt.Errorf("failed to run go build: %w", err)
	}

	return out.String(), nil
}
