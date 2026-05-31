package core

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are set at build time using -ldflags
var (
	// Version is the current version of Juleson
	Version = "dev"
	// BuildDate is the build date (set at build time)
	BuildDate = "unknown"
	// GitCommit is the git commit hash (set at build time)
	GitCommit = "unknown"
	// JulesAPIVersion is the Jules API version
	JulesAPIVersion = "v1alpha"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version information including build details and runtime information.`,
	RunE:  runVersion,
}

// runVersion displays version information
func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Print(VersionText())
	return nil
}

// VersionText returns the complete CLI version output.
func VersionText() string {
	return fmt.Sprintf("Juleson CLI %s\nJules API Version: %s\nBuild Date: %s\nGit Commit: %s\nGo Version: %s\nOS/Arch: %s/%s\n",
		Version,
		JulesAPIVersion,
		BuildDate,
		GitCommit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	return versionCmd
}
