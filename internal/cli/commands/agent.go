package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/orchestration"
	"github.com/SamyRai/juleson/internal/orchestration/app"
	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/spf13/cobra"
)

// NewAgentCommand creates the agent command
func NewAgentCommand(cfg *config.Config, initializeRuntime func() (*orchestration.Runtime, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Run the AI agent for automated development tasks",
		Long: `The agent command provides an intelligent AI agent that orchestrates
development tasks, performs code reviews, and learns from experience.

The agent implements a sophisticated loop:
  1. Perceive - Understand the goal and gather context
  2. Plan - Generate multi-step execution plan
  3. Act - Execute tasks using appropriate tools (Jules, GitHub, etc.)
  4. Review - Intelligent code review before approval
  5. Reflect - Learn from outcomes and adapt

Example:
  juleson agent execute "improve test coverage to 80%" --source my-repo
  juleson agent execute "fix security vulnerabilities" --priority CRITICAL
  juleson agent execute "modernize API to OpenAPI 3.0" --dry-run`,
	}

	cmd.AddCommand(newAgentExecuteCommand(initializeRuntime))
	cmd.AddCommand(newAgentStatusCommand(cfg, initializeRuntime))

	return cmd
}

// newAgentExecuteCommand creates the execute subcommand
func newAgentExecuteCommand(initializeRuntime func() (*orchestration.Runtime, error)) *cobra.Command {
	var (
		sourceID    string
		priority    string
		constraints []string
		dryRun      bool
		strictness  string
		maxIters    int
	)

	cmd := &cobra.Command{
		Use:   "execute [goal]",
		Short: "Execute a goal using the AI agent",
		Long: `Execute a goal using the AI agent.

The agent will:
- Analyze the goal and codebase
- Create an execution plan
- Execute tasks using appropriate tools
- Review all changes before applying
- Learn from the experience

Examples:
  juleson agent execute "improve test coverage" --source my-repo
  juleson agent execute "fix security issues" --priority CRITICAL --strictness high
  juleson agent execute "refactor auth module" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			goalDescription := args[0]
			normalizedStrictness, err := validateReviewStrictness(strictness)
			if err != nil {
				return err
			}

			// Validate priority
			var goalPriority domain.Priority
			switch priority {
			case "CRITICAL", "critical":
				goalPriority = domain.PriorityCritical
			case "HIGH", "high":
				goalPriority = domain.PriorityHigh
			case "MEDIUM", "medium":
				goalPriority = domain.PriorityMedium
			case "LOW", "low":
				goalPriority = domain.PriorityLow
			default:
				goalPriority = domain.PriorityMedium
			}

			runtime, err := initializeRuntime()
			if err != nil {
				return fmt.Errorf("failed to create orchestration runtime: %w", err)
			}

			// Create goal
			goal := domain.Goal{
				ID:          fmt.Sprintf("goal-%d", time.Now().Unix()),
				Description: goalDescription,
				Constraints: constraints,
				Priority:    goalPriority,
				Context: domain.GoalContext{
					SourceID: sourceID,
				},
			}

			// Execute
			fmt.Printf("\n🤖 Starting AI Agent\n")
			fmt.Printf("Goal: %s\n", goalDescription)
			fmt.Printf("Priority: %s\n", priority)
			if dryRun {
				fmt.Printf("Mode: DRY RUN (no changes will be applied)\n")
			}
			fmt.Printf("\n")

			ctx := context.Background()
			result, err := runtime.AgentRunner().RunWithOptions(ctx, goal, app.AgentRunOptions{
				DryRun:           dryRun,
				MaxIterations:    maxIters,
				ReviewStrictness: normalizedStrictness,
			})

			// Display results
			separator := strings.Repeat("=", 60)
			fmt.Printf("\n%s\n", separator)
			fmt.Printf("📊 Execution Results\n")
			fmt.Printf("%s\n\n", separator)

			if err != nil {
				fmt.Printf("❌ Status: FAILED\n")
				fmt.Printf("Error: %v\n", err)
				return err
			}

			if result.Success {
				fmt.Printf("✅ Status: SUCCESS\n")
			} else {
				fmt.Printf("⚠️  Status: INCOMPLETE\n")
			}

			fmt.Printf("Duration: %s\n", result.Duration)
			fmt.Printf("Final State: %s\n", result.State)
			fmt.Printf("Tasks Completed: %d\n", len(result.Tasks))
			if result.Plan != nil {
				fmt.Printf("Tasks Planned: %d\n", len(result.Plan.Tasks))
			}

			if len(result.Tasks) > 0 {
				fmt.Printf("\nTasks:\n")
				for i, task := range result.Tasks {
					status := "✅"
					if !task.Success {
						status = "❌"
					}
					fmt.Printf("  %s %d. %s (%s)\n", status, i+1, task.TaskName, task.Tool)
					if task.Review != nil {
						fmt.Printf("     Review: %.1f/100\n", task.Review.Score)
						if len(task.Review.Diagnostics) > 0 {
							fmt.Printf("     Comments: %d\n", len(task.Review.Diagnostics))
						}
					}
				}
			}
			if dryRun && result.Plan != nil && len(result.Plan.Tasks) > 0 {
				fmt.Printf("\nPlanned Tasks:\n")
				for i, task := range result.Plan.Tasks {
					fmt.Printf("  %d. %s", i+1, task.Name)
					if task.Tool != "" {
						fmt.Printf(" (%s)", task.Tool)
					}
					fmt.Printf("\n")
				}
			}

			if len(result.Learnings) > 0 {
				fmt.Printf("\n📚 Learnings:\n")
				for _, learning := range result.Learnings {
					fmt.Printf("  • %s\n", learning)
				}
			}

			if result.Summary != "" {
				fmt.Printf("\n%s\n", result.Summary)
			}

			fmt.Printf("\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&sourceID, "source", "", "Source context ID (required)")
	cmd.Flags().StringVar(&priority, "priority", "MEDIUM", "Goal priority (CRITICAL, HIGH, MEDIUM, LOW)")
	cmd.Flags().StringSliceVar(&constraints, "constraint", []string{}, "Constraints for execution")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run in dry-run mode (no changes applied)")
	cmd.Flags().StringVar(&strictness, "strictness", "medium", "Code review strictness (low, medium, high)")
	cmd.Flags().IntVar(&maxIters, "max-iterations", 20, "Maximum number of iterations")

	mustMarkFlagRequired(cmd, "source")

	return cmd
}

// newAgentStatusCommand creates the status subcommand
func newAgentStatusCommand(cfg *config.Config, initializeRuntime func() (*orchestration.Runtime, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show agent status and capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := initializeRuntime()
			if err != nil {
				return fmt.Errorf("failed to inspect orchestration runtime: %w", err)
			}
			capabilities := runtime.Capabilities()
			separator := strings.Repeat("=", 60)
			fmt.Printf("\n🤖 AI Agent Status\n")
			fmt.Printf("%s\n\n", separator)

			fmt.Printf("%s Agent runtime\n", statusMarker(capabilities.Planning && capabilities.TaskExecution))
			fmt.Printf("%s Jules API: %s\n", statusMarker(cfg.Jules.APIKey != ""), configuredText(cfg.Jules.APIKey != ""))
			fmt.Printf("%s Gemini planning: %s\n", statusMarker(cfg.Gemini.APIKey != ""), configuredText(cfg.Gemini.APIKey != ""))
			fmt.Printf("%s Reviewer: %s\n", statusMarker(capabilities.Review), configuredText(capabilities.Review))
			fmt.Printf("%s Memory: %s\n", statusMarker(capabilities.Memory), configuredText(capabilities.Memory))
			fmt.Printf("%s Checkpointing: %s\n", statusMarker(capabilities.Checkpointing), configuredText(capabilities.Checkpointing))
			fmt.Printf("%s Dry-run planning: %s\n", statusMarker(capabilities.DryRunPlanning), configuredText(capabilities.DryRunPlanning))

			fmt.Printf("\nCapabilities:\n")
			fmt.Printf("  • Project analysis: %s\n", configuredText(capabilities.ProjectAnalysis))
			fmt.Printf("  • Plan generation: %s\n", configuredText(capabilities.Planning))
			fmt.Printf("  • Jules task execution: %s\n", configuredText(capabilities.TaskExecution && cfg.Jules.APIKey != ""))
			fmt.Printf("  • Review adapter: %s\n", configuredText(capabilities.Review))
			fmt.Printf("  • Memory adapter: %s\n", configuredText(capabilities.Memory))

			fmt.Printf("\nAvailable Tools:\n")
			fmt.Printf("  • Jules sessions: %s\n", configuredText(cfg.Jules.APIKey != ""))
			fmt.Printf("  • Analyzer: %s\n", configuredText(capabilities.ProjectAnalysis))

			fmt.Printf("\nConfiguration:\n")
			fmt.Printf("  Jules API: %s\n", cfg.Jules.BaseURL)

			fmt.Printf("\n")
			return nil
		},
	}

	return cmd
}

func validateReviewStrictness(value string) (string, error) {
	strictness := strings.ToLower(strings.TrimSpace(value))
	switch strictness {
	case "", "low", "medium", "high":
		if strictness == "" {
			return "medium", nil
		}
		return strictness, nil
	default:
		return "", fmt.Errorf("invalid --strictness %q: expected low, medium, or high", value)
	}
}

func statusMarker(available bool) string {
	if available {
		return "✅"
	}
	return "⚠️ "
}

func configuredText(available bool) string {
	if available {
		return "configured"
	}
	return "not configured"
}
