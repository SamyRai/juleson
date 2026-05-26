package main

import (
	"strings"
	"testing"
)

func TestShouldPrintVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "version command", args: []string{"version"}, want: true},
		{name: "version flag", args: []string{"--version"}, want: true},
		{name: "short version flag", args: []string{"-v"}, want: true},
		{name: "no args", args: nil, want: false},
		{name: "extra args", args: []string{"--version", "--verbose"}, want: false},
		{name: "help", args: []string{"--help"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPrintVersion(tt.args); got != tt.want {
				t.Fatalf("shouldPrintVersion(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestVersionTextUsesBuildMetadata(t *testing.T) {
	oldVersion := version
	oldBuildTime := buildTime
	oldGitCommit := gitCommit
	t.Cleanup(func() {
		version = oldVersion
		buildTime = oldBuildTime
		gitCommit = oldGitCommit
	})

	version = "v0.1.0"
	buildTime = "2026-05-26T07:00:00Z"
	gitCommit = "abc123"

	output := versionText()
	for _, want := range []string{
		"Juleson MCP v0.1.0",
		"Jules API Version: v1alpha",
		"Build Date: 2026-05-26T07:00:00Z",
		"Git Commit: abc123",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("version output missing %q:\n%s", want, output)
		}
	}
}
