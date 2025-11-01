package services

import (
	"fmt"
	"sync"

	"github.com/SamyRai/juleson/internal/automation"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/jules"
	"github.com/SamyRai/juleson/internal/templates"
)

// Container manages application dependencies and services
type Container struct {
	config           *config.Config
	julesClient      *jules.Client
	templateManager  *templates.Manager
	automationEngine *automation.Engine
	mu               sync.RWMutex
}

// NewContainer creates a new service container
func NewContainer(cfg *config.Config) *Container {
	return &Container{
		config: cfg,
	}
}

// JulesClient returns the Jules API client (lazy initialization)
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
		c.julesClient = jules.NewClient(
			c.config.Jules.APIKey,
			c.config.Jules.BaseURL,
			c.config.Jules.Timeout,
			c.config.Jules.RetryAttempts,
		)
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
			// In MCP context, templates might not be available, return a graceful error
			return nil, fmt.Errorf("template manager initialization failed (templates directory may not be accessible): %w", err)
		}
		c.templateManager = manager
	}

	return c.templateManager, nil
}

// AutomationEngine returns the automation engine (lazy initialization)
func (c *Container) AutomationEngine() (*automation.Engine, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.automationEngine == nil {
		// Use internal non-locking methods to avoid deadlock
		templateManager, err := c.templateManagerLocked()
		if err != nil {
			return nil, err
		}

		c.automationEngine = automation.NewEngine(c.julesClientLocked(), templateManager)
	}

	return c.automationEngine, nil
}

// Config returns the application configuration
func (c *Container) Config() *config.Config {
	return c.config
}

// Close cleans up any resources held by the container
func (c *Container) Close() error {
	// Add cleanup logic if needed in the future
	// For example: closing database connections, flushing logs, etc.
	return nil
}
