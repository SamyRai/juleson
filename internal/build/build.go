package build

import (
	"context"
	"fmt"
	"io"
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

type TestConfig struct {
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

type LintConfig struct {
	Packages []string
	FixMode  bool
	Verbose  bool
	Fast     bool
	Timeout  string
}

type LintResult struct {
	Duration time.Duration
	Success  bool
	Error    error
	Output   string
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

type Cleaner struct {
	binDir string
	files  []string
}

func NewCleaner(binDir string, files []string) *Cleaner {
	return &Cleaner{binDir: binDir, files: files}
}

func (c *Cleaner) Clean(ctx context.Context) error {
	_ = ctx
	if c.binDir != "" {
		if err := os.RemoveAll(c.binDir); err != nil {
			return err
		}
	}
	for _, file := range c.files {
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cleaner) CleanAll(ctx context.Context) error {
	if err := c.Clean(ctx); err != nil {
		return err
	}
	if err := c.CleanCache(ctx); err != nil {
		return err
	}
	return c.CleanTestCache(ctx)
}

func (c *Cleaner) CleanCache(ctx context.Context) error {
	return runGo(ctx, "clean", "-cache")
}

func (c *Cleaner) CleanModCache(ctx context.Context) error {
	return runGo(ctx, "clean", "-modcache")
}

func (c *Cleaner) CleanTestCache(ctx context.Context) error {
	return runGo(ctx, "clean", "-testcache")
}

type ModuleManager struct{}

func NewModuleManager() *ModuleManager {
	return &ModuleManager{}
}

func (m *ModuleManager) Tidy(ctx context.Context) error {
	return runGo(ctx, "mod", "tidy")
}

func (m *ModuleManager) Download(ctx context.Context) error {
	return runGo(ctx, "mod", "download")
}

func (m *ModuleManager) Verify(ctx context.Context) error {
	return runGo(ctx, "mod", "verify")
}

func (m *ModuleManager) Vendor(ctx context.Context) error {
	return runGo(ctx, "mod", "vendor")
}

func (m *ModuleManager) Graph(ctx context.Context) error {
	return runGo(ctx, "mod", "graph")
}

func (m *ModuleManager) Why(ctx context.Context, packages ...string) error {
	args := append([]string{"mod", "why"}, packages...)
	return runGo(ctx, args...)
}

type InstallResult struct {
	InstallDir string
	Installed  []string
	Failed     []string
}

type Installer struct {
	binDir   string
	binaries []string
}

func NewInstaller(binDir string, binaries []string) *Installer {
	return &Installer{binDir: binDir, binaries: binaries}
}

func (i *Installer) GetInstallPath() (string, error) {
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		return gobin, nil
	}
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", err
	}
	gopath := strings.TrimSpace(string(out))
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}
	return filepath.Join(gopath, "bin"), nil
}

func (i *Installer) Install(ctx context.Context) (*InstallResult, error) {
	path, err := i.GetInstallPath()
	if err != nil {
		return nil, err
	}
	return i.InstallTo(ctx, path)
}

func (i *Installer) InstallTo(ctx context.Context, installDir string) (*InstallResult, error) {
	_ = ctx
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, err
	}
	result := &InstallResult{InstallDir: installDir}
	for _, binary := range i.binaries {
		src := filepath.Join(i.binDir, binary)
		if runtime.GOOS == "windows" && !strings.HasSuffix(src, ".exe") {
			src += ".exe"
		}
		dst := filepath.Join(installDir, filepath.Base(src))
		data, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(dst, data, 0755); err != nil {
			return nil, err
		}
		result.Installed = append(result.Installed, dst)
	}
	return result, nil
}

func (i *Installer) Uninstall(ctx context.Context) error {
	installDir, err := i.GetInstallPath()
	if err != nil {
		return err
	}
	return i.UninstallFrom(ctx, installDir)
}

func (i *Installer) UninstallFrom(ctx context.Context, installDir string) error {
	_ = ctx
	for _, binary := range i.binaries {
		binPath := filepath.Join(installDir, binary)
		if runtime.GOOS == "windows" && !strings.HasSuffix(binPath, ".exe") {
			binPath += ".exe"
		}
		if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (i *Installer) IsInPath(dir string) bool {
	for _, pathDir := range filepath.SplitList(os.Getenv("PATH")) {
		if pathDir == dir {
			return true
		}
	}
	return false
}

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
