package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SamyRai/juleson/internal/automation"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/gemini"
	"github.com/SamyRai/juleson/pkg/jules"
	"github.com/spf13/cobra"
)

// NewAIOrchest rateCommand creates the AI orchestrate command
func NewAIOrchestCommand(cfg *config.Config) *cobra.Command {
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

			// Setup Jules client
			julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(30*time.Second), jules.WithRetryAttempts(3))

			// Setup Gemini client
			geminiConfig := &gemini.Config{
				APIKey:  geminiKey,
				Backend: "gemini-api",
				Model:   geminiModel,
				Timeout: 30 * time.Second,
			}
			if geminiKey == "" {
				geminiConfig.APIKey = os.Getenv("GEMINI_API_KEY")
			}
			if geminiConfig.APIKey == "" {
				return fmt.Errorf("Gemini API key required. Set --gemini-key or GEMINI_API_KEY environment variable")
			}

			geminiClient, err := gemini.NewClient(geminiConfig)
			if err != nil {
				return fmt.Errorf("failed to create Gemini client: %w", err)
			}
			defer geminiClient.Close()

			// Create AI orchestrator
			orchestratorConfig := &automation.AIOrchestrationConfig{
				MaxIterations: maxIters,
				CheckInterval: 15 * time.Second,
				AllowedTools: []string{
					"execute_template",
					"run_tests",
					"apply_patches",
					"create_issue",
				},
				AutoApprove:    autoApprove,
				MaxSessionTime: 4 * time.Hour,
			}

			orchestrator := automation.NewAIOrchestrator(
				julesClient,
				geminiClient,
				orchestratorConfig,
			)

			// Setup context with cancellation
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle Ctrl+C gracefully
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Println("\n\n⚠️  Stopping AI orchestration...")
				orchestrator.Stop()
				cancel()
			}()

			// Monitor progress in a goroutine
			go func() {
				for progress := range orchestrator.ProgressChannel() {
					fmt.Printf("\n🤖 AI: %s\n", progress.Message)
					if len(progress.NextSteps) > 0 {
						fmt.Println("   Next steps:")
						for _, step := range progress.NextSteps {
							fmt.Printf("   - %s\n", step)
						}
					}
					fmt.Printf("   Progress: %.0f%% | Phase: %s\n", progress.Progress, progress.Phase)
				}
			}()

			// Monitor AI decisions
			go func() {
				for decision := range orchestrator.DecisionChannel() {
					fmt.Printf("\n🧠 AI Decision: %s\n", decision.DecisionType)
					fmt.Printf("   Reasoning: %s\n", decision.Reasoning)
					fmt.Printf("   Confidence: %.0f%%\n", decision.Confidence*100)
					if decision.Action != "" {
						fmt.Printf("   Action: %s\n", decision.Action)
					}
				}
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
			fmt.Println() // Execute AI orchestration
			err = orchestrator.Execute(ctx, sourceID, goal, projectPath, constraints)
			if err != nil {
				fmt.Printf("\n❌ AI orchestration failed: %v\n", err)
				return err
			}

			// Print summary
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("✅ AI Orchestration Complete!")
			fmt.Printf("📊 Session ID: %s\n", orchestrator.GetSessionID())
			fmt.Printf("📝 AI made %d decisions during execution\n", len(orchestrator.GetDecisionHistory()))
			fmt.Println("\nTo view full details:")
			fmt.Printf("  juleson sessions show %s\n", orchestrator.GetSessionID())

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

	cmd.MarkFlagRequired("source")

	return cmd
}
