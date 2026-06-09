package build

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type LintConfig struct {
	Timeout  string
	Packages []string
	FixMode  bool
	Verbose  bool
	Fast     bool
}

type LintResult struct {
	Error    error
	Output   string
	Duration time.Duration
	Success  bool
}

func (r *LintResult) String() string {
	if r == nil {
		return "no lint result"
	}
	if !r.Success {
		return fmt.Sprintf("lint failed after %s: %v", r.Duration.Round(time.Millisecond), r.Error)
	}
	return fmt.Sprintf("lint passed in %s", r.Duration.Round(time.Millisecond))
}

type Linter struct {
	config LintConfig
}

func DefaultLintConfig() LintConfig {
	return LintConfig{Packages: []string{"./..."}}
}

func NewLinter(config LintConfig) *Linter {
	return &Linter{config: config}
}

func (l *Linter) Lint(ctx context.Context) error {
	result := l.LintWithResult(ctx)
	return result.Error
}

func (l *Linter) LintWithResult(ctx context.Context) *LintResult {
	start := time.Now()
	packages := l.config.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}
	args := append([]string{"vet"}, packages...)
	cmd := exec.CommandContext(ctx, "go", args...)
	out, err := cmd.CombinedOutput()
	result := &LintResult{Duration: time.Since(start), Output: string(out)}
	if err != nil {
		result.Error = fmt.Errorf("%w: %s", err, strings.TrimSpace(result.Output))
		return result
	}
	result.Success = true
	return result
}
