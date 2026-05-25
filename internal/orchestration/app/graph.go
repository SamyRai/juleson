package app

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

const (
	graphEndNode  = "__end__"
	graphFailNode = "__fail__"
)

type graphNode func(context.Context, *appRunState) (string, error)

type agentGraph struct {
	start string
	nodes map[string]graphNode
}

type appRunState struct {
	goal          domain.Goal
	project       *domain.ProjectContext
	template      *domain.Template
	values        map[string]string
	plan          *domain.Plan
	ordered       []domain.Task
	result        *domain.Result
	outputFiles   []domain.OutputFile
	outputs       []string
	execution     domain.ExecutionContext
	agentOptions  AgentRunOptions
	aiOptions     AIWorkflowRunOptions
	startedAt     time.Time
	iteration     int
	maxIterations int
	decision      *domain.Decision
	selectedTask  domain.Task
	err           error
}

func newAgentGraph(start string, nodes map[string]graphNode) (agentGraph, error) {
	if start == "" {
		return agentGraph{}, fmt.Errorf("graph start node is required")
	}
	if _, ok := nodes[start]; !ok {
		return agentGraph{}, fmt.Errorf("graph start node %q is not registered", start)
	}
	if _, ok := nodes[graphEndNode]; ok {
		return agentGraph{}, fmt.Errorf("graph node %q is reserved", graphEndNode)
	}
	if _, ok := nodes[graphFailNode]; ok {
		return agentGraph{}, fmt.Errorf("graph node %q is reserved", graphFailNode)
	}
	return agentGraph{start: start, nodes: nodes}, nil
}

func (g agentGraph) run(ctx context.Context, state *appRunState) error {
	if state == nil {
		return fmt.Errorf("graph state is required")
	}

	current := g.start
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		switch current {
		case graphEndNode:
			return nil
		case graphFailNode:
			if state.err != nil {
				return state.err
			}
			return fmt.Errorf("graph failed")
		}

		node, ok := g.nodes[current]
		if !ok {
			return fmt.Errorf("unknown graph node %q", current)
		}
		next, err := node(ctx, state)
		if err != nil {
			return err
		}
		if next == "" {
			return fmt.Errorf("graph node %q returned empty next node", current)
		}
		current = next
	}
}
