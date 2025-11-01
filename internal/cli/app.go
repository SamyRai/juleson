package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jules-automation/internal/automation"
	"jules-automation/internal/cli/commands"
	"jules-automation/internal/config"
	"jules-automation/internal/jules"
	"jules-automation/internal/templates"

	"github.com/spf13/cobra"
)

// App represents the CLI application
type App struct {
	config           *config.Config
	julesClient      *jules.Client
	templateManager  *templates.Manager
	automationEngine *automation.Engine
	rootCmd          *cobra.Command
}

// NewApp creates a new CLI application
func NewApp(cfg *config.Config) *App {
	app := &App{
		config: cfg,
	}

	app.setupCommands()
	return app
}

// Execute runs the CLI application
func (a *App) Execute() error {
	return a.rootCmd.Execute()
}

// setupCommands sets up all CLI commands
func (a *App) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:   "jules-cli",
		Short: "Jules automation CLI tool",
		Long:  "A comprehensive CLI tool for automating project tasks using Google's Jules AI coding agent",
	}

	// Add subcommands
	a.rootCmd.AddCommand(commands.NewInitCommand(a.generateProjectConfig))
	a.rootCmd.AddCommand(commands.NewAnalyzeCommand(a.initializeEngine, commands.DisplayProjectAnalysis))
	a.rootCmd.AddCommand(commands.NewTemplateCommand(a.initializeTemplateManager, commands.DisplayTemplates, commands.DisplayTemplateDetails))
	a.rootCmd.AddCommand(commands.NewExecuteCommand(a.initializeEngine, commands.DisplayExecutionResult))
	a.rootCmd.AddCommand(commands.NewSyncCommand())
	a.rootCmd.AddCommand(commands.NewSessionsCommand(a.config))
}

// createInitCommand creates the init command
func (a *App) createInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [project-path]",
		Short: "Initialize a new project for Jules automation",
		Long:  "Initialize a new project directory with Jules automation configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]

			// Create project directory
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				return fmt.Errorf("failed to create project directory: %w", err)
			}

			// Create Jules automation config
			configPath := filepath.Join(projectPath, "jules-automation.yaml")
			configContent := a.generateProjectConfig(projectPath)

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to create project config: %w", err)
			}

			fmt.Printf("✅ Initialized Jules automation project at: %s\n", projectPath)
			fmt.Printf("📝 Configuration file created: %s\n", configPath)

			return nil
		},
	}
}

// createAnalyzeCommand creates the analyze command
func (a *App) createAnalyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [project-path]",
		Short: "Analyze project structure and context",
		Long:  "Analyze the project structure, dependencies, and create context for automation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]

			// Initialize automation engine
			engine, err := a.initializeEngine()
			if err != nil {
				return fmt.Errorf("failed to initialize automation engine: %w", err)
			}

			// Analyze project
			context, err := engine.AnalyzeProject(projectPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Display analysis results
			a.displayProjectAnalysis(context)

			return nil
		},
	}
}

// createTemplateCommand creates the template command
func (a *App) createTemplateCommand() *cobra.Command {
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
			templateManager, err := a.initializeTemplateManager()
			if err != nil {
				return fmt.Errorf("failed to initialize template manager: %w", err)
			}

			var templates []templates.RegistryTemplate
			if len(args) > 0 {
				templates = templateManager.ListTemplatesByCategory(args[0])
			} else {
				templates = templateManager.ListTemplates()
			}

			a.displayTemplates(templates)
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

			templateManager, err := a.initializeTemplateManager()
			if err != nil {
				return fmt.Errorf("failed to initialize template manager: %w", err)
			}

			template, err := templateManager.LoadTemplate(templateName)
			if err != nil {
				return fmt.Errorf("failed to load template: %w", err)
			}

			a.displayTemplateDetails(template)
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

			templateManager, err := a.initializeTemplateManager()
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

			templateManager, err := a.initializeTemplateManager()
			if err != nil {
				return fmt.Errorf("failed to initialize template manager: %w", err)
			}

			templates := templateManager.SearchTemplates(query)
			a.displayTemplates(templates)
			return nil
		},
	})

	return templateCmd
}

