package main

import (
	"fmt"
	"os"

	"github.com/SamyRai/juleson/cmd/orchestrator/commands"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildDate = "dev"
	gitCommit = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "orchestrator",
		Short: "Internal CLI tool for project orchestration",
		Long:  `A CLI tool that provides orchestration commands previously handled by Makefile`,
	}

	// Add all commands
	commands.AddCommands(rootCmd, version, buildDate, gitCommit)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
