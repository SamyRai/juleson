package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SamyRai/juleson/internal/cli"
	"github.com/SamyRai/juleson/internal/config"
)

func main() {
	// Load configuration. Commands that require Jules API access validate
	// credentials at use time; local commands such as version/help should work
	// without JULES_API_KEY.
	cfg, err := config.LoadOptional()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create CLI application
	app := cli.NewApp(cfg)

	// Execute CLI
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
