package main

import (
	"log"

	"jules-automation/internal/config"
	"jules-automation/internal/mcp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create MCP server using official SDK
	server := mcp.NewServer(cfg)

	// Start server (runs over stdin/stdout)
	log.Println("Starting Jules Automation MCP Server...")
	log.Println("Server will run over stdin/stdout transport")
	
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}
