package dev

import (
	"github.com/SamyRai/juleson/pkg/builder"
)

// CommandHandler encapsulates the dependencies for dev commands.
type CommandHandler struct {
	svc *builder.Service
}

// NewCommandHandler creates a new handler.
func NewCommandHandler(svc *builder.Service) *CommandHandler {
	return &CommandHandler{svc: svc}
}
