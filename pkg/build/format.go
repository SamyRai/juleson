package build

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Format(ctx context.Context, paths ...string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	args := append([]string{"fmt"}, paths...)
	cmd := exec.CommandContext(ctx, "go", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (f *Formatter) FormatWithGofumpt(ctx context.Context, paths ...string) error {
	if _, err := exec.LookPath("gofumpt"); err == nil {
		if len(paths) == 0 {
			paths = []string{"."}
		}
		args := append([]string{"-w"}, paths...)
		cmd := exec.CommandContext(ctx, "gofumpt", args...)
		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			return fmt.Errorf("%w: %s", runErr, strings.TrimSpace(string(out)))
		}
		return nil
	}
	return f.Format(ctx, paths...)
}
