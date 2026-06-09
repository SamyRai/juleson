package jmcp

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/pkg/build"
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type devProvider struct {
	svc *builder.Service
}

// NewDevProvider creates a ToolProvider for development commands.
func NewDevProvider(svc *builder.Service) ToolProvider {
	return &devProvider{svc: svc}
}

func (p *devProvider) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dev_build",
		Description: "Build Juleson binaries. Target is all, cli, or alias.",
	}, p.devBuild)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dev_test",
		Description: "Run Juleson Go tests through the builder service.",
	}, p.devTest)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dev_check",
		Description: "Run Juleson quality checks through the builder service. Requires confirm=true because formatting may modify files.",
	}, p.devCheck)
}

type devBuildInput struct {
	Target  string `json:"target,omitempty" jsonschema:"all, cli, or alias"`
	Version string `json:"version,omitempty"`
	GOOS    string `json:"goos,omitempty"`
	GOARCH  string `json:"goarch,omitempty"`
	Race    bool   `json:"race,omitempty"`
}

func (p *devProvider) devBuild(ctx context.Context, _ *mcp.CallToolRequest, in devBuildInput) (*mcp.CallToolResult, *builder.BuildSummary, error) {
	target := in.Target
	if target == "" {
		target = "all"
	}
	if target != "all" && target != "cli" && target != "alias" {
		return nil, nil, fmt.Errorf("target must be all, cli, or alias")
	}
	version := in.Version
	if version == "" {
		version = "dev"
	}
	summary, err := p.svc.BuildWithResults(ctx, builder.BuildOptions{
		Target:  target,
		Version: version,
		GOOS:    in.GOOS,
		GOARCH:  in.GOARCH,
		Race:    in.Race,
	})
	return nil, summary, err
}

type devTestInput struct {
	RunPattern     *string  `json:"run_pattern,omitempty"`
	SkipPattern    *string  `json:"skip_pattern,omitempty"`
	Shuffle        *string  `json:"shuffle,omitempty"`
	Packages       []string `json:"packages,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
	Verbose        bool     `json:"verbose,omitempty"`
	Race           bool     `json:"race,omitempty"`
	Cover          bool     `json:"cover,omitempty"`
	Short          bool     `json:"short,omitempty"`
	FailFast       bool     `json:"fail_fast,omitempty"`
}

func (p *devProvider) devTest(ctx context.Context, _ *mcp.CallToolRequest, in devTestInput) (*mcp.CallToolResult, *build.TestResult, error) {
	testConfig := builder.DefaultTestConfig()
	testConfig.Verbose = in.Verbose
	testConfig.Race = in.Race
	testConfig.Cover = in.Cover
	testConfig.Short = in.Short
	testConfig.RunPattern = optionalString(in.RunPattern)
	testConfig.SkipPattern = optionalString(in.SkipPattern)
	testConfig.FailFast = in.FailFast
	testConfig.Shuffle = optionalString(in.Shuffle)
	testConfig.Packages = in.Packages
	if in.TimeoutSeconds > 0 {
		testConfig.Timeout = time.Duration(in.TimeoutSeconds) * time.Second
	}
	result := p.svc.RunTestsWithResult(ctx, testConfig)
	return nil, result, result.Error
}

type devCheckInput struct {
	Confirm bool `json:"confirm"`
}

func (p *devProvider) devCheck(ctx context.Context, _ *mcp.CallToolRequest, in devCheckInput) (*mcp.CallToolResult, *builder.QualitySummary, error) {
	if err := requireConfirm(in.Confirm, "dev_check"); err != nil {
		return nil, nil, err
	}
	testConfig := builder.DefaultTestConfig()
	testConfig.Cover = true
	testConfig.CoverProfile = "coverage.out"
	summary, err := p.svc.RunQualityChecks(ctx, builder.QualityOptions{
		Format:     true,
		Lint:       true,
		Test:       true,
		TestConfig: testConfig,
		Build:      true,
	})
	return nil, summary, err
}
