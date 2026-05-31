package cli

import (
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/presentation/cli/actions"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"github.com/SamyRai/juleson/internal/presentation/cli/dev"
	"github.com/SamyRai/juleson/internal/presentation/cli/github"
	"github.com/SamyRai/juleson/internal/presentation/cli/sessions"
	"github.com/SamyRai/juleson/internal/presentation/views"
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
	Template  *views.TemplateFormatter
	Session   *views.SessionFormatter
	ConfigGen *views.ConfigGenerator
}

// NewApp creates a new CLI application with dependency injection
func NewApp(cfg *config.Config) *App {
	app := &App{
		container: services.NewContainer(cfg),
		formatters: &Formatters{
			Template:  views.NewTemplateFormatter(),
			Session:   views.NewSessionFormatter(),
			ConfigGen: views.NewConfigGenerator(),
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
		Use:     "juleson",
		Short:   "Jules automation CLI tool",
		Long:    "A comprehensive CLI tool for automating project tasks using Google's Jules AI coding agent",
		Version: core.Version,
	}
	a.rootCmd.SetVersionTemplate(core.VersionText())

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

	// Core commands
	a.rootCmd.AddCommand(core.NewSetupCommand())
	a.rootCmd.AddCommand(core.NewSourcesCommand(a.container.Config()))
	a.rootCmd.AddCommand(core.NewActivitiesCommand(a.container.Config()))
	a.rootCmd.AddCommand(core.NewCompletionCommand())
	a.rootCmd.AddCommand(core.NewVersionCommand())
	a.rootCmd.AddCommand(core.NewConfigCommand(a.container.Config()))
	a.rootCmd.AddCommand(core.NewInitCommand(a.formatters.ConfigGen.GenerateProjectConfig))
	a.rootCmd.AddCommand(core.NewTemplateCommand(
		a.container.TemplateManager,
		core.DisplayTemplates,
		core.DisplayTemplateDetails,
	))
	a.rootCmd.AddCommand(core.NewSyncCommand())
	a.rootCmd.AddCommand(core.NewOfficialCommand())

	// Vertical Slices
	a.rootCmd.AddCommand(sessions.NewSessionsCommand(a.container.Config()))
	a.rootCmd.AddCommand(github.NewPRCommand(a.container.Config()))
	a.rootCmd.AddCommand(github.NewGitHubCommand(a.container.Config()))
	a.rootCmd.AddCommand(actions.NewActionsCommand(a.container.Config()))
	a.rootCmd.AddCommand(dev.NewDevCommand())
}
