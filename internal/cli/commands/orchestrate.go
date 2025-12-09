package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/automation"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/jules"

	"github.com/spf13/cobra"
)

// NewOrchestrateCommand creates the orchestrate command for multi-task workflows
func NewOrchestrateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orchestrate",
		Short: "Orchestrate complex multi-task workflows in a single session",
		Long: `Execute complex multi-phase workflows efficiently within a single Jules session.

The orchestrator manages the entire workflow lifecycle:
- Creates a session with comprehensive initial prompt
- Monitors progress in real-time
- Dynamically adds tasks based on progress
- Handles plan approval gates
- Tracks execution and provides detailed reporting

This approach is much more efficient than creating separate sessions for each task.`,
	}

	// Add workflow presets
	cmd.AddCommand(newOrchestrateAPICommand(cfg))
	cmd.AddCommand(newOrchestrateMicroservicesCommand(cfg))
	cmd.AddCommand(newOrchestrateCustomCommand(cfg))

	return cmd
}

// newOrchestrateAPICommand creates command for API modernization workflow
func newOrchestrateAPICommand(cfg *config.Config) *cobra.Command {
	var sourceID string
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "api-modernization",
		Short: "Modernize API with authentication, testing, and CI/CD",
		Long: `Execute a comprehensive API modernization workflow:

Phase 1: Foundation (30 min)
  - Analyze current API structure
  - Create OpenAPI 3.0 specification
  - Plan database migrations

Phase 2: Core Implementation (1 hour)
  - Implement JWT authentication
  - Refactor API endpoints
  - Add request validation

Phase 3: Quality Assurance (45 min)
  - Add comprehensive tests
  - Generate API documentation

Phase 4: DevOps (30 min)
  - Set up CI/CD pipeline
  - Add monitoring and logging

Total estimated duration: 2h 45min in a single session!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPIModernizationWorkflow(cfg, sourceID, autoApprove)
		},
	}

	cmd.Flags().StringVarP(&sourceID, "source", "s", "", "Source ID (GitHub repo)")
	cmd.Flags().BoolVarP(&autoApprove, "auto-approve", "a", false, "Automatically approve plans")
	cmd.MarkFlagRequired("source")

	return cmd
}

// newOrchestrateMicroservicesCommand creates command for microservices migration
func newOrchestrateMicroservicesCommand(cfg *config.Config) *cobra.Command {
	var sourceID string
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "microservices-migration",
		Short: "Migrate monolith to microservices architecture",
		Long: `Execute a microservices migration workflow:

Phase 1: Analysis (45 min)
  - Analyze current architecture
  - Identify bounded contexts
  - Create decomposition plan

Phase 2: Service Extraction (2 hours)
  - Extract first microservice
  - Implement service communication
  - Set up API gateway

Phase 3: Data Migration (1 hour)
  - Split databases
  - Implement data migration
  - Add consistency handling

Total estimated duration: 3h 45min in a single session!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMicroservicesMigrationWorkflow(cfg, sourceID, autoApprove)
		},
	}

	cmd.Flags().StringVarP(&sourceID, "source", "s", "", "Source ID (GitHub repo)")
	cmd.Flags().BoolVarP(&autoApprove, "auto-approve", "a", false, "Automatically approve plans")
	cmd.MarkFlagRequired("source")

	return cmd
}

// newOrchestrateCustomCommand creates command for custom workflows
func newOrchestrateCustomCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom [workflow-file]",
		Short: "Execute a custom workflow from YAML file",
		Long: `Execute a custom workflow defined in a YAML file.

The workflow file should define phases and tasks:

  name: My Custom Workflow
  description: Custom automation workflow
  phases:
    - name: Phase 1
      description: First phase
      tasks:
        - name: Task 1
          prompt: "Do something"
          wait_for_plan: true
          auto_approve: false`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("custom workflows not yet implemented - coming soon")
		},
	}

	return cmd
}