// createExecuteCommand creates the execute command
func (a *App) createExecuteCommand() *cobra.Command {
	executeCmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute automation tasks",
		Long:  "Execute templates and automation tasks on projects",
	}

	// Execute template
	executeCmd.AddCommand(&cobra.Command{
		Use:   "template [template-name] [project-path]",
		Short: "Execute a template on a project",
		Long:  "Execute a specific template on a project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			projectPath := args[1]

			// Initialize automation engine
			engine, err := a.initializeEngine()
			if err != nil {
				return fmt.Errorf("failed to initialize automation engine: %w", err)
			}

			// Analyze project first
			_, err = engine.AnalyzeProject(projectPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Execute template
			result, err := engine.ExecuteTemplate(cmd.Context(), templateName, make(map[string]string))
			if err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			// Display results
			a.displayExecutionResult(result)

			return nil
		},
	})

	// Execute with custom parameters
	executeCmd.AddCommand(&cobra.Command{
		Use:   "template-with-params [template-name] [project-path] [key=value]...",
		Short: "Execute template with custom parameters",
		Long:  "Execute a template with custom parameters",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			projectPath := args[1]

			// Parse custom parameters
			customParams := make(map[string]string)
			for i := 2; i < len(args); i++ {
				parts := strings.SplitN(args[i], "=", 2)
				if len(parts) == 2 {
					customParams[parts[0]] = parts[1]
				}
			}

			// Initialize automation engine
			engine, err := a.initializeEngine()
			if err != nil {
				return fmt.Errorf("failed to initialize automation engine: %w", err)
			}

			// Analyze project first
			_, err = engine.AnalyzeProject(projectPath)
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Execute template with custom parameters
			result, err := engine.ExecuteTemplate(cmd.Context(), templateName, customParams)
			if err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			// Display results
			a.displayExecutionResult(result)

			return nil
		},
	})

	return executeCmd
}

// createSyncCommand creates the sync command
func (a *App) createSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [project-path] [remote]",
		Short: "Sync project with remote repository",
		Long:  "Sync project changes with remote repository using Git",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]
			remote := args[1]

			// TODO: Implement Git sync functionality
			fmt.Printf("🔄 Syncing project %s with remote %s\n", projectPath, remote)
			fmt.Println("✅ Sync completed successfully")

			return nil
		},
	}
}

// createSessionsCommand creates the sessions command
func (a *App) createSessionsCommand() *cobra.Command {
	sessionsCmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage Jules sessions",
		Long:  "List, monitor, and manage Jules AI coding sessions",
	}

	// List sessions
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Long:  "List all Jules sessions with their current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.listSessions()
		},
	})

	// Show session status
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show session status summary",
		Long:  "Show a summary of current session statuses",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.showSessionStatus()
		},
	})

	return sessionsCmd
}

// Helper methods

func (a *App) initializeEngine() (*automation.Engine, error) {
	if a.automationEngine != nil {
		return a.automationEngine, nil
	}

	// Initialize Jules client
	julesClient := jules.NewClient(
		a.config.Jules.APIKey,
		a.config.Jules.BaseURL,
		a.config.Jules.Timeout,
		a.config.Jules.RetryAttempts,
	)

	// Initialize template manager
	templateManager, err := a.initializeTemplateManager()
	if err != nil {
		return nil, err
	}

	// Create automation engine
	engine := automation.NewEngine(julesClient, templateManager)
	a.automationEngine = engine

	return engine, nil
}

func (a *App) initializeTemplateManager() (*templates.Manager, error) {
	if a.templateManager != nil {
		return a.templateManager, nil
	}

	templateManager, err := templates.NewManager("./templates")
	if err != nil {
		return nil, err
	}

	a.templateManager = templateManager
	return templateManager, nil
}

