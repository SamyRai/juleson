package core

import (
	"log/slog"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/agent/memory"
	"github.com/SamyRai/juleson/internal/agent/review"
	"github.com/SamyRai/juleson/internal/agent/tools"
	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/llm"
)

// Default agent configuration constants
const (
	DefaultMaxIterations      = 20
	DefaultCheckpointInterval = 5 * time.Minute
	DefaultMinTestCoverage    = 0.8
	// DefaultConfidenceThreshold is the default confidence threshold for initial observations
	DefaultConfidenceThreshold = 0.5
	// PercentageScale is used to convert scores to percentage (0-1 range)
	PercentageScale = 100.0
)

// CoreAgent implements the main agent loop
type CoreAgent struct {
	state        agent.AgentState
	toolRegistry tools.ToolRegistry
	reviewer     review.Reviewer
	memory       memory.Memory
	logger       *slog.Logger
	analyzer     *analyzer.ProjectAnalyzer

	// New production-ready components
	planner       *Planner
	retryStrategy *RetryStrategy
	checkpointMgr *CheckpointManager
	telemetry     *Metrics
	validator     *ConstraintValidator
	llmProvider   llm.Provider
	executor      *taskExecutor

	// Current execution context
	currentGoal    *agent.Goal
	currentPlan    []agent.Task
	decisions      []agent.Decision
	projectContext *analyzer.ProjectContext

	// Configuration
	maxIterations int
	dryRun        bool
}

// Config holds agent configuration
type Config struct {
	MaxIterations   int
	DryRun          bool
	ReviewConfig    *review.Config
	Logger          *slog.Logger
	LLMProvider     llm.Provider
	CheckpointDir   string
	AutoSave        bool
	SaveInterval    time.Duration
	RetryConfig     *RetryStrategy
	EnableTelemetry bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxIterations:   DefaultMaxIterations,
		DryRun:          false,
		ReviewConfig:    review.DefaultConfig(),
		Logger:          slog.Default(),
		CheckpointDir:   "./checkpoints",
		AutoSave:        true,
		SaveInterval:    DefaultCheckpointInterval,
		RetryConfig:     DefaultRetryStrategy(),
		EnableTelemetry: true,
	}
}

// NewAgent creates a new core agent
func NewAgent(toolRegistry tools.ToolRegistry, config *Config) agent.Agent {
	if config == nil {
		config = DefaultConfig()
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.RetryConfig == nil {
		config.RetryConfig = DefaultRetryStrategy()
	}

	agent := &CoreAgent{
		state:         agent.StateIdle,
		toolRegistry:  toolRegistry,
		reviewer:      review.NewReviewer(config.ReviewConfig),
		memory:        memory.NewMemory(),
		logger:        config.Logger,
		analyzer:      analyzer.NewProjectAnalyzer(),
		maxIterations: config.MaxIterations,
		dryRun:        config.DryRun,
		decisions:     make([]agent.Decision, 0),
	}

	// Initialize new components
	if config.LLMProvider != nil {
		agent.llmProvider = config.LLMProvider
		agent.planner = NewPlanner(config.LLMProvider, config.Logger)
	}

	agent.retryStrategy = config.RetryConfig
	agent.checkpointMgr = NewCheckpointManager(config.CheckpointDir, config.AutoSave, config.SaveInterval, config.Logger)

	if config.EnableTelemetry {
		agent.telemetry = NewMetrics()
	}

	// Initialize validator with empty constraints (can be set later)
	agent.validator = NewConstraintValidator([]string{})
	agent.executor = newTaskExecutor(agent.toolRegistry, agent.validator, agent.telemetry, agent.logger, agent.dryRun)

	return agent
}
