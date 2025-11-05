package cli

import (
	"github.com/SamyRai/juleson/internal/cli/commands"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/presentation"
	"github.com/SamyRai/juleson/internal/services"

	"github.com/spf13/cobra"
)

// App represents the CLI application
type App struct {
	container  *services.Container
	formatters *Formatters
	rootCmd    *cobra.Command
}

// Formatters holds all presentation formatters
type Formatters struct {
	Analysis  *presentation.ProjectAnalysisFormatter
	Template  *presentation.TemplateFormatter
	Execution *presentation.ExecutionFormatter
	Session   *presentation.SessionFormatter
	ConfigGen *presentation.ConfigGenerator
}

// NewApp creates a new CLI application with dependency injection
func NewApp(cfg *config.Config) *App {
	app := &App{
		container: services.NewContainer(cfg),
		formatters: &Formatters{
			Analysis:  presentation.NewProjectAnalysisFormatter(),
			Template:  presentation.NewTemplateFormatter(),
			Execution: presentation.NewExecutionFormatter(),
			Session:   presentation.NewSessionFormatter(),
			ConfigGen: presentation.NewConfigGenerator(),
		},
	}

	app.setupCommands()
	return app
}

// Execute runs the CLI application
func (a *App) Execute() error {
	return a.rootCmd.Execute()
}

// setupCommands configures all CLI commands with proper dependency injection
func (a *App) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:   "juleson",
		Short: "Jules automation CLI tool",
		Long:  "A comprehensive CLI tool for automating project tasks using Google's Jules AI coding agent",
	}

	// Customize help template to reduce redundancy
	a.rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
`)

	// Register commands with their dependencies
	a.rootCmd.AddCommand(commands.NewSetupCommand())
	a.rootCmd.AddCommand(commands.NewInitCommand(a.formatters.ConfigGen.GenerateProjectConfig))
	a.rootCmd.AddCommand(commands.NewAnalyzeCommand(
		a.container.AutomationEngine,
		commands.DisplayProjectAnalysis,
	))
	a.rootCmd.AddCommand(commands.NewTemplateCommand(
		a.container.TemplateManager,
		commands.DisplayTemplates,
		commands.DisplayTemplateDetails,
	))
	a.rootCmd.AddCommand(commands.NewExecuteCommand(
		a.container.AutomationEngine,
		commands.DisplayExecutionResult,
	))
	a.rootCmd.AddCommand(commands.NewSyncCommand())
	a.rootCmd.AddCommand(commands.NewSessionsCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewSourcesCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewActivitiesCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewPRCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewGitHubCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewActionsCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewOrchestrateCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewAIOrchestCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewAgentCommand(a.container.Config()))
	a.rootCmd.AddCommand(commands.NewDevCommand())
	a.rootCmd.AddCommand(commands.NewCompletionCommand())
	a.rootCmd.AddCommand(commands.NewVersionCommand())
}
