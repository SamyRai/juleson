package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

func TestAgentGraphRunsLinearNodes(t *testing.T) {
	var order []string
	graph, err := newAgentGraph("a", map[string]graphNode{
		"a": func(ctx context.Context, state *appRunState) (string, error) {
			order = append(order, "a")
			return "b", nil
		},
		"b": func(ctx context.Context, state *appRunState) (string, error) {
			order = append(order, "b")
			return graphEndNode, nil
		},
	},
	)
	if err != nil {
		t.Fatalf("newAgentGraph() error = %v", err)
	}

	if err := graph.run(context.Background(), &appRunState{}); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if strings.Join(order, ",") != "a,b" {
		t.Fatalf("order = %v, want [a b]", order)
	}
}

func TestAgentGraphRoutesConditionally(t *testing.T) {
	graph, err := newAgentGraph("start", map[string]graphNode{
		"start": func(ctx context.Context, state *appRunState) (string, error) {
			if state.agentOptions.DryRun {
				return "dry", nil
			}
			return "execute", nil
		},
		"dry": func(ctx context.Context, state *appRunState) (string, error) {
			state.result = &domain.Result{Summary: "dry"}
			return graphEndNode, nil
		},
		"execute": func(ctx context.Context, state *appRunState) (string, error) {
			state.result = &domain.Result{Summary: "execute"}
			return graphEndNode, nil
		},
	},
	)
	if err != nil {
		t.Fatalf("newAgentGraph() error = %v", err)
	}

	state := &appRunState{agentOptions: AgentRunOptions{DryRun: true}}
	if err := graph.run(context.Background(), state); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if state.result.Summary != "dry" {
		t.Fatalf("summary = %q, want dry", state.result.Summary)
	}
}

func TestAgentGraphRejectsUnknownNextNode(t *testing.T) {
	graph, err := newAgentGraph("start", map[string]graphNode{
		"start": func(ctx context.Context, state *appRunState) (string, error) {
			return "missing", nil
		},
	},
	)
	if err != nil {
		t.Fatalf("newAgentGraph() error = %v", err)
	}

	runErr := graph.run(context.Background(), &appRunState{})
	if runErr == nil || !strings.Contains(runErr.Error(), "unknown graph node") {
		t.Fatalf("run() error = %v, want unknown graph node", runErr)
	}
}

func TestAgentGraphHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	graph, err := newAgentGraph("start", map[string]graphNode{
		"start": func(ctx context.Context, state *appRunState) (string, error) {
			return graphEndNode, nil
		},
	},
	)
	if err != nil {
		t.Fatalf("newAgentGraph() error = %v", err)
	}

	runErr := graph.run(ctx, &appRunState{})
	if !errors.Is(runErr, context.Canceled) {
		t.Fatalf("run() error = %v, want context.Canceled", runErr)
	}
}

func TestAgentGraphRejectsMissingStartNode(t *testing.T) {
	_, err := newAgentGraph("missing", map[string]graphNode{
		"start": func(ctx context.Context, state *appRunState) (string, error) {
			return graphEndNode, nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "start node") {
		t.Fatalf("newAgentGraph() error = %v, want missing start node", err)
	}
}

func TestAgentGraphRejectsReservedTerminalNames(t *testing.T) {
	for _, reserved := range []string{graphEndNode, graphFailNode} {
		_, err := newAgentGraph("start", map[string]graphNode{
			"start": func(ctx context.Context, state *appRunState) (string, error) {
				return graphEndNode, nil
			},
			reserved: func(ctx context.Context, state *appRunState) (string, error) {
				return graphEndNode, nil
			},
		})
		if err == nil || !strings.Contains(err.Error(), "reserved") {
			t.Fatalf("newAgentGraph(%q) error = %v, want reserved node error", reserved, err)
		}
	}
}
