package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SamyRai/juleson/internal/cli"
	"github.com/SamyRai/juleson/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
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
