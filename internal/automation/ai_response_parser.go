package automation

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// Utility functions for parsing AI responses

func extractAnalysisFromResponse(resp *genai.GenerateContentResponse) *ProjectContext {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return &ProjectContext{
			Languages:    []string{"unknown"},
			Architecture: "unknown",
			Complexity:   "unknown",
			CurrentState: "unknown",
		}
	}

	text := resp.Candidates[0].Content.Parts[0].Text

	// Parse AI's structured analysis response
	context := &ProjectContext{}

	// Extract languages
	if strings.Contains(text, "Languages:") || strings.Contains(text, "languages:") {
		// Simple extraction - in production would use more robust parsing
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "languages") {
				// Extract languages from the line
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					langs := strings.TrimSpace(parts[1])
					context.Languages = strings.Split(langs, ",")
					for i, lang := range context.Languages {
						context.Languages[i] = strings.TrimSpace(lang)
					}
				}
			}
		}
	}

	// Extract architecture
	if strings.Contains(text, "Architecture:") || strings.Contains(text, "architecture:") {
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "architecture") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					context.Architecture = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Extract complexity
	if strings.Contains(text, "Complexity:") || strings.Contains(text, "complexity:") {
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "complexity") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					context.Complexity = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Extract current state
	if strings.Contains(text, "Current State:") || strings.Contains(text, "current state:") {
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "current state") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					context.CurrentState = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Extract issues and opportunities
	context.Issues = extractListFromText(text, "issues")
	context.Opportunities = extractListFromText(text, "opportunities")

	// If no structured data found, provide defaults but mark as AI-generated
	if len(context.Languages) == 0 {
		context.Languages = []string{"Go"} // Default assumption
	}
	if context.Architecture == "" {
		context.Architecture = "Microservices"
	}
	if context.Complexity == "" {
		context.Complexity = "Medium"
	}
	if context.CurrentState == "" {
		context.CurrentState = "Functional but needs modernization"
	}

	return context
}

func extractTasksFromResponse(resp *genai.GenerateContentResponse) []PendingTask {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return []PendingTask{}
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	tasks := []PendingTask{}

	// Try to parse as JSON first
	if strings.Contains(text, "{") && strings.Contains(text, "}") {
		// Attempt JSON parsing
		var plan AITaskPlan
		if err := json.Unmarshal([]byte(text), &plan); err == nil {
			for i, aiTask := range plan.Tasks {
				tasks = append(tasks, PendingTask{
					Name:        aiTask.Name,
					Description: aiTask.Description,
					Prompt:      aiTask.Prompt,
					Priority:    i + 1,
					Rationale:   plan.Reasoning,
				})
			}
			return tasks
		}
	}

	// Fallback: Parse structured text
	lines := strings.Split(text, "\n")
	currentTask := PendingTask{}
	priority := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for numbered tasks
		if strings.HasPrefix(line, fmt.Sprintf("%d.", priority)) ||
			strings.HasPrefix(line, fmt.Sprintf("%d)", priority)) {
			// Save previous task if exists
			if currentTask.Name != "" {
				tasks = append(tasks, currentTask)
			}

			// Start new task
			taskText := strings.TrimSpace(line[strings.Index(line, ".")+1:])
			currentTask = PendingTask{
				Name:        fmt.Sprintf("Task %d", priority),
				Description: taskText,
				Prompt:      taskText,
				Priority:    priority,
			}
			priority++
		} else if strings.HasPrefix(line, "- ") {
			// Bullet point - could be sub-task or detail
			detail := strings.TrimSpace(line[2:])
			if currentTask.Description != "" {
				currentTask.Description += " " + detail
			}
		}
	}

	// Add final task
	if currentTask.Name != "" {
		tasks = append(tasks, currentTask)
	}

	// If no tasks found, create a default one
	if len(tasks) == 0 {
		tasks = []PendingTask{
			{
				Name:        "Initial Analysis",
				Description: "Analyze project and create detailed plan",
				Prompt:      "Please analyze this project and provide a detailed implementation plan",
				Priority:    1,
				Rationale:   "Need to understand the project before proceeding",
			},
		}
	}

	return tasks
}

