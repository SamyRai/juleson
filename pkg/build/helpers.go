package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func runGo(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "go", args...)
	var output strings.Builder
	cmd.Stdout = io.MultiWriter(os.Stdout, &output)
	cmd.Stderr = io.MultiWriter(os.Stderr, &output)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(output.String()))
	}
	return nil
}
