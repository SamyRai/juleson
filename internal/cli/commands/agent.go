package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/services"
	"github.com/spf13/cobra"
)

// NewAgentCommand creates the agent command
func NewAgentCommand(cfg *config.Config) *cobra.Command {
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

	cmd.AddCommand(newAgentExecuteCommand(cfg))
	cmd.AddCommand(newAgentStatusCommand(cfg))

	return cmd
}

// newAgentExecuteCommand creates the execute subcommand
func newAgentExecuteCommand(cfg *config.Config) *cobra.Command {
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

			container := services.NewContainer(cfg)
			runtime, err := container.OrchestrationRuntime()
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
				fmt.Printf("\nDry run stops before orchestration side effects.\n\n")
				return nil
			}
			fmt.Printf("\n")

			ctx := context.Background()
			result, err := runtime.AgentRunner().Run(ctx, goal)

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

	cmd.MarkFlagRequired("source")

	return cmd
}

// newAgentStatusCommand creates the status subcommand
func newAgentStatusCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show agent status and capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			separator := strings.Repeat("=", 60)
			fmt.Printf("\n🤖 AI Agent Status\n")
			fmt.Printf("%s\n\n", separator)

			fmt.Printf("✅ Agent: Available\n")
			fmt.Printf("✅ Tool Registry: Implemented\n")
			fmt.Printf("✅ Code Review: Implemented\n")
			fmt.Printf("✅ Memory System: Implemented\n")
			fmt.Printf("✅ Jules Integration: Available\n")

			fmt.Printf("\nCapabilities:\n")
			fmt.Printf("  • Perceive - Context gathering and analysis\n")
			fmt.Printf("  • Plan - Multi-step execution planning\n")
			fmt.Printf("  • Act - Tool selection and execution\n")
			fmt.Printf("  • Review - Intelligent code review\n")
			fmt.Printf("  • Reflect - Learning and adaptation\n")

			fmt.Printf("\nAvailable Tools:\n")
			fmt.Printf("  • Jules - AI-powered code generation and modification\n")
			fmt.Printf("  • Analyzer - Project structure and code quality analysis\n")
			fmt.Printf("  • Docker - Container and image management\n")

			fmt.Printf("\nConfiguration:\n")
			fmt.Printf("  Jules API: %s\n", cfg.Jules.BaseURL)

			fmt.Printf("\n")
			return nil
		},
	}

	return cmd
}