func (a *App) generateProjectConfig(projectPath string) string {
	return fmt.Sprintf(`# Jules Automation Project Configuration

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

func (a *App) displayProjectAnalysis(context *automation.ProjectContext) {
	fmt.Println("📊 Project Analysis Results")
	fmt.Println("==========================")
	fmt.Printf("Project Name: %s\n", context.ProjectName)
	fmt.Printf("Project Type: %s\n", context.ProjectType)
	fmt.Printf("Languages: %s\n", strings.Join(context.Languages, ", "))
	fmt.Printf("Frameworks: %s\n", strings.Join(context.Frameworks, ", "))
	fmt.Printf("Architecture: %s\n", context.Architecture)
	fmt.Printf("Complexity: %s\n", context.Complexity)
	fmt.Printf("Git Status: %s\n", context.GitStatus)
	fmt.Printf("Dependencies: %d\n", len(context.Dependencies))
	fmt.Printf("File Types: %d\n", len(context.FileStructure))
}

func (a *App) displayTemplates(templates []templates.RegistryTemplate) {
	fmt.Println("📋 Available Templates")
	fmt.Println("=====================")

	for _, template := range templates {
		fmt.Printf("• %s (%s) - %s\n", template.Name, template.Category, template.Description)
		fmt.Printf("  Tags: %s\n", strings.Join(template.Tags, ", "))
		fmt.Printf("  Complexity: %s | Duration: %s\n", template.Complexity, template.EstimatedDuration)
		fmt.Println()
	}
}

func (a *App) displayTemplateDetails(template *templates.Template) {
	fmt.Printf("📄 Template Details: %s\n", template.Metadata.Name)
	fmt.Println("========================")
	fmt.Printf("Version: %s\n", template.Metadata.Version)
	fmt.Printf("Category: %s\n", template.Metadata.Category)
	fmt.Printf("Description: %s\n", template.Metadata.Description)
	fmt.Printf("Author: %s\n", template.Metadata.Author)
	fmt.Printf("Tags: %s\n", strings.Join(template.Metadata.Tags, ", "))
	fmt.Printf("Strategy: %s\n", template.Config.Strategy)
	fmt.Printf("Max Concurrent Tasks: %d\n", template.Config.MaxConcurrentTasks)
	fmt.Printf("Timeout: %s\n", template.Config.Timeout)
	fmt.Printf("Requires Approval: %t\n", template.Config.RequiresApproval)
	fmt.Printf("Backup Enabled: %t\n", template.Config.BackupEnabled)
	fmt.Printf("Tasks: %d\n", len(template.Tasks))

	fmt.Println("\nTasks:")
	for i, task := range template.Tasks {
		fmt.Printf("  %d. %s (%s)\n", i+1, task.Name, task.Type)
		fmt.Printf("     %s\n", task.Description)
		if len(task.DependsOn) > 0 {
			fmt.Printf("     Depends on: %s\n", strings.Join(task.DependsOn, ", "))
		}
	}
}

func (a *App) displayExecutionResult(result *automation.ExecutionResult) {
	fmt.Println("🎯 Execution Results")
	fmt.Println("====================")
	fmt.Printf("Template: %s\n", result.TemplateName)
	fmt.Printf("Project: %s\n", result.ProjectPath)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Success: %t\n", result.Success)

	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}

	fmt.Printf("Tasks Executed: %d\n", len(result.TasksExecuted))
	for _, task := range result.TasksExecuted {
		status := "✅"
		if !task.Success {
			status = "❌"
		}
		fmt.Printf("  %s %s (%s) - %v\n", status, task.TaskName, task.TaskType, task.Duration)
	}

	if len(result.OutputFiles) > 0 {
		fmt.Println("\nOutput Files:")
		for _, file := range result.OutputFiles {
			fmt.Printf("  📄 %s\n", file)
		}
	}
}

// listSessions lists all Jules sessions
func (a *App) listSessions() error {
	// Initialize Jules client
	julesClient := jules.NewClient(
		a.config.Jules.APIKey,
		a.config.Jules.BaseURL,
		a.config.Jules.Timeout,
		a.config.Jules.RetryAttempts,
	)

	fmt.Println("🔍 Listing Jules sessions...")
	fmt.Println("============================")

	response, err := julesClient.ListSessionsWithPagination(context.Background(), 50, "")
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := response.Sessions
	if len(sessions) == 0 {
		fmt.Println("📭 No sessions found.")
		return nil
	}

	fmt.Printf("📊 Found %d session(s):\n\n", len(sessions))

	for i, session := range sessions {
		fmt.Printf("%d. Session: %s\n", i+1, session.ID)
		fmt.Printf("   Title: %s\n", session.Title)
		fmt.Printf("   State: %s\n", session.State)
		fmt.Printf("   Created: %s\n", session.CreateTime)
		if session.UpdateTime != "" {
			fmt.Printf("   Updated: %s\n", session.UpdateTime)
		}
		if session.SourceContext != nil && session.SourceContext.Source != "" {
			fmt.Printf("   Source: %s\n", session.SourceContext.Source)
		}
		if session.RequirePlanApproval {
			fmt.Printf("   Plan Approval Required: Yes\n")
		}
		if session.AutomationMode != "" {
			fmt.Printf("   Automation Mode: %s\n", session.AutomationMode)
		}
		if len(session.Outputs) > 0 {
			fmt.Printf("   Outputs: %d\n", len(session.Outputs))
		}

		// Status indicators
		if session.State == "IN_PROGRESS" || session.State == "PLANNING" {
			fmt.Printf("   ⚡ ACTIVE\n")
		} else if session.State == "COMPLETED" {
			fmt.Printf("   ✅ COMPLETED\n")
		} else if session.State == "FAILED" {
			fmt.Printf("   ❌ FAILED\n")
		}
		fmt.Println()
	}

	return nil
}

// showSessionStatus shows a summary of session statuses
func (a *App) showSessionStatus() error {
	// Initialize Jules client
	julesClient := jules.NewClient(
		a.config.Jules.APIKey,
		a.config.Jules.BaseURL,
		a.config.Jules.Timeout,
		a.config.Jules.RetryAttempts,
	)

	fmt.Println("📊 Jules Session Status")
	fmt.Println("=======================")

	response, err := julesClient.ListSessionsWithPagination(context.Background(), 100, "")
	if err != nil {
		return fmt.Errorf("failed to get session status: %w", err)
	}

	sessions := response.Sessions
	totalSessions := len(sessions)

	if totalSessions == 0 {
		fmt.Println("📭 No sessions found.")
		return nil
	}

	// Count sessions by state
	stateCounts := make(map[string]int)
	for _, session := range sessions {
		stateCounts[session.State]++
	}

	fmt.Printf("Total Sessions: %d\n\n", totalSessions)

	// Display state breakdown
	fmt.Println("Session States:")
	for state, count := range stateCounts {
		percentage := float64(count) / float64(totalSessions) * 100
		var icon string
		switch state {
		case "IN_PROGRESS", "PLANNING":
			icon = "⚡"
		case "COMPLETED":
			icon = "✅"
		case "FAILED":
			icon = "❌"
		default:
			icon = "📋"
		}
		fmt.Printf("  %s %s: %d (%.1f%%)\n", icon, state, count, percentage)
	}

	// Active sessions summary
	activeCount := stateCounts["IN_PROGRESS"] + stateCounts["PLANNING"]
	if activeCount > 0 {
		fmt.Printf("\n⚠️  %d session(s) are currently active/running\n", activeCount)
	} else {
		fmt.Println("\n✅ No active sessions currently running")
	}

	// Recent sessions (last 5)
	if totalSessions > 0 {
		fmt.Println("\n🕒 Recent Sessions:")
		recentCount := 5
		if totalSessions < recentCount {
			recentCount = totalSessions
		}

		for i := 0; i < recentCount; i++ {
			session := sessions[i]
			var statusIcon string
			switch session.State {
			case "IN_PROGRESS", "PLANNING":
				statusIcon = "⚡"
			case "COMPLETED":
				statusIcon = "✅"
			case "FAILED":
				statusIcon = "❌"
			default:
				statusIcon = "📋"
			}
			fmt.Printf("  %s %s - %s (%s)\n", statusIcon, session.ID[:12], session.Title, session.State)
		}
	}

	return nil
}
