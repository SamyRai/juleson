package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type EmbeddedStore struct {
	templatesDir string
}

func NewEmbeddedStore(templatesDir string) *EmbeddedStore {
	return &EmbeddedStore{templatesDir: templatesDir}
}

func (s *EmbeddedStore) LoadRegistry() (*Registry, error) {
	var registryPath string
	if s.templatesDir == "" {
		registryPath = filepath.Join("templates", "registry", "registry.yaml")
	} else {
		registryPath = filepath.Join(filepath.Dir(s.templatesDir), "registry", "registry.yaml")
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file %s: %w", registryPath, err)
	}

	var registry Registry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registry: %w", err)
	}

	return &registry, nil
}

func (s *EmbeddedStore) LoadTemplate(filePath string) (*Template, error) {
	var templatePath string
	if s.templatesDir == "" {
		templatePath = filepath.Join("templates", filePath)
	} else {
		relativePath := strings.TrimPrefix(filePath, "builtin/")
		templatePath = filepath.Join(s.templatesDir, relativePath)
	}

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &template, nil
}