func runAPIModernizationWorkflow(cfg *config.Config, sourceID string, autoApprove bool) error {
	ctx := context.Background()

	// Initialize Jules client
	julesClient := jules.NewClient(
		cfg.Jules.APIKey,
		cfg.Jules.BaseURL,
		cfg.Jules.Timeout,
		cfg.Jules.RetryAttempts,
	)

	// Define API Modernization workflow
	workflow := &automation.WorkflowDefinition{
		Name:        "API Modernization",
		Description: "Modernize legacy API with authentication, testing, and CI/CD",
		MaxDuration: 4 * time.Hour,
		Phases: []automation.Phase{
			{
				Name:        "Foundation",
				Description: "Set up project structure and dependencies",
				Timeout:     45 * time.Minute,
				Tasks: []automation.Task{
					{
						Name:        "Project Analysis & Planning",
						Description: "Analyze API and create modernization plan",
						Prompt: `Analyze the current API structure and create a comprehensive modernization plan:

1. Review existing API endpoints and data models
2. Create OpenAPI 3.0 specification
3. Plan database schema updates and migrations
4. Identify breaking changes and create migration strategy
5. List all required dependencies and updates

Provide a detailed plan with estimated effort for each component.`,
						WaitForPlan: true,
						AutoApprove: autoApprove,
					},
				},
			},
			{
				Name:        "Core Implementation",
				Description: "Implement authentication and refactor API",
				Timeout:     90 * time.Minute,
				Tasks: []automation.Task{
					{
						Name:        "JWT Authentication",
						Description: "Implement JWT-based authentication system",
						Prompt: `Implement a complete JWT authentication system:

1. Set up JWT token generation and validation
2. Implement refresh token mechanism
3. Add role-based access control (RBAC)
4. Create middleware for route protection
5. Add secure session management
6. Implement logout and token revocation

Ensure all security best practices are followed.`,
						WaitForPlan: false,
						AutoApprove: true,
					},
					{
						Name:        "API Endpoint Refactoring",
						Description: "Refactor existing endpoints",
						Prompt: `Refactor all API endpoints to use the new authentication system:

1. Update endpoints to use auth middleware
2. Add comprehensive request validation
3. Implement proper error handling with meaningful messages
4. Add request/response logging
5. Update endpoint documentation
6. Ensure backward compatibility where possible`,
						WaitForPlan: false,
						AutoApprove: true,
					},
				},
			},
			{
				Name:        "Quality Assurance",
				Description: "Add testing and documentation",
				Timeout:     60 * time.Minute,
				Tasks: []automation.Task{
					{
						Name:        "Comprehensive Testing",
						Description: "Add unit, integration, and e2e tests",
						Prompt: `Create a comprehensive test suite:

1. Unit tests for authentication logic (token generation, validation, RBAC)
2. Integration tests for all API endpoints
3. End-to-end tests for critical user flows
4. Test data factories and fixtures
5. Mock services and external dependencies
6. Achieve >80% code coverage

Ensure tests are maintainable and well-documented.`,
						WaitForPlan: false,
						AutoApprove: true,
					},
					{
						Name:        "API Documentation",
						Description: "Generate comprehensive documentation",
						Prompt: `Generate complete API documentation:

1. Generate API docs from OpenAPI specification
2. Add usage examples for each endpoint
3. Document authentication flows
4. Create migration guide from old API
5. Add troubleshooting section
6. Include code samples in multiple languages`,
						WaitForPlan: false,
						AutoApprove: true,
					},
				},
			},
			{
				Name:        "DevOps Setup",
				Description: "Configure CI/CD and monitoring",
				Timeout:     45 * time.Minute,
				Tasks: []automation.Task{
					{
						Name:        "CI/CD Pipeline",
						Description: "Set up automated testing and deployment",
						Prompt: `Create a complete CI/CD pipeline using GitHub Actions:

1. Set up automated testing on pull requests
2. Add code quality checks (linting, formatting)
3. Implement security scanning (SAST, dependency check)
4. Configure automated deployment to staging
5. Add manual approval gate for production
6. Implement rollback mechanism

Ensure the pipeline is fast and reliable.`,
						WaitForPlan: false,
						AutoApprove: true,
					},
					{
						Name:        "Monitoring & Observability",
						Description: "Add logging, metrics, and health checks",
						Prompt: `Implement comprehensive monitoring and observability:

1. Set up structured logging with context
2. Add metrics collection (request count, latency, errors)
3. Implement health check endpoints
4. Add performance monitoring
5. Set up error tracking and alerting
6. Create dashboards for key metrics

Make the system observable and debuggable.`,
						WaitForPlan: false,
						AutoApprove: true,
					},
				},
			},
		},
		OnPhaseComplete: func(phaseIndex int, result automation.PhaseResult) error {
			fmt.Printf("\n✅ Phase %d completed: %s\n", phaseIndex+1, result.PhaseName)
			fmt.Printf("   Duration: %v\n", result.Duration)
			fmt.Printf("   Tasks: %d\n", len(result.Tasks))
			return nil
		},
		OnWorkflowComplete: func(result automation.WorkflowResult) error {
			fmt.Printf("\n🎉 Workflow completed successfully!\n")
			fmt.Printf("   Total duration: %v\n", result.TotalDuration)
			fmt.Printf("   Phases: %d\n", result.TotalPhases)
			fmt.Printf("   Session: %s\n", result.SessionID)
			return nil
		},
	}

	// Create orchestrator
	orchestratorConfig := &automation.OrchestratorConfig{
		AutoApprove:     autoApprove,
		CheckInterval:   10 * time.Second,
		MaxSessionAge:   4 * time.Hour,
		RetryAttempts:   3,
		ContinueOnError: false,
		SaveState:       true,
	}

	orchestrator := automation.NewSessionOrchestrator(julesClient, workflow, orchestratorConfig)

	// Start monitoring progress in a goroutine
	go monitorProgress(orchestrator)

	// Start workflow
	fmt.Println("🚀 Starting API Modernization Workflow")
	fmt.Println("======================================")
	fmt.Printf("Source: %s\n", sourceID)
	fmt.Printf("Auto-approve: %v\n\n", autoApprove)

	err := orchestrator.Start(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("workflow failed: %w", err)
	}

	// Display execution summary
	displayExecutionSummary(orchestrator)

	return nil
}