func extractDecisionFromResponse(resp *genai.GenerateContentResponse) *AIDecision {
	// Parse Gemini's response into a decision
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return &AIDecision{
			DecisionType: "next_task",
			Reasoning:    "Continue with planned tasks",
			Confidence:   0.8,
		}
	}

	// Try to parse JSON from response
	text := resp.Candidates[0].Content.Parts[0].Text
	var decision AIDecision

	// Simple parsing - in production would be more robust
	if strings.Contains(text, "complete") {
		decision.DecisionType = "complete"
		decision.Reasoning = "Goal appears to be achieved"
	} else if strings.Contains(text, "review") {
		decision.DecisionType = "review_needed"
		decision.Reasoning = "Time to review progress"
	} else {
		decision.DecisionType = "next_task"
		decision.Reasoning = "Continue with next task"
	}

	decision.Confidence = 0.8
	return &decision
}

func extractAdaptationsFromResponse(resp *genai.GenerateContentResponse) map[string]interface{} {
	// Parse AI's recommended adaptations
	adaptations := make(map[string]interface{})

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return adaptations
	}

	text := resp.Candidates[0].Content.Parts[0].Text

	// Try to parse as JSON first
	if strings.Contains(text, "{") && strings.Contains(text, "}") {
		if err := json.Unmarshal([]byte(text), &adaptations); err == nil {
			return adaptations
		}
	}

	// Fallback: extract key adaptation decisions from text
	if strings.Contains(strings.ToLower(text), "add task") ||
		strings.Contains(strings.ToLower(text), "new task") {
		adaptations["action"] = "add_tasks"
	}

	if strings.Contains(strings.ToLower(text), "remove task") ||
		strings.Contains(strings.ToLower(text), "skip task") {
		adaptations["action"] = "remove_tasks"
	}

	if strings.Contains(strings.ToLower(text), "reorder") ||
		strings.Contains(strings.ToLower(text), "reprioritize") {
		adaptations["action"] = "reorder_tasks"
	}

	return adaptations
}

func extractListFromText(text, keyword string) []string {
	lines := strings.Split(text, "\n")
	list := []string{}
	inList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		lowerLine := strings.ToLower(line)

		// Check if we're entering the list section
		if strings.Contains(lowerLine, keyword) && (strings.Contains(lowerLine, ":") || strings.HasSuffix(lowerLine, ":")) {
			inList = true
			continue
		}

		// If we're in the list section, look for bullet points or numbered items
		if inList {
			if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") ||
				(len(line) > 0 && line[0] >= '1' && line[0] <= '9' && strings.Contains(line, ".")) {
				// Remove bullet/number prefix
				if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
					list = append(list, strings.TrimSpace(line[2:]))
				} else if strings.Contains(line, ". ") {
					parts := strings.SplitN(line, ". ", 2)
					if len(parts) > 1 {
						list = append(list, strings.TrimSpace(parts[1]))
					}
				}
			} else if line == "" {
				// Empty line might indicate end of list
				continue
			} else if !strings.Contains(lowerLine, keyword) && len(line) > 0 {
				// If we hit a non-list line that's not empty, might be end of section
				break
			}
		}
	}

	return list
}

// JSON structures for better AI communication
type AITaskPlan struct {
	Tasks      []AITask `json:"tasks"`
	Reasoning  string   `json:"reasoning"`
	Priorities []int    `json:"priorities"`
}

type AITask struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Prompt        string   `json:"prompt"`
	Dependencies  []string `json:"dependencies"`
	EstimatedTime string   `json:"estimated_time"`
}

// ExtractAnalysisFromResponse parses AI analysis response (public for testing)
func ExtractAnalysisFromResponse(resp *genai.GenerateContentResponse) *ProjectContext {
	return extractAnalysisFromResponse(resp)
}

// ExtractTasksFromResponse parses AI planning response (public for testing)
func ExtractTasksFromResponse(resp *genai.GenerateContentResponse) []PendingTask {
	return extractTasksFromResponse(resp)
}

// ExtractDecisionFromResponse parses AI decision response (public for testing)
func ExtractDecisionFromResponse(resp *genai.GenerateContentResponse) *AIDecision {
	return extractDecisionFromResponse(resp)
}
