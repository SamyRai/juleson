package main

import (
	"testing"

	"github.com/SamyRai/juleson/internal/presentation/cli/core"
)

func TestApplyBuildMetadata(t *testing.T) {
	oldVersion := core.Version
	oldBuildDate := core.BuildDate
	oldGitCommit := core.GitCommit
	oldMainVersion := version
	oldBuildTime := buildTime
	oldMainGitCommit := gitCommit
	t.Cleanup(func() {
		core.Version = oldVersion
		core.BuildDate = oldBuildDate
		core.GitCommit = oldGitCommit
		version = oldMainVersion
		buildTime = oldBuildTime
		gitCommit = oldMainGitCommit
	})

	version = "test-version"
	buildTime = "2023-10-27"
	gitCommit = "abcdef1"

	applyBuildMetadata()

	if core.Version != "test-version" {
		t.Errorf("expected Version to be 'test-version', got %s", core.Version)
	}
	if core.BuildDate != "2023-10-27" {
		t.Errorf("expected BuildDate to be '2023-10-27', got %s", core.BuildDate)
	}
	if core.GitCommit != "abcdef1" {
		t.Errorf("expected GitCommit to be 'abcdef1', got %s", core.GitCommit)
	}
}

func TestIsConfigValidateCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "config validate",
			args: []string{"config", "validate"},
			want: true,
		},
		{
			name: "config validate with flags",
			args: []string{"config", "validate", "--help"},
			want: true,
		},
		{
			name: "config parent",
			args: []string{"config"},
			want: false,
		},
		{
			name: "other command",
			args: []string{"sessions", "list"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isConfigValidateCommand(tt.args); got != tt.want {
				t.Fatalf("isConfigValidateCommand(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
