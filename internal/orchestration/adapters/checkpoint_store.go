package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

const defaultCheckpointPath = "./data/checkpoints"

type JSONCheckpointStore struct {
	dir string
}

func NewJSONCheckpointStore(dir string) *JSONCheckpointStore {
	if strings.TrimSpace(dir) == "" {
		dir = defaultCheckpointPath
	}
	return &JSONCheckpointStore{dir: dir}
}

func (s *JSONCheckpointStore) SaveCheckpoint(ctx context.Context, checkpoint domain.Checkpoint) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if checkpoint.ID == "" {
		return fmt.Errorf("checkpoint ID cannot be empty")
	}
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("create checkpoint directory: %w", err)
	}
	path := s.path(checkpoint.ID)
	tempPath := path + ".tmp"
	data, err := json.MarshalIndent(sanitizeCheckpoint(checkpoint), "", "  ")
	if err != nil {
		return fmt.Errorf("encode checkpoint: %w", err)
	}
	if err := os.WriteFile(tempPath, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("write checkpoint: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace checkpoint: %w", err)
	}
	return nil
}

func (s *JSONCheckpointStore) LoadCheckpoint(ctx context.Context, id string) (*domain.Checkpoint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("checkpoint ID cannot be empty")
	}
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}
	var checkpoint domain.Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("decode checkpoint: %w", err)
	}
	return &checkpoint, nil
}

func (s *JSONCheckpointStore) path(id string) string {
	filename := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(id) + ".json"
	return filepath.Join(s.dir, filename)
}

func sanitizeCheckpoint(checkpoint domain.Checkpoint) domain.Checkpoint {
	checkpoint.Context.Completed = sanitizedTaskResults(checkpoint.Context.Completed)
	if checkpoint.Context.Plan != nil {
		for i := range checkpoint.Context.Plan.Tasks {
			if checkpoint.Context.Plan.Tasks[i].Result != nil {
				checkpoint.Context.Plan.Tasks[i].Result.Error = nil
			}
		}
	}
	for i := range checkpoint.Context.Decisions {
		if checkpoint.Context.Decisions[i].Outcome != nil {
			checkpoint.Context.Decisions[i].Outcome.Error = nil
		}
	}
	return checkpoint
}

func sanitizedTaskResults(results []domain.TaskResult) []domain.TaskResult {
	sanitized := append([]domain.TaskResult(nil), results...)
	for i := range sanitized {
		sanitized[i].Error = nil
	}
	return sanitized
}
