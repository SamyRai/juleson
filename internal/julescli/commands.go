package julescli

import (
	"github.com/SamyRai/juleson/internal/cli/commands"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

// NewCoreCommands returns the Jules-focused CLI commands.
func NewCoreCommands(cfg *config.Config) []*cobra.Command {
	return []*cobra.Command{
		commands.NewSetupCommand(),
		commands.NewSessionsCommand(cfg),
		commands.NewSourcesCommand(cfg),
		commands.NewActivitiesCommand(cfg),
		commands.NewCompletionCommand(),
		commands.NewVersionCommand(),
	}
}
