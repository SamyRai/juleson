package main

import (
	"fmt"
	"os"

	"github.com/SamyRai/juleson/cmd/builder/commands"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildDate = "dev"
	gitCommit = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "builder",
		Short: "Internal CLI tool for project build workflows",
		Long:  `A CLI tool that provides build, test, install, and release workflows.`,
	}

	// Add all commands
	commands.AddCommands(rootCmd, version, buildDate, gitCommit)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
