package commands

import (
	"log/slog"
	"os"
)

func getLogger(debugLog bool) *slog.Logger {
	level := slog.LevelInfo
	if debugLog {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}
