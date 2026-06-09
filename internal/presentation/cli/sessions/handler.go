package sessions

import (
	"github.com/SamyRai/juleson/internal/config"
)

// CommandHandler provides commands for managing Jules sessions.
type CommandHandler struct {
	cfg *config.Config
}

// NewCommandHandler creates a new command handler with the provided configuration.
func NewCommandHandler(cfg *config.Config) *CommandHandler {
	return &CommandHandler{
		cfg: cfg,
	}
}
