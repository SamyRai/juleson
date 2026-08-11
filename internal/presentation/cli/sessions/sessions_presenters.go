package sessions

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/pkg/build"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

// previewActivityArtifactsContent displays artifact content based on type.
func previewActivityArtifactsContent(cfg *config.Config, artifacts []jules.Artifact) error {
	for i, artifact := range artifacts {
		fmt.Printf("\n  📄 Artifact %d:\n", i+1)

		// Handle different artifact types
		if artifact.BashOutput != nil {
			previewBashOutput(artifact.BashOutput)
		} else if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			err := previewGitPatch(cfg, artifact.ChangeSet.GitPatch)
			if err != nil {
				fmt.Printf("    ⚠️  Failed to preview git patch: %v\n", err)
			}
		} else if artifact.Media != nil {
			previewMedia(artifact.Media)
		} else {
			fmt.Printf("    📄 Unknown artifact type\n")
		}
	}
	return nil
}

// previewBashOutput displays bash command output.
func previewBashOutput(output *jules.BashOutput) error {
	fmt.Printf("    🖥️  Bash Output:\n")
	fmt.Printf("    Command: %s\n", output.Command)
	fmt.Printf("    Exit Code: %d\n", output.ExitCode)

	// Truncate output if too long
	content := output.Output
	if len(content) > 1000 {
		content = content[:1000] + "\n... (truncated)"
	}

	fmt.Printf("    Output:\n")
	fmt.Printf("    ```\n")
	for _, line := range strings.Split(content, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Printf("    ```\n")
	return nil
}

// previewGitPatch displays git diff content.
func previewGitPatch(cfg *config.Config, patch *jules.GitPatch) error {
	fmt.Printf("    🔀 Git Patch:\n")

	if patch.SuggestedCommitMessage != "" {
		fmt.Printf("    Commit Message: %s\n", patch.SuggestedCommitMessage)
	}

	if patch.BaseCommitID != "" {
		fmt.Printf("    Base Commit: %s\n", patch.BaseCommitID)
	}

	if patch.UnidiffPatch == "" {
		fmt.Printf("    No diff content.\n")
		return nil
	}

	if !cfg.Diff.ForceNative {
		var diffTool string
		if cfg.Diff.Tool != "" {
			diffTool = cfg.Diff.Tool
		} else {
			// Look for common diff pagers
			if path, err := build.LookPath("difftastic"); err == nil {
				diffTool = path
			} else if path, err := build.LookPath("delta"); err == nil {
				diffTool = path
			}
		}

		if diffTool != "" {
			err := build.RunDiffTool(context.Background(), diffTool, patch.UnidiffPatch)
			if err != nil {
				fmt.Printf("    ⚠️  Diff tool exited with error: %v\n", err)
			}
			return nil
		}
	}

	// Fallback to native text diff
	fmt.Printf("    Diff:\n")

	// Parse the patch using go-gitdiff
	files, _, err := gitdiff.Parse(strings.NewReader(patch.UnidiffPatch))
	if err == nil && len(files) > 0 {
		var b strings.Builder
		for _, file := range files {
			fmt.Fprintf(&b, "diff --git a/%s b/%s\n", file.OldName, file.NewName)
			for _, fragment := range file.TextFragments {
				b.WriteString(fragment.Header())
				for _, line := range fragment.Lines {
					switch line.Op {
					case gitdiff.OpAdd:
						b.WriteString("+")
					case gitdiff.OpDelete:
						b.WriteString("-")
					case gitdiff.OpContext:
						b.WriteString(" ")
					}
					b.WriteString(line.Line)
				}
			}
		}

		err = quick.Highlight(os.Stdout, b.String(), "diff", "terminal256", "monokai")
		if err != nil {
			// Fallback if highlight fails
			fmt.Println(b.String())
		}
		return nil
	}

	// Fallback if parsing fails
	fmt.Printf("    ```diff\n")
	// Split into lines and add proper indentation
	lines := strings.Split(patch.UnidiffPatch, "\n")
	for _, line := range lines {
		if len(line) > 120 { // Truncate very long lines
			line = line[:120] + "..."
		}
		fmt.Printf("    %s\n", line)
	}
	fmt.Printf("    ```\n")

	return nil
}

// previewMedia displays media artifact information.
func previewMedia(media *jules.Media) error {
	fmt.Printf("    🖼️  Media:\n")
	fmt.Printf("    Type: %s\n", media.MimeType)
	fmt.Printf("    Size: %d bytes\n", len(media.Data))

	// Don't display binary data, just metadata
	if strings.Contains(media.MimeType, "image/") {
		fmt.Printf("    📷 Image data (base64 encoded)\n")
	} else {
		fmt.Printf("    📄 Binary data\n")
	}

	return nil
}
