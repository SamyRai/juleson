package commands

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
	fmt.Printf("Juleson CLI %s\n", Version)
	fmt.Printf("Jules API Version: %s\n", JulesAPIVersion)
	fmt.Printf("Build Date: %s\n", BuildDate)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	return nil
}

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	return versionCmd
}
