package commands

import (
	"fmt"

	"jules-automation/internal/templates"

	"github.com/spf13/cobra"
)

// NewTemplateCommand creates the template command
func NewTemplateCommand(initializeTemplateManager func() (*templates.Manager, error), displayTemplates func([]templates.RegistryTemplate), displayTemplateDetails func(*templates.Template)) *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage templates",
		Long:  "List, create, and manage Jules automation templates",
	}

	// List templates
	templateCmd.AddCommand(&cobra.Command{
		Use:   "list [category]",
		Short: "List available templates",
		Long:  "List all available templates, optionally filtered by category",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateManager, err := initializeTemplateManager()
			if err != nil {
				return fmt.Errorf("failed to initialize template manager: %w", err)
			}

			var templates []templates.RegistryTemplate
			if len(args) > 0 {
				templates = templateManager.ListTemplatesByCategory(args[0])
			} else {
				templates = templateManager.ListTemplates()
			}

			displayTemplates(templates)
			return nil
		},
	})

	// Show template details
	templateCmd.AddCommand(&cobra.Command{
		Use:   "show [template-name]",
		Short: "Show template details",
		Long:  "Show detailed information about a specific template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

			templateManager, err := initializeTemplateManager()
			if err != nil {
				return fmt.Errorf("failed to initialize template manager: %w", err)
			}

			template, err := templateManager.LoadTemplate(templateName)
			if err != nil {
				return fmt.Errorf("failed to load template: %w", err)
			}

			displayTemplateDetails(template)
			return nil
		},
	})

	// Create template
	templateCmd.AddCommand(&cobra.Command{
		Use:   "create [template-name] [category] [description]",
		Short: "Create a new template",
		Long:  "Create a new custom template",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			category := args[1]
			description := args[2]

			templateManager, err := initializeTemplateManager()
			if err != nil {
				return fmt.Errorf("failed to initialize template manager: %w", err)
			}

			template, err := templateManager.CreateTemplate(templateName, category, description)
			if err != nil {
				return fmt.Errorf("failed to create template: %w", err)
			}

			if err := templateManager.SaveTemplate(template); err != nil {
				return fmt.Errorf("failed to save template: %w", err)
			}

			fmt.Printf("✅ Created template '%s' in category '%s'\n", templateName, category)
			return nil
		},
	})

	// Search templates
	templateCmd.AddCommand(&cobra.Command{
		Use:   "search [query]",
		Short: "Search templates",
		Long:  "Search templates by name, description, or tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			templateManager, err := initializeTemplateManager()
			if err != nil {
				return fmt.Errorf("failed to initialize template manager: %w", err)
			}

			templates := templateManager.SearchTemplates(query)
			displayTemplates(templates)
			return nil
		},
	})

	return templateCmd
}
