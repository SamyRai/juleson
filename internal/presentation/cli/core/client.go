package core

import (
	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/logger"
)

func NewJulesClient(cfg *config.Config) *jules.Client {
	return jules.NewClient(
		cfg.Jules.APIKey,
		jules.WithBaseURL(cfg.Jules.BaseURL),
		jules.WithTimeout(cfg.Jules.Timeout),
		jules.WithRetryAttempts(cfg.Jules.RetryAttempts),
		jules.WithDebugLog(cfg.Jules.DebugLog),
		jules.WithLogger(logger.New(logger.Config{Debug: cfg.Jules.DebugLog})),
	)
}
