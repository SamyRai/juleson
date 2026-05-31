package builder

import (
	"context"
	"fmt"
	"runtime"

	"github.com/SamyRai/juleson/pkg/build"
)

func (s *Service) BuildWithResults(ctx context.Context, options BuildOptions) (*BuildSummary, error) {
	if options.Target == "" {
		options.Target = "all"
	}
	if options.Version == "" {
		options.Version = "dev"
	}
	if options.GOOS == "" {
		options.GOOS = runtime.GOOS
	}
	if options.GOARCH == "" {
		options.GOARCH = runtime.GOARCH
	}

	binaries := s.selectedBinaries(options.Target)
	summary := &BuildSummary{
		Target:  options.Target,
		Results: make([]*build.BuildResult, 0, len(binaries)),
	}

	for _, binary := range binaries {
		config := build.DefaultConfig(binary.name, binary.path)
		config.Version = options.Version
		config.GOOS = options.GOOS
		config.GOARCH = options.GOARCH
		config.Race = options.Race
		if options.Version != "" && options.Version != "dev" {
			config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", options.Version))
		}

		result := build.NewBuilder(config).BuildWithResult(ctx)
		summary.Results = append(summary.Results, result)
		if result.Success {
			summary.SuccessCount++
			summary.TotalDuration += result.Duration
			summary.TotalSize += result.OutputSize
		}
	}

	if summary.SuccessCount < len(summary.Results) {
		return summary, fmt.Errorf("some builds failed")
	}
	return summary, nil
}

// BuildAll builds all configured binaries.
func (s *Service) BuildAll(ctx context.Context) error {
	_, err := s.BuildWithResults(ctx, BuildOptions{
		Target:  "all",
		Version: s.config.Version,
	})
	return err
}

// BuildCLI builds the CLI binary.
func (s *Service) BuildCLI(ctx context.Context) error {
	_, err := s.BuildWithResults(ctx, BuildOptions{
		Target:  "cli",
		Version: s.config.Version,
	})
	return err
}

// BuildMCP builds the MCP server binary.
func (s *Service) BuildMCP(ctx context.Context) error {
	_, err := s.BuildWithResults(ctx, BuildOptions{
		Target:  "mcp",
		Version: s.config.Version,
	})
	return err
}

func (s *Service) ReleaseWithResults(ctx context.Context, version string) (*BuildSummary, error) {
	platforms := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	summary := &BuildSummary{Target: "release"}
	for _, platform := range platforms {
		for _, binary := range s.selectedBinaries("all") {
			config := build.DefaultConfig(binary.name, binary.path)
			config.Version = version
			config.GOOS = platform.goos
			config.GOARCH = platform.goarch
			config.OutputDir = fmt.Sprintf("dist/%s-%s-%s", binary.name, platform.goos, platform.goarch)
			config.LDFlags = append(config.LDFlags, fmt.Sprintf("-X main.version=%s", version))

			result := build.NewBuilder(config).BuildWithResult(ctx)
			summary.Results = append(summary.Results, result)
			if result.Success {
				summary.SuccessCount++
				summary.TotalDuration += result.Duration
				summary.TotalSize += result.OutputSize
			}
		}
	}

	if summary.SuccessCount < len(summary.Results) {
		return summary, fmt.Errorf("some builds failed")
	}
	return summary, nil
}

type binaryTarget struct {
	name string
	path string
}

func (s *Service) selectedBinaries(target string) []binaryTarget {
	switch target {
	case "cli":
		return []binaryTarget{{name: s.config.BinaryCLI, path: "./" + s.config.CmdCLIDir}}
	case "mcp":
		return []binaryTarget{{name: s.config.BinaryMCP, path: "./" + s.config.CmdMCPDir}}
	default:
		return []binaryTarget{
			{name: s.config.BinaryCLI, path: "./" + s.config.CmdCLIDir},
			{name: s.config.BinaryMCP, path: "./" + s.config.CmdMCPDir},
		}
	}
}
