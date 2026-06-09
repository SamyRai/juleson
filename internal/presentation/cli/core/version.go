package core

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are set at build time using -ldflags.
var (
	// Version is the current version of Juleson.
	Version = "dev"
	// BuildDate is the build date (set at build time).
	BuildDate = "unknown"
	// GitCommit is the git commit hash (set at build time).
	GitCommit = "unknown"
	// JulesAPIVersion is the Jules API version.
	JulesAPIVersion = "v1alpha"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version information including build details and runtime information.`,
	RunE:  runVersion,
}

// VersionInfo contains all version details.
type VersionInfo struct {
	Version         string `json:"version"`
	BuildDate       string `json:"build_date"`
	GitCommit       string `json:"git_commit"`
	JulesAPIVersion string `json:"jules_api_version"`
	GoVersion       string `json:"go_version"`
	OS              string `json:"os"`
	Arch            string `json:"arch"`
}

// GetVersionInfo returns structured version information.
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:         Version,
		BuildDate:       BuildDate,
		GitCommit:       GitCommit,
		JulesAPIVersion: JulesAPIVersion,
		GoVersion:       runtime.Version(),
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
	}
}

// runVersion displays version information.
func runVersion(cmd *cobra.Command, args []string) error {
	info := GetVersionInfo()
	fmt.Print(FormatVersion(info))
	return nil
}

// FormatVersion returns the complete CLI version output.
func FormatVersion(info VersionInfo) string {
	return fmt.Sprintf("Juleson CLI %s\nJules API Version: %s\nBuild Date: %s\nGit Commit: %s\nGo Version: %s\nOS/Arch: %s/%s\n",
		info.Version,
		info.JulesAPIVersion,
		info.BuildDate,
		info.GitCommit,
		info.GoVersion,
		info.OS,
		info.Arch,
	)
}

// NewVersionCommand creates the version command.
func NewVersionCommand() *cobra.Command {
	return versionCmd
}
