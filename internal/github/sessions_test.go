package github

import (
	"context"
	"testing"
)

func TestCreateSessionFromRepo_NoJulesClient(t *testing.T) {
	service := &SessionService{}

	_, err := service.CreateSessionFromRepo(context.Background(), "test prompt", "owner", "repo", "main")
	if err == nil {
		t.Errorf("expected error when Jules client is nil")
	}
	if err != nil && err.Error() != "Jules client not available" {
		t.Errorf("expected 'Jules client not available', got: %v", err)
	}
}

func TestCreateSessionFromCurrentRepo_NoJulesClient(t *testing.T) {
	service := &SessionService{}

	_, err := service.CreateSessionFromCurrentRepo(context.Background(), "test prompt", "main")
	if err == nil {
		t.Errorf("expected error when Jules client is nil")
	}
	if err != nil && err.Error() != "Jules client not available" {
		t.Errorf("expected 'Jules client not available', got: %v", err)
	}
}