func runMicroservicesMigrationWorkflow(cfg *config.Config, sourceID string, autoApprove bool) error {
	ctx := context.Background()

	// Initialize Jules client
	julesClient := jules.NewClient(
		cfg.Jules.APIKey,
		cfg.Jules.BaseURL,
		cfg.Jules.Timeout,
		cfg.Jules.RetryAttempts,
	)

	// Define Microservices Migration workflow
	workflow := &automation.WorkflowDefinition{
		Name:        "Microservices Migration",
		Description: "Migrate monolith to microservices architecture",
		MaxDuration: 5 * time.Hour,
		Phases: []automation.Phase{
			{
				Name:        "Architecture Analysis",
				Description: "Analyze current architecture and plan decomposition",
				Timeout:     60 * time.Minute,
				Tasks: []automation.Task{
					{
						Name:        "Architecture Analysis",
						Description: "Analyze monolith and create migration plan",
						Prompt: `Perform comprehensive architecture analysis and create microservices decomposition plan:

1. Analyze current monolith architecture and dependencies
2. Identify bounded contexts using Domain-Driven Design
3. Map data dependencies between domains
4. Create service decomposition plan with clear boundaries
5. Define communication patterns (sync/async, protocols)
6. Identify shared services and infrastructure needs
7. Create phased migration strategy with minimal risk

Provide detailed architecture diagrams and migration roadmap.`,
						WaitForPlan: true,
						AutoApprove: autoApprove,
					},
				},
			},
			{
				Name:        "First Service Extraction",
				Description: "Extract user service as proof of concept",
				Timeout:     150 * time.Minute,
				Tasks: []automation.Task{
					{
						Name:        "User Service Extraction",
						Description: "Extract user management into microservice",
						Prompt: `Extract user management into a separate microservice:

1. Create new service with its own codebase and structure
2. Implement user CRUD operations
3. Set up dedicated database for user service
4. Create service API (REST/gRPC)
5. Implement authentication and authorization
6. Add service discovery registration
7. Create data migration scripts

Ensure the service is production-ready and well-tested.`,
						WaitForPlan: false,
						AutoApprove: true,
					},
					{
						Name:        "Service Communication",
						Description: "Implement inter-service communication",
						Prompt: `Implement robust communication between services:

1. Set up gRPC for synchronous communication
2. Implement message queue for async events (e.g., Kafka/RabbitMQ)
3. Add circuit breakers for resilience (e.g., Hystrix pattern)
4. Implement retry logic with exponential backoff
5. Set up service discovery (e.g., Consul, Eureka)
6. Add API gateway for external access
7. Implement distributed tracing

Make the system resilient to failures.`,
						WaitForPlan: false,
						AutoApprove: true,
					},
				},
			},
			{
				Name:        "Data Migration",
				Description: "Migrate data and ensure consistency",
				Timeout:     90 * time.Minute,
				Tasks: []automation.Task{
					{
						Name:        "Database Splitting",
						Description: "Split databases and migrate data",
						Prompt: `Implement database per service pattern:

1. Create separate database for user service
2. Implement zero-downtime data migration scripts
3. Add eventual consistency handling
4. Implement Saga pattern for distributed transactions
5. Add event sourcing for audit trail
6. Create rollback procedures
7. Add data synchronization monitoring

Ensure data integrity throughout the migration.`,
						WaitForPlan: true,
						AutoApprove: autoApprove,
					},
				},
			},
		},
		OnPhaseComplete: func(phaseIndex int, result automation.PhaseResult) error {
			fmt.Printf("\n✅ Phase %d completed: %s\n", phaseIndex+1, result.PhaseName)
			fmt.Printf("   Duration: %v\n", result.Duration)
			return nil
		},
		OnWorkflowComplete: func(result automation.WorkflowResult) error {
			fmt.Printf("\n🎉 Microservices migration workflow completed!\n")
			fmt.Printf("   Total duration: %v\n", result.TotalDuration)
			fmt.Printf("   Session: %s\n", result.SessionID)
			return nil
		},
	}

	// Create and run orchestrator
	orchestratorConfig := automation.DefaultOrchestratorConfig()
	orchestratorConfig.AutoApprove = autoApprove

	orchestrator := automation.NewSessionOrchestrator(julesClient, workflow, orchestratorConfig)

	go monitorProgress(orchestrator)

	fmt.Println("🚀 Starting Microservices Migration Workflow")
	fmt.Println("============================================")
	fmt.Printf("Source: %s\n\n", sourceID)

	err := orchestrator.Start(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("workflow failed: %w", err)
	}

	displayExecutionSummary(orchestrator)
	return nil
}

