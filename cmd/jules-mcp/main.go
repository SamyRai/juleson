package main

import (
	"log"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/mcp"
)

func main() {
	// Logging goes to stderr by default, which is correct for MCP stdio transport
	// MCP JSON-RPC messages go to stdout, logs to stderr
	log.SetPrefix("[juleson-mcp] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	log.Println("Starting Juleson MCP Server...")

	// Load configuration (allow missing for MCP context)
	cfg, err := loadConfigForMCP()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")

	// Create MCP server using official SDK
	server := mcp.NewServer(cfg)

	log.Println("MCP server initialized, starting transport...")

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}

	log.Println("MCP server stopped")
}

// loadConfigForMCP loads configuration for MCP context, allowing missing API keys
func loadConfigForMCP() (*config.Config, error) {
	log.Println("Loading configuration...")

	// Try to load config normally
	cfg, err := config.Load()
	if err != nil {
		// If config loading fails due to missing API key, create a minimal config
		if strings.Contains(err.Error(), "Jules API key is required") {
			log.Println("Warning: Jules API key not configured. Some tools may not be available.")
			log.Println("Using minimal configuration")
			return createMinimalConfig(), nil
		}
		return nil, err
	}

	log.Printf("Configuration loaded from file")
	return cfg, nil
}

// createMinimalConfig creates a minimal configuration for MCP when API key is missing
func createMinimalConfig() *config.Config {
	return &config.Config{
		Jules: config.JulesConfig{
			APIKey:        "", // Will be empty, tools that need it will handle gracefully
			BaseURL:       "https://jules.googleapis.com/v1alpha",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
		MCP: config.MCPConfig{
			Server: config.MCPServerConfig{
				Port: 8080,
				Host: "localhost",
			},
			Client: config.MCPClientConfig{
				Timeout: 10 * time.Second,
			},
		},
		Automation: config.AutomationConfig{
			Strategies:         []string{"modular", "layered", "microservices"},
			MaxConcurrentTasks: 5,
			TaskTimeout:        300 * time.Second,
		},
		Projects: config.ProjectsConfig{
			DefaultPath:    "./projects",
			BackupEnabled:  true,
			GitIntegration: true,
		},
		Templates: config.TemplatesConfig{
			BuiltinPath:  "./templates/builtin",
			CustomPath:   "./templates/custom",
			EnableCustom: true,
		},
	}
}
