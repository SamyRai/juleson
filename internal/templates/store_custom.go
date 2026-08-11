package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type CustomStore struct {
	customPath string
}

func NewCustomStore(customPath string) *CustomStore {
	return &CustomStore{customPath: customPath}
}

func (s *CustomStore) LoadRegistry() (*Registry, error) {
	if s.customPath == "" {
		return &Registry{}, nil
	}
	if _, err := os.Stat(s.customPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("custom templates path does not exist: %s", s.customPath)
	}

	registry := &Registry{}
	err := filepath.WalkDir(s.customPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".yaml") {
			return nil
		}

		template, err := s.LoadTemplate(path)
		if err != nil {
			return fmt.Errorf("failed to load custom template %s: %w", path, err)
		}

		registry.Templates = append(registry.Templates, RegistryTemplate{
			Name:         template.Metadata.Name,
			Version:      template.Metadata.Version,
			Category:     template.Metadata.Category,
			Description:  template.Metadata.Description,
			Author:       template.Metadata.Author,
			Tags:         template.Metadata.Tags,
			File:         path,
			Dependencies: []string{},
			Compatibility: RegistryCompatibility{
				Languages:  []string{"all"},
				Frameworks: []string{"all"},
			},
			Features:          []string{"custom"},
			Complexity:        "custom",
			EstimatedDuration: "custom",
		})
		return nil
	})

	return registry, err
}

func (s *CustomStore) LoadTemplate(filePath string) (*Template, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read custom template file: %w", err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom template: %w", err)
	}

	return &template, nil
}