func monitorProgress(orchestrator *automation.SessionOrchestrator) {
	for {
		select {
		case progress := <-orchestrator.ProgressChannel():
			fmt.Printf("📊 [Phase %d/%d] %s (%.1f%%)\n",
				progress.Phase+1,
				progress.TotalPhases,
				progress.Message,
				progress.Progress,
			)
		case activity := <-orchestrator.ActivityChannel():
			fmt.Printf("⚡ Activity: %s\n", activity.Message)
		case err := <-orchestrator.ErrorChannel():
			fmt.Printf("❌ Error: %v\n", err)
		}
	}
}

func displayExecutionSummary(orchestrator *automation.SessionOrchestrator) {
	log := orchestrator.GetExecutionLog()

	fmt.Println("\n📋 Execution Summary")
	fmt.Println("===================")
	fmt.Printf("Session ID: %s\n", orchestrator.GetSessionID())
	fmt.Printf("Total tasks: %d\n", len(log))

	successCount := 0
	totalDuration := time.Duration(0)

	for _, record := range log {
		if record.Success {
			successCount++
		}
		totalDuration += record.Duration
	}

	fmt.Printf("Successful: %d/%d\n", successCount, len(log))
	fmt.Printf("Total duration: %v\n", totalDuration)

	fmt.Println("\nTask Details:")
	for i, record := range log {
		status := "✅"
		if !record.Success {
			status = "❌"
		}
		fmt.Printf("  %d. %s %s (%v)\n", i+1, status, record.TaskName, record.Duration)
		if record.ArtifactCount > 0 {
			fmt.Printf("     Artifacts: %d\n", record.ArtifactCount)
		}
		if record.Error != "" {
			fmt.Printf("     Error: %s\n", record.Error)
		}
	}

	fmt.Printf("\n💡 View session at: https://jules.google.com/session/%s\n", orchestrator.GetSessionID())
}
