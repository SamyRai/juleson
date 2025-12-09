package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

// Checkpoint file system permissions
const (
	// CheckpointDirPerm is the permission for checkpoint directories (rwxr-xr-x)
	CheckpointDirPerm = 0755
	// CheckpointFilePerm is the permission for checkpoint files (rw-r--r--)
	CheckpointFilePerm = 0644
)

// Checkpoint represents a saved agent state
type Checkpoint struct {
	ID             string                 `json:"id"`
	Timestamp      time.Time              `json:"timestamp"`
	State          agent.AgentState       `json:"state"`
	Goal           agent.Goal             `json:"goal"`
	CurrentPlan    []agent.Task           `json:"current_plan"`
	Decisions      []agent.Decision       `json:"decisions"`
	CompletedTasks []agent.TaskResult     `json:"completed_tasks"`
	Iteration      int                    `json:"iteration"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CheckpointManager handles saving and restoring agent state
type CheckpointManager struct {
	checkpointDir string
	autoSave      bool
	saveInterval  time.Duration
	logger        *slog.Logger
	mu            sync.RWMutex
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(checkpointDir string, autoSave bool, saveInterval time.Duration, logger *slog.Logger) *CheckpointManager {
	if checkpointDir == "" {
		checkpointDir = "./checkpoints"
	}
	if saveInterval <= 0 {
		saveInterval = DefaultCheckpointInterval
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CheckpointManager{
		checkpointDir: checkpointDir,
		autoSave:      autoSave,
		saveInterval:  saveInterval,
		logger:        logger,
	}
}

// Save creates a checkpoint of the current agent state
func (cm *CheckpointManager) Save(ctx context.Context, agent *CoreAgent) (*Checkpoint, error) {
	if agent == nil {
		return nil, fmt.Errorf("cannot save checkpoint: agent is nil")
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	checkpoint := &Checkpoint{
		ID:          fmt.Sprintf("checkpoint-%d", time.Now().Unix()),
		Timestamp:   time.Now(),
		State:       agent.state,
		Decisions:   agent.decisions,
		CurrentPlan: agent.currentPlan,
		Iteration:   0, // TODO: Add iteration tracking to CoreAgent
		Metadata:    make(map[string]interface{}),
	}

	if agent.currentGoal != nil {
		checkpoint.Goal = *agent.currentGoal
	}

	// Collect completed tasks
	for _, task := range agent.currentPlan {
		if task.Result != nil && task.State == "COMPLETE" {
			checkpoint.CompletedTasks = append(checkpoint.CompletedTasks, *task.Result)
		}
	}

	// Save to disk
	if err := cm.writeCheckpoint(checkpoint); err != nil {
		cm.logger.Error("failed to write checkpoint", "error", err)
		return nil, fmt.Errorf("failed to write checkpoint: %w", err)
	}

	cm.logger.Info("checkpoint saved", "id", checkpoint.ID)
	return checkpoint, nil
}

// Restore loads a checkpoint and restores agent state
func (cm *CheckpointManager) Restore(ctx context.Context, checkpointID string, agent *CoreAgent) error {
	if agent == nil {
		return fmt.Errorf("cannot restore checkpoint: agent is nil")
	}
	if checkpointID == "" {
		return fmt.Errorf("cannot restore checkpoint: checkpointID is empty")
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	checkpoint, err := cm.readCheckpoint(checkpointID)
	if err != nil {
		return fmt.Errorf("failed to read checkpoint: %w", err)
	}

	// Restore agent state
	agent.state = checkpoint.State
	agent.currentGoal = &checkpoint.Goal
	agent.currentPlan = checkpoint.CurrentPlan
	agent.decisions = checkpoint.Decisions

	return nil
}

// List returns all available checkpoints
func (cm *CheckpointManager) List() ([]Checkpoint, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	files, err := filepath.Glob(filepath.Join(cm.checkpointDir, "checkpoint-*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}

	var checkpoints []Checkpoint
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var checkpoint Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	return checkpoints, nil
}

// Delete removes a checkpoint
func (cm *CheckpointManager) Delete(checkpointID string) error {
	if checkpointID == "" {
		return fmt.Errorf("cannot delete checkpoint: checkpointID is empty")
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	checkpointPath := filepath.Join(cm.checkpointDir, checkpointID+".json")
	if err := os.Remove(checkpointPath); err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}

	return nil
}

// writeCheckpoint writes a checkpoint to disk
func (cm *CheckpointManager) writeCheckpoint(checkpoint *Checkpoint) error {
	// Ensure checkpoint directory exists
	if err := os.MkdirAll(cm.checkpointDir, CheckpointDirPerm); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Marshal checkpoint
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	// Write to file
	checkpointPath := filepath.Join(cm.checkpointDir, checkpoint.ID+".json")
	if err := os.WriteFile(checkpointPath, data, CheckpointFilePerm); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return nil
}

// readCheckpoint reads a checkpoint from disk
func (cm *CheckpointManager) readCheckpoint(checkpointID string) (*Checkpoint, error) {
	checkpointPath := filepath.Join(cm.checkpointDir, checkpointID+".json")

	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// StartAutoSave begins automatic checkpoint creation
func (cm *CheckpointManager) StartAutoSave(ctx context.Context, agent *CoreAgent) {
	if !cm.autoSave {
		return
	}

	cm.logger.Info("checkpoint.auto_save.started", "interval", cm.saveInterval)

	go func() {
		ticker := time.NewTicker(cm.saveInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if agent.state != "IDLE" && agent.state != "COMPLETE" {
					if _, err := cm.Save(ctx, agent); err != nil {
						cm.logger.Error("checkpoint.auto_save.failed", "error", err)
					}
				}
			case <-ctx.Done():
				cm.logger.Info("checkpoint.auto_save.stopped")
				return
			}
		}
	}()
}
