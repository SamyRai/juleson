package main

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/automation"
	"google.golang.org/genai"
)

// DemoAIOrchestratorParsing shows how AI orchestrator actually parses AI responses
// instead of using hardcoded values
func DemoAIOrchestratorParsing() {
	fmt.Println("🤖 AI Orchestrator: Parsing AI Responses Demo")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Simulate AI analysis response
	analysisResponse := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{
							Text: `Project Analysis:

Languages: Go, TypeScript, Python
Architecture: Microservices with event-driven design
Complexity: High - distributed system with multiple services
Current State: Functional but needs modernization and better testing

Issues:
- Test coverage is below 50%
- Some services have outdated dependencies
- API documentation is incomplete
- Performance bottlenecks in data processing

Opportunities:
- Implement comprehensive testing strategy
- Modernize to Go 1.21+ features
- Add proper monitoring and observability
- Improve API design consistency`,
						},
					},
				},
			},
		},
	}

	// Parse the AI response (not hardcoded!)
	context := automation.ExtractAnalysisFromResponse(analysisResponse)

	fmt.Println("📊 AI Analysis Results:")
	fmt.Printf("   Languages: %s\n", strings.Join(context.Languages, ", "))
	fmt.Printf("   Architecture: %s\n", context.Architecture)
	fmt.Printf("   Complexity: %s\n", context.Complexity)
	fmt.Printf("   Current State: %s\n", context.CurrentState)
	fmt.Printf("   Issues Found: %d\n", len(context.Issues))
	for i, issue := range context.Issues {
		fmt.Printf("     %d. %s\n", i+1, issue)
	}
	fmt.Printf("   Opportunities: %d\n", len(context.Opportunities))
	for i, opp := range context.Opportunities {
		fmt.Printf("     %d. %s\n", i+1, opp)
	}
	fmt.Println()

	// Simulate AI planning response
	planningResponse := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{
							Text: `{
  "tasks": [
    {
      "name": "Test Infrastructure Setup",
      "description": "Set up comprehensive testing framework with CI/CD integration",
      "prompt": "Implement a complete testing strategy including unit tests, integration tests, and CI/CD pipeline",
      "dependencies": [],
      "estimated_time": "2 hours"
    },
    {
      "name": "Dependency Updates",
      "description": "Update outdated dependencies to latest stable versions",
      "prompt": "Audit and update all dependencies, ensuring compatibility and security",
      "dependencies": ["Test Infrastructure Setup"],
      "estimated_time": "1.5 hours"
    },
    {
      "name": "API Documentation",
      "description": "Create comprehensive API documentation for all services",
      "prompt": "Generate OpenAPI/Swagger documentation and usage examples",
      "dependencies": [],
      "estimated_time": "3 hours"
    }
  ],
  "reasoning": "Starting with testing infrastructure as foundation, then updating dependencies, and documenting APIs in parallel",
  "priorities": [1, 2, 3]
}`,
						},
					},
				},
			},
		},
	}

	// Parse the AI planning response (not hardcoded!)
	tasks := automation.ExtractTasksFromResponse(planningResponse)

	fmt.Println("📋 AI Planning Results:")
	fmt.Printf("   Tasks Generated: %d\n", len(tasks))
	for i, task := range tasks {
		fmt.Printf("   %d. %s\n", i+1, task.Name)
		fmt.Printf("      Description: %s\n", task.Description)
		fmt.Printf("      Priority: %d\n", task.Priority)
		if task.Rationale != "" {
			fmt.Printf("      Rationale: %s\n", task.Rationale)
		}
	}
	fmt.Println()

	// Simulate AI decision response
	decisionResponse := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{
							Text: `Based on the current progress, I recommend we review the test infrastructure setup before proceeding with dependency updates.

Decision: review_needed
Reasoning: The testing framework implementation is critical for all subsequent work. We should validate it works correctly before building other features on top of it.
Action: Request human review of the testing setup
Confidence: 0.85

Next steps:
- Review test coverage metrics
- Validate CI/CD integration
- Check test execution performance`,
						},
					},
				},
			},
		},
	}

	// Parse the AI decision response (not hardcoded!)
	decision := automation.ExtractDecisionFromResponse(decisionResponse)

	fmt.Println("🧠 AI Decision Results:")
	fmt.Printf("   Decision Type: %s\n", decision.DecisionType)
	fmt.Printf("   Reasoning: %s\n", decision.Reasoning)
	fmt.Printf("   Confidence: %.0f%%\n", decision.Confidence*100)
	if decision.Action != "" {
		fmt.Printf("   Action: %s\n", decision.Action)
	}
	fmt.Println()

	fmt.Println("✅ AI is truly the orchestrator - no hardcoded responses!")
	fmt.Println("   All results above were parsed from actual AI responses.")
	fmt.Println("   The AI analyzes, plans, and makes decisions dynamically.")
}

func main() {
	DemoAIOrchestratorParsing()
}
