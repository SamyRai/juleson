package services

import (
	"fmt"
	"github.com/SamyRai/juleson/internal/logger"
	"log/slog"
	"sync"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/templates"
)

// Container manages application dependencies and services
// It follows the Dependency Injection pattern for lazy initialization
type Container struct {
	config          *config.Config
	julesClient     *jules.Client
	templateManager *templates.Manager
	logger          *slog.Logger
	mu              sync.RWMutex
}

// NewContainer creates a new service container
// Event coordination should be initialized separately by the application
func NewContainer(cfg *config.Config) *Container {
	logger := slog.Default()

	container := &Container{
		config: cfg,
		logger: logger,
	}

	return container
}

func (c *Container) JulesClient() *jules.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.julesClientLocked()
}

// julesClientLocked returns the Jules API client without locking (internal use)
func (c *Container) julesClientLocked() *jules.Client {
	if c.julesClient == nil {
		// Only create client if API key is available
		if c.config.Jules.APIKey == "" {
			return nil // Return nil to indicate client is not available
		}
		c.julesClient = jules.NewClient(c.config.Jules.APIKey, jules.WithBaseURL(c.config.Jules.BaseURL), jules.WithTimeout(c.config.Jules.Timeout), jules.WithRetryAttempts(c.config.Jules.RetryAttempts), jules.WithDebugLog(c.config.Jules.DebugLog), jules.WithLogger(logger.New(logger.Config{Debug: c.config.Jules.DebugLog})))

	}

	return c.julesClient
}

// TemplateManager returns the template manager (lazy initialization)
func (c *Container) TemplateManager() (*templates.Manager, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.templateManagerLocked()
}

// templateManagerLocked returns the template manager without locking (internal use)
func (c *Container) templateManagerLocked() (*templates.Manager, error) {
	if c.templateManager == nil {
		manager, err := templates.NewManager(
			c.config.Templates.BuiltinPath,
			c.config.Templates.CustomPath,
			c.config.Templates.EnableCustom,
		)
		if err != nil {
			// Templates might not be available, return a graceful error
			return nil, fmt.Errorf("template manager initialization failed (templates directory may not be accessible): %w", err)
		}
		c.templateManager = manager
	}

	return c.templateManager, nil
}

// Config returns the application configuration
func (c *Container) Config() *config.Config {
	return c.config
}

// Close cleans up any resources held by the container
func (c *Container) Close() error {
	// No resources to clean up currently
	return nil
}
