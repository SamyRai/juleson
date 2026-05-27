package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration"
	"github.com/SamyRai/juleson/internal/orchestration/domain"

	"github.com/spf13/cobra"
)

// NewOrchestrateCommand creates the orchestrate command for multi-task workflows
func NewOrchestrateCommand(initializeRuntime func() (*orchestration.Runtime, error)) *cobra.Command {
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
	cmd.AddCommand(newOrchestrateAPICommand(initializeRuntime))
	cmd.AddCommand(newOrchestrateMicroservicesCommand(initializeRuntime))
	cmd.AddCommand(newOrchestrateCustomCommand())

	return cmd
}

// newOrchestrateAPICommand creates command for API modernization workflow
func newOrchestrateAPICommand(initializeRuntime func() (*orchestration.Runtime, error)) *cobra.Command {
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
			return runAPIModernizationWorkflow(initializeRuntime, sourceID, autoApprove)
		},
	}

	cmd.Flags().StringVarP(&sourceID, "source", "s", "", "Source ID (GitHub repo)")
	cmd.Flags().BoolVarP(&autoApprove, "auto-approve", "a", false, "Automatically approve plans")
	mustMarkFlagRequired(cmd, "source")

	return cmd
}

// newOrchestrateMicroservicesCommand creates command for microservices migration
func newOrchestrateMicroservicesCommand(initializeRuntime func() (*orchestration.Runtime, error)) *cobra.Command {
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
			return runMicroservicesMigrationWorkflow(initializeRuntime, sourceID, autoApprove)
		},
	}

	cmd.Flags().StringVarP(&sourceID, "source", "s", "", "Source ID (GitHub repo)")
	cmd.Flags().BoolVarP(&autoApprove, "auto-approve", "a", false, "Automatically approve plans")
	mustMarkFlagRequired(cmd, "source")

	return cmd
}

