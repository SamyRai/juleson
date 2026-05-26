package main

import (
	"testing"

	"github.com/SamyRai/juleson/internal/cli/commands"
)

func TestApplyBuildMetadata(t *testing.T) {
	oldVersion := commands.Version
	oldBuildDate := commands.BuildDate
	oldGitCommit := commands.GitCommit
	oldMainVersion := version
	oldBuildTime := buildTime
	oldMainGitCommit := gitCommit
	t.Cleanup(func() {
		commands.Version = oldVersion
		commands.BuildDate = oldBuildDate
		commands.GitCommit = oldGitCommit
		version = oldMainVersion
		buildTime = oldBuildTime
		gitCommit = oldMainGitCommit
	})

	version = "v0.1.0"
	buildTime = "2026-05-26T07:00:00Z"
	gitCommit = "abc123"

	applyBuildMetadata()

	if commands.Version != version {
		t.Fatalf("Version = %q, want %q", commands.Version, version)
	}
	if commands.BuildDate != buildTime {
		t.Fatalf("BuildDate = %q, want %q", commands.BuildDate, buildTime)
	}
	if commands.GitCommit != gitCommit {
		t.Fatalf("GitCommit = %q, want %q", commands.GitCommit, gitCommit)
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
