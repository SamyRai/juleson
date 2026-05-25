package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	Name          string
	Path          string
	OutputDir     string
	Version       string
	GOOS          string
	GOARCH        string
	Race          bool
	LDFlags       []string
	BuildFlags    []string
	Tags          []string
	TrimPath      bool
	CGOEnabled    bool
	CGOConfigured bool
}

type BuildResult struct {
	Name       string
	OutputPath string
	OutputSize int64
	Duration   time.Duration
	Success    bool
	Error      error
	Output     string
}

func (r *BuildResult) String() string {
	if r == nil {
		return "no build result"
	}
	if !r.Success {
		return fmt.Sprintf("%s failed after %s: %v", r.Name, r.Duration.Round(time.Millisecond), r.Error)
	}
	return fmt.Sprintf("%s built at %s in %s (%.2f MB)", r.Name, r.OutputPath, r.Duration.Round(time.Millisecond), float64(r.OutputSize)/(1024*1024))
}

type Builder struct {
	config Config
}

func DefaultConfig(name, path string) Config {
	return Config{
		Name:      name,
		Path:      path,
		OutputDir: "bin",
		Version:   "dev",
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
	}
}

func NewBuilder(config Config) *Builder {
	return &Builder{config: config}
}

func (b *Builder) Build(ctx context.Context) error {
	result := b.BuildWithResult(ctx)
	return result.Error
}

func (b *Builder) BuildWithResult(ctx context.Context) *BuildResult {
	start := time.Now()
	outputPath := b.outputPath()
	result := &BuildResult{Name: b.config.Name, OutputPath: outputPath}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		result.Duration = time.Since(start)
		result.Error = err
		return result
	}

	args := append([]string{"build"}, b.config.BuildFlags...)
	if len(b.config.Tags) > 0 {
		args = append(args, "-tags", strings.Join(b.config.Tags, ","))
	}
	if b.config.TrimPath {
		args = append(args, "-trimpath")
	}
	if b.config.Race {
		args = append(args, "-race")
	}
	if len(b.config.LDFlags) > 0 {
		args = append(args, "-ldflags", strings.Join(b.config.LDFlags, " "))
	}
	args = append(args, "-o", outputPath)
	args = append(args, b.config.Path)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = os.Environ()
	if b.config.GOOS != "" {
		cmd.Env = append(cmd.Env, "GOOS="+b.config.GOOS)
	}
	if b.config.GOARCH != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+b.config.GOARCH)
	}
	if b.config.CGOConfigured {
		cgoEnabled := "0"
		if b.config.CGOEnabled {
			cgoEnabled = "1"
		}
		cmd.Env = append(cmd.Env, "CGO_ENABLED="+cgoEnabled)
	}
	out, err := cmd.CombinedOutput()

	result.Duration = time.Since(start)
	result.Output = string(out)
	if err != nil {
		result.Error = fmt.Errorf("%w: %s", err, strings.TrimSpace(result.Output))
		return result
	}

	if info, statErr := os.Stat(outputPath); statErr == nil {
		result.OutputSize = info.Size()
	}
	result.Success = true
	return result
}

func (b *Builder) outputPath() string {
	name := b.config.Name
	if b.config.GOOS == "windows" && !strings.HasSuffix(name, ".exe") {
		name += ".exe"
	}
	return filepath.Join(b.config.OutputDir, name)
}
