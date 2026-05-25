package automation

import "fmt"

// buildInitialPrompt creates a comprehensive initial prompt from workflow.
func (o *SessionOrchestrator) buildInitialPrompt() string {
	prompt := fmt.Sprintf("Execute workflow: %s\n\n%s\n\n", o.workflow.Name, o.workflow.Description)

	prompt += "This is a multi-phase workflow that will be executed progressively. "
	prompt += "Please be ready to receive follow-up messages for each phase.\n\n"

	if len(o.workflow.Phases) > 0 {
		firstPhase := o.workflow.Phases[0]
		prompt += fmt.Sprintf("Starting with Phase 1: %s\n%s\n\n", firstPhase.Name, firstPhase.Description)

		if len(firstPhase.Tasks) > 0 {
			prompt += "Initial tasks:\n"
			for _, task := range firstPhase.Tasks {
				prompt += fmt.Sprintf("- %s: %s\n", task.Name, task.Description)
			}
		}
	}

	return prompt
}
