package build

import (
	"context"
	"io"
	"os"
	"os/exec"
)

// RunDiffTool runs the specified diff tool and pipes the diff string to its stdin.
// This executes an external command, so it belongs in the internal/build or similar
// boundary package that handles os/exec.
func RunDiffTool(ctx context.Context, diffTool, diffText string) error {
	cmd := exec.CommandContext(ctx, diffTool)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}

	_, _ = io.WriteString(stdin, diffText)
	_ = stdin.Close()

	return cmd.Wait()
}

// LookPath wraps exec.LookPath.
func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
