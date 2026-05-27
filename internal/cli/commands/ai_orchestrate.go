package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/orchestration"
	"github.com/SamyRai/juleson/internal/orchestration/app"
	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/spf13/cobra"
)

// NewAIOrchest rateCommand creates the AI orchestrate command
func NewAIOrchestCommand(cfg *config.Config, initializeRuntime func() (*orchestration.Runtime, error)) *cobra.Command {
	var (
		sourceID    string
		projectPath string
		constraints []string
		geminiModel string
		geminiKey   string
		maxIters    int
		autoApprove bool
	)

	cmd := &cobra.Command{
		Use:   "ai-orchestrate [goal]",
		Short: "Let AI orchestrate a complex workflow intelligently",
		Long: `Let Gemini AI orchestrate a complex software development workflow.

The AI will:
- Analyze your project deeply
- Create an adaptive execution plan
- Execute tasks one by one
- Make decisions about what to do next
- Adapt the plan based on progress
- Determine when the goal is achieved

Unlike predefined workflows, the AI decides the steps dynamically based on:
- Your project's specific needs
- Results from previous tasks
- Emerging insights and challenges
- Your feedback and constraints

Examples:
  # Let AI modernize your API
  juleson ai-orchestrate "Modernize our REST API to GraphQL" \\
    --source my-repo \\
    --path ./services/api

  # AI-driven refactoring with constraints
  juleson ai-orchestrate "Improve code quality and test coverage" \\
    --source my-repo \\
    --constraint "Don't change public APIs" \\
    --constraint "Maintain backward compatibility"

  # Complex migration
  juleson ai-orchestrate "Migrate from MongoDB to PostgreSQL" \\
    --source backend \\
    --gemini-model gemini-2.0-flash-exp \\
    --max-iterations 30`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			goal := args[0]

			if geminiKey == "" {
				geminiKey = os.Getenv("GEMINI_API_KEY")
			}
			if geminiKey == "" {
				return fmt.Errorf("Gemini API key required. Set --gemini-key or GEMINI_API_KEY environment variable")
			}
			cfg.Gemini.APIKey = geminiKey
			cfg.Gemini.Backend = "gemini-api"
			cfg.Gemini.Model = geminiModel
			cfg.Gemini.Timeout = 30 * time.Second
			cfg.Jules.Timeout = 30 * time.Second
			cfg.Jules.RetryAttempts = 3

			runtime, err := initializeRuntime()
			if err != nil {
				return fmt.Errorf("failed to create orchestration runtime: %w", err)
			}

			// Setup context with cancellation
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle Ctrl+C gracefully
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Println("\n\n⚠️  Stopping AI orchestration...")
				cancel()
			}()

			// Print header
			fmt.Println("╔══════════════════════════════════════════════════════════════╗")
			fmt.Println("║           AI-Powered Workflow Orchestration                  ║")
			fmt.Println("╚══════════════════════════════════════════════════════════════╝")
			fmt.Printf("\n🎯 Goal: %s\n", goal)
			fmt.Printf("📁 Project: %s\n", projectPath)
			if len(constraints) > 0 {
				fmt.Println("⚠️  Constraints:")
				for _, c := range constraints {
					fmt.Printf("   - %s\n", c)
				}
			}
			fmt.Println("\n🤖 AI is analyzing your project and creating a plan...")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()

			runOptions := aiWorkflowOptionsFromFlags(maxIters, autoApprove)
			fmt.Printf("Approval: %s\n", approvalModeText(runOptions.ApprovalPolicy))

			result, err := runtime.AIWorkflowRunner().RunWithOptions(ctx, domain.Goal{
				ID:          fmt.Sprintf("ai-goal-%d", time.Now().Unix()),
				Description: goal,
				Constraints: constraints,
				Context: domain.GoalContext{
					SourceID:    sourceID,
					ProjectPath: projectPath,
				},
				Priority: domain.PriorityMedium,
			}, runOptions)
			if err != nil {
				fmt.Printf("\n❌ AI orchestration failed: %v\n", err)
				return err
			}

			// Print summary
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("✅ AI Orchestration Complete!")
			fmt.Printf("📝 Tasks completed: %d\n", len(result.Tasks))
			for _, task := range result.Tasks {
				if task.SessionID != "" {
					fmt.Printf("📊 Session ID: %s\n", task.SessionID)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sourceID, "source", "", "Source ID to use for the session (required)")
	cmd.Flags().StringVar(&projectPath, "path", ".", "Project path to analyze")
	cmd.Flags().StringSliceVar(&constraints, "constraint", []string{}, "Constraints for AI to follow (can be specified multiple times)")
	cmd.Flags().StringVar(&geminiModel, "gemini-model", "gemini-2.0-flash-exp", "Gemini model to use")
	cmd.Flags().StringVar(&geminiKey, "gemini-key", "", "Gemini API key (or use GEMINI_API_KEY env var)")
	cmd.Flags().IntVar(&maxIters, "max-iterations", 20, "Maximum number of AI decision iterations")
	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Automatically approve AI plans (use with caution)")

	mustMarkFlagRequired(cmd, "source")

	return cmd
}

func aiWorkflowOptionsFromFlags(maxIterations int, autoApprove bool) app.AIWorkflowRunOptions {
	return app.AIWorkflowRunOptions{
		MaxIterations: maxIterations,
		ApprovalPolicy: domain.ApprovalPolicy{
			RequirePlanApproval: !autoApprove,
			AutoApprove:         autoApprove,
		},
	}
}

func approvalModeText(policy domain.ApprovalPolicy) string {
	if policy.AutoApprove {
		return "auto-approve enabled"
	}
	return "plan approval required"
}
