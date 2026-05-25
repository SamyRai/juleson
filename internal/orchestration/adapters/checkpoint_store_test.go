package adapters

import (
	"context"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

func TestJSONCheckpointStoreSavesAndLoadsCheckpoint(t *testing.T) {
	store := NewJSONCheckpointStore(t.TempDir())
	checkpoint := domain.Checkpoint{
		ID:     "goal/planned:0",
		GoalID: "goal",
		State:  domain.StatePlanning,
		Context: domain.ExecutionContext{
			Goal: domain.Goal{ID: "goal", Description: "ship"},
		},
		CreatedAt: time.Unix(100, 0),
		Metadata:  map[string]string{"phase": "planned"},
	}

	if err := store.SaveCheckpoint(context.Background(), checkpoint); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}
	loaded, err := store.LoadCheckpoint(context.Background(), checkpoint.ID)
	if err != nil {
		t.Fatalf("LoadCheckpoint() error = %v", err)
	}
	if loaded.ID != checkpoint.ID || loaded.Metadata["phase"] != "planned" {
		t.Fatalf("loaded checkpoint = %+v", loaded)
	}
}
