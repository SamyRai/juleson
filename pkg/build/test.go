package build

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type TestConfig struct {
	WorkingDir   string
	Packages     []string
	Verbose      bool
	Race         bool
	Cover        bool
	CoverProfile string
	Short        bool
	Timeout      time.Duration
	Parallel     int
	RunPattern   string
	SkipPattern  string
	FailFast     bool
	Shuffle      string
}

type TestResult struct {
	Duration time.Duration
	Success  bool
	Error    error
	Output   string
}

func (r *TestResult) String() string {
	if r == nil {
		return "no test result"
	}
	if !r.Success {
		return fmt.Sprintf("tests failed after %s: %v", r.Duration.Round(time.Millisecond), r.Error)
	}
	return fmt.Sprintf("tests passed in %s", r.Duration.Round(time.Millisecond))
}

type Tester struct {
	config TestConfig
}

func DefaultTestConfig() TestConfig {
	return TestConfig{
		Packages: []string{"./..."},
		Verbose:  true,
		Timeout:  10 * time.Minute,
	}
}

func NewTester(config TestConfig) *Tester {
	return &Tester{config: config}
}

func (t *Tester) Test(ctx context.Context) error {
	result := t.TestWithResult(ctx)
	return result.Error
}

func (t *Tester) TestWithResult(ctx context.Context) *TestResult {
	start := time.Now()
	args := []string{"test"}
	if t.config.Verbose {
		args = append(args, "-v")
	}
	if t.config.Race {
		args = append(args, "-race")
	}
	if t.config.Cover {
		args = append(args, "-cover")
	}
	if t.config.CoverProfile != "" {
		args = append(args, "-coverprofile", t.config.CoverProfile)
	}
	if t.config.Short {
		args = append(args, "-short")
	}
	if t.config.Timeout > 0 {
		args = append(args, "-timeout", t.config.Timeout.String())
	}
	if t.config.Parallel > 0 {
		args = append(args, "-parallel", fmt.Sprintf("%d", t.config.Parallel))
	}
	if t.config.RunPattern != "" {
		args = append(args, "-run", t.config.RunPattern)
	}
	if t.config.SkipPattern != "" {
		args = append(args, "-skip", t.config.SkipPattern)
	}
	if t.config.FailFast {
		args = append(args, "-failfast")
	}
	if t.config.Shuffle != "" {
		args = append(args, "-shuffle", t.config.Shuffle)
	}
	packages := t.config.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}
	args = append(args, packages...)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = t.config.WorkingDir
	out, err := cmd.CombinedOutput()
	result := &TestResult{Duration: time.Since(start), Output: string(out)}
	if err != nil {
		result.Error = fmt.Errorf("%w: %s", err, strings.TrimSpace(result.Output))
		return result
	}
	result.Success = true
	return result
}

func (t *Tester) GenerateCoverageHTML(ctx context.Context, outputPath string) error {
	profile := t.config.CoverProfile
	if profile == "" {
		profile = "coverage.out"
	}
	cmd := exec.CommandContext(ctx, "go", "tool", "cover", "-html="+profile, "-o", outputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
