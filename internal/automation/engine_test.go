package automation

import (
	"testing"

	"github.com/SamyRai/juleson/internal/jules"
)

func TestNewSessionOrchestrator(t *testing.T) {
	client := &jules.Client{}
	workflow := &WorkflowDefinition{}
	config := DefaultOrchestratorConfig()

	// Test valid creation
	orchestrator := NewSessionOrchestrator(client, workflow, config)
	if orchestrator == nil {
		t.Fatal("Expected non-nil orchestrator")
	}

	if orchestrator.client != client {
		t.Error("Expected client to be set")
	}

	if orchestrator.workflow != workflow {
		t.Error("Expected workflow to be set")
	}

	// Check default values
	if orchestrator.checkInterval != DefaultCheckInterval {
		t.Errorf("Expected checkInterval %v, got %v", DefaultCheckInterval, orchestrator.checkInterval)
	}

	if orchestrator.maxSessionAge != DefaultMaxSessionAge {
		t.Errorf("Expected maxSessionAge %v, got %v", DefaultMaxSessionAge, orchestrator.maxSessionAge)
	}
}

func TestDefaultOrchestratorConfig(t *testing.T) {
	config := DefaultOrchestratorConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.CheckInterval != DefaultCheckInterval {
		t.Errorf("Expected CheckInterval %v, got %v", DefaultCheckInterval, config.CheckInterval)
	}

	if config.MaxSessionAge != DefaultMaxSessionAge {
		t.Errorf("Expected MaxSessionAge %v, got %v", DefaultMaxSessionAge, config.MaxSessionAge)
	}

	if config.RetryAttempts != DefaultRetryAttempts {
		t.Errorf("Expected RetryAttempts %d, got %d", DefaultRetryAttempts, config.RetryAttempts)
	}
}
