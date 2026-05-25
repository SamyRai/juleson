package tools

import (
	"context"
	"testing"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestrateWorkflowStopsOnFailureByDefault(t *testing.T) {
	container := services.NewContainer(&config.Config{})

	_, output, err := orchestrateWorkflow(context.Background(), nil, OrchestrateWorkflowInput{
		WorkflowSteps: []WorkflowStep{
			{Name: "missing template", Tool: "execute_template", Parameters: map[string]string{}},
			{Name: "later", Tool: "analyze_project", Parameters: map[string]string{}},
		},
	}, container)

	require.NoError(t, err)
	assert.Equal(t, "failed", output.OverallStatus)
	assert.Len(t, output.ExecutionResults, 1)
	assert.Equal(t, "template_name parameter required", output.ExecutionResults[0].Error)
	assert.Contains(t, output.Issues[0], "missing template")
}

func TestOrchestrateWorkflowContinuesOnError(t *testing.T) {
	container := services.NewContainer(&config.Config{})

	_, output, err := orchestrateWorkflow(context.Background(), nil, OrchestrateWorkflowInput{
		ContinueOnError: true,
		WorkflowSteps: []WorkflowStep{
			{Name: "missing template", Tool: "execute_template", Parameters: map[string]string{}},
			{Name: "analysis", Tool: "analyze_project", Parameters: map[string]string{}},
		},
	}, container)

	require.NoError(t, err)
	assert.Equal(t, "failed", output.OverallStatus)
	assert.Len(t, output.ExecutionResults, 2)
	assert.Equal(t, "completed", output.ExecutionResults[1].Status)
}
