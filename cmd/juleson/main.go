package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/logger"
	"github.com/SamyRai/juleson/internal/presentation/cli"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	applyBuildMetadata()

	// Load configuration. Commands that require Jules API access validate
	// credentials at use time; local commands such as version/help should work
	// without JULES_API_KEY.
	cfg, err := loadConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create CLI application
	app := cli.NewApp(cfg)

	// Setup global logger
	logger.SetupGlobal(cfg.Jules.DebugLog)

	// Execute CLI
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func applyBuildMetadata() {
	core.Version = version
	core.BuildDate = buildTime
	core.GitCommit = gitCommit
}

func loadConfig(args []string) (*config.Config, error) {
	if isConfigValidateCommand(args) {
		return config.LoadForValidation()
	}
	return config.LoadOptional()
}

func isConfigValidateCommand(args []string) bool {
	return len(args) >= 2 && args[0] == "config" && args[1] == "validate"
}
