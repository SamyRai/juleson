package orchestrator

import (
	"time"

	"github.com/SamyRai/juleson/internal/build"
)

type BuildOptions struct {
	Target  string
	Version string
	Race    bool
	GOOS    string
	GOARCH  string
}

type BuildSummary struct {
	Target        string
	Results       []*build.BuildResult
	SuccessCount  int
	TotalDuration time.Duration
	TotalSize     int64
}

type CleanOptions struct {
	All       bool
	Cache     bool
	ModCache  bool
	TestCache bool
}

type InstallOptions struct {
	Path       string
	SkipBuild  bool
	SkipChecks bool
	SkipLint   bool
	SkipTests  bool
}

type InstallResult = build.InstallResult
