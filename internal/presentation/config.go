package presentation

import (
	"fmt"
	"path/filepath"
)

// ConfigGenerator generates project configuration files
type ConfigGenerator struct{}

// NewConfigGenerator creates a new config generator
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// GenerateProjectConfig generates a YAML configuration for a new project
func (g *ConfigGenerator) GenerateProjectConfig(projectPath string) string {
	return fmt.Sprintf(`# Juleson Project Configuration

project:
  name: "%s"
  path: "%s"
  type: "auto-detect"

jules:
  api_key: "${JULES_API_KEY}"
  base_url: "https://jules.googleapis.com/v1alpha"
  timeout: "30s"

automation:
  strategies:
    - "modular"
    - "layered"
    - "microservices"
  max_concurrent_tasks: 3
  backup_enabled: true

templates:
  custom_path: "./templates/custom"
  builtin_enabled: true

git:
  integration: true
  auto_commit: false
  commit_message_template: "Jules automation: {{.TemplateName}}"
`, filepath.Base(projectPath), projectPath)
}
