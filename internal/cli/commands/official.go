package commands

import (
	"os"

	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/spf13/cobra"
)

// NewOfficialCommand creates optional passthroughs to the official Jules CLI.
func NewOfficialCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "official",
		Short: "Bridge to the official Jules CLI when installed",
		Long:  "Bridge to the official Jules CLI for exact official behavior such as remote new --parallel, remote pull, and TUI handoff.",
	}

	remoteCmd := &cobra.Command{
		Use:                "remote [args...]",
		Short:              "Run official 'jules remote' commands",
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOfficialJules(append([]string{"remote"}, args...)...)
		},
	}
	cmd.AddCommand(remoteCmd)

	tuiCmd := &cobra.Command{
		Use:                "tui [args...]",
		Short:              "Open the official Jules CLI TUI",
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOfficialJules(args...)
		},
	}
	cmd.AddCommand(tuiCmd)

	return cmd
}

func runOfficialJules(args ...string) error {
	return julesops.RunOfficialJulesCLI(julesops.OfficialCLIStreams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}, args...)
}
