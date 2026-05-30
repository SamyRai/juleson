package services

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/automation"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/gemini"
	"github.com/SamyRai/juleson/internal/llm"
	"github.com/SamyRai/juleson/internal/orchestration"
	"github.com/SamyRai/juleson/internal/orchestration/adapters"
	"github.com/SamyRai/juleson/internal/templates"
)

// Container manages application dependencies and services
// It follows the Dependency Injection pattern for lazy initialization
type Container struct {
	config               *config.Config
	julesClient          *jules.Client
	llmProvider          llm.Provider
	templateManager      *templates.Manager
	automationEngine     *automation.Engine
	orchestrationRuntime *orchestration.Runtime
	logger               *slog.Logger
	mu                   sync.RWMutex
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

// JulesClient returns the Jules API client (lazy initialization)
func (c *Container) JulesClient() *jules.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.julesClientLocked()
}

// LLMProvider returns the abstract LLM provider based on configuration.
func (c *Container) LLMProvider() llm.Provider {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.llmProviderLocked()
}

// julesClientLocked returns the Jules API client without locking (internal use)
func (c *Container) julesClientLocked() *jules.Client {
	if c.julesClient == nil {
		// Only create client if API key is available
		if c.config.Jules.APIKey == "" {
			return nil // Return nil to indicate client is not available
		}
		c.julesClient = jules.NewClient(c.config.Jules.APIKey, jules.WithBaseURL(c.config.Jules.BaseURL), jules.WithTimeout(c.config.Jules.Timeout), jules.WithRetryAttempts(c.config.Jules.RetryAttempts), jules.WithDebugLog(c.config.Jules.DebugLog), jules.WithLogger(getLogger(c.config.Jules.DebugLog)))

	}

	return c.julesClient
}

// llmProviderLocked returns the abstract LLM provider
func (c *Container) llmProviderLocked() llm.Provider {
	if c.llmProvider == nil {
		var primary llm.Provider
		var secondary llm.Provider

		// Always try to initialize Gemini as it can serve as a primary or fallback
		geminiConfig := gemini.Config{
			APIKey:    c.config.Gemini.APIKey,
			Backend:   c.config.Gemini.Backend,
			Project:   c.config.Gemini.Project,
			Location:  c.config.Gemini.Location,
			Model:     c.config.Gemini.Model,
			Timeout:   c.config.Gemini.Timeout,
			MaxTokens: c.config.Gemini.MaxTokens,
		}

		var geminiProvider *llm.GeminiProvider
		client, err := gemini.NewClient(&geminiConfig)
		if err == nil && c.config.Gemini.APIKey != "" { // verify we actually have credentials
			geminiProvider = llm.NewGeminiProvider(client, c.config.Gemini.Model)
		}

		if c.config.ActiveBackend == "ollama" {
			primary = llm.NewOllamaProvider(c.config.Ollama.BaseURL, c.config.Ollama.Model)
			if geminiProvider != nil {
				secondary = geminiProvider
			}
		} else {
			if geminiProvider != nil {
				primary = geminiProvider
			}
		}

		if primary != nil && secondary != nil {
			c.llmProvider = llm.NewFallbackProvider(primary, secondary, c.logger)
		} else if primary != nil {
			c.llmProvider = primary
		} else if secondary != nil {
			c.llmProvider = secondary
		} else {
			// Fallback placeholder or leave nil
			c.llmProvider = nil
		}
	}
	return c.llmProvider
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

// AnalyzeProject returns analyzer context without initializing the legacy automation engine.
func (c *Container) AnalyzeProject(projectPath string) (*analyzer.ProjectContext, error) {
	return analyzer.NewProjectAnalyzer().Analyze(projectPath)
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

// OrchestrationRuntime returns the extraction-ready orchestration runtime.
func (c *Container) OrchestrationRuntime() (*orchestration.Runtime, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.orchestrationRuntime == nil {
		templateManager, err := c.templateManagerLocked()
		if err != nil {
			return nil, err
		}

		sessionGateway := adapters.NewJulesSessionGateway(c.julesClientLocked())
		sourceMatcher := adapters.NewSourceMatcherAdapter()
		c.orchestrationRuntime = orchestration.NewRuntime(orchestration.RuntimeDeps{
			ProjectAnalyzer: adapters.NewAnalyzerAdapter(nil),
			Planner:         adapters.NewLLMPlanner(c.llmProviderLocked()),
			DecisionMaker:   adapters.NewLLMDecisionMaker(c.llmProviderLocked()),
			TaskExecutor:    adapters.NewJulesTaskExecutor(sessionGateway, sourceMatcher),
			CheckpointStore: adapters.NewJSONCheckpointStore(c.config.Automation.CheckpointPath),
			SessionGateway:  sessionGateway,
			TemplateStore:   adapters.NewTemplateStoreAdapter(templateManager),
			PromptRenderer:  adapters.NewPromptRendererAdapter(),
			SourceMatcher:   sourceMatcher,
			OutputWriter:    adapters.NewMarkdownOutputWriter(),
			ProgressSink:    adapters.NoopProgressSink{},
			Clock:           adapters.SystemClock{},
		})
	}

	return c.orchestrationRuntime, nil
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