// newOrchestrateCustomCommand creates command for custom workflows
func newOrchestrateCustomCommand() *cobra.Command {
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

func runAPIModernizationWorkflow(initializeRuntime func() (*orchestration.Runtime, error), sourceID string, autoApprove bool) error {
	ctx := context.Background()

	workflow := domain.Workflow{
		Name:        "API Modernization",
		Description: "Modernize legacy API with authentication, testing, and CI/CD",
		MaxDuration: 4 * time.Hour,
		Phases: []domain.Phase{
			{
				Name:        "Foundation",
				Description: "Set up project structure and dependencies",
				Timeout:     45 * time.Minute,
				Tasks: []domain.Task{
					{
						ID:          "Project Analysis & Planning",
						Name:        "Project Analysis & Planning",
						Description: "Analyze API and create modernization plan",
						Prompt: `Analyze the current API structure and create a comprehensive modernization plan:

1. Review existing API endpoints and data models
2. Create OpenAPI 3.0 specification
3. Plan database schema updates and migrations
4. Identify breaking changes and create migration strategy
5. List all required dependencies and updates

Provide a detailed plan with estimated effort for each component.`,
						RequiresApproval: true,
					},
				},
			},
			{
				Name:        "Core Implementation",
				Description: "Implement authentication and refactor API",
				Timeout:     90 * time.Minute,
				Tasks: []domain.Task{
					{
						ID:          "JWT Authentication",
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
					},
					{
						ID:          "API Endpoint Refactoring",
						Name:        "API Endpoint Refactoring",
						Description: "Refactor existing endpoints",
						Prompt: `Refactor all API endpoints to use the new authentication system:

1. Update endpoints to use auth middleware
2. Add comprehensive request validation
3. Implement proper error handling with meaningful messages
4. Add request/response logging
5. Update endpoint documentation
6. Ensure backward compatibility where possible`,
					},
				},
			},
			{
				Name:        "Quality Assurance",
				Description: "Add testing and documentation",
				Timeout:     60 * time.Minute,
				Tasks: []domain.Task{
					{
						ID:          "Comprehensive Testing",
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
					},
					{
						ID:          "API Documentation",
						Name:        "API Documentation",
						Description: "Generate comprehensive documentation",
						Prompt: `Generate complete API documentation:

1. Generate API docs from OpenAPI specification
2. Add usage examples for each endpoint
3. Document authentication flows
4. Create migration guide from old API
5. Add troubleshooting section
6. Include code samples in multiple languages`,
					},
				},
			},
			{
				Name:        "DevOps Setup",
				Description: "Configure CI/CD and monitoring",
				Timeout:     45 * time.Minute,
				Tasks: []domain.Task{
					{
						ID:          "CI/CD Pipeline",
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
					},
					{
						ID:          "Monitoring & Observability",
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
					},
				},
			},
		},
	}

	// Start workflow
	fmt.Println("🚀 Starting API Modernization Workflow")
	fmt.Println("======================================")
	fmt.Printf("Source: %s\n", sourceID)
	fmt.Printf("Auto-approve: %v\n\n", autoApprove)

	return runDomainWorkflow(ctx, initializeRuntime, sourceID, autoApprove, workflow)
}

func runMicroservicesMigrationWorkflow(initializeRuntime func() (*orchestration.Runtime, error), sourceID string, autoApprove bool) error {
	ctx := context.Background()

	workflow := domain.Workflow{
		Name:        "Microservices Migration",
		Description: "Migrate monolith to microservices architecture",
		MaxDuration: 5 * time.Hour,
		Phases: []domain.Phase{
			{
				Name:        "Architecture Analysis",
				Description: "Analyze current architecture and plan decomposition",
				Timeout:     60 * time.Minute,
				Tasks: []domain.Task{
					{
						ID:          "Architecture Analysis",
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
						RequiresApproval: true,
					},
				},
			},
			{
				Name:        "First Service Extraction",
				Description: "Extract user service as proof of concept",
				Timeout:     150 * time.Minute,
				Tasks: []domain.Task{
					{
						ID:          "User Service Extraction",
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
					},
					{
						ID:          "Service Communication",
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
					},
				},
			},
			{
				Name:        "Data Migration",
				Description: "Migrate data and ensure consistency",
				Timeout:     90 * time.Minute,
				Tasks: []domain.Task{
					{
						ID:          "Database Splitting",
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
						RequiresApproval: true,
					},
				},
			},
		},
	}

	fmt.Println("🚀 Starting Microservices Migration Workflow")
	fmt.Println("============================================")
	fmt.Printf("Source: %s\n\n", sourceID)

	return runDomainWorkflow(ctx, initializeRuntime, sourceID, autoApprove, workflow)
}

func runDomainWorkflow(ctx context.Context, initializeRuntime func() (*orchestration.Runtime, error), sourceID string, autoApprove bool, workflow domain.Workflow) error {
	runtime, err := initializeRuntime()
	if err != nil {
		return err
	}
	execution := domain.ExecutionContext{
		Goal: domain.Goal{
			ID:          workflow.Name,
			Description: workflow.Description,
			Context: domain.GoalContext{
				SourceID: sourceID,
			},
		},
		ApprovalPolicy: domain.ApprovalPolicy{AutoApprove: autoApprove},
	}
	result, err := runtime.SessionWorkflowRunner().Run(ctx, workflow, execution)
	if err != nil {
		return fmt.Errorf("workflow failed: %w", err)
	}

	fmt.Println("\n📋 Execution Summary")
	fmt.Println("===================")
	fmt.Printf("Session ID: %s\n", result.SessionID)
	fmt.Printf("Phases: %d\n", result.TotalPhases)
	fmt.Printf("Total duration: %v\n", result.TotalDuration)
	for i, phase := range result.PhaseResults {
		status := "✅"
		if !phase.Success {
			status = "❌"
		}
		fmt.Printf("  %d. %s %s (%v)\n", i+1, status, phase.PhaseName, phase.Duration)
	}
	return nil
}
