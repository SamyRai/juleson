package app

import (
	"fmt"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

type taskScheduler struct{}

func (taskScheduler) Order(tasks []domain.Task) ([]domain.Task, error) {
	byID := make(map[string]domain.Task, len(tasks))
	for _, task := range tasks {
		key := task.ID
		if key == "" {
			key = task.Name
		}
		if key == "" {
			return nil, fmt.Errorf("task id or name is required")
		}
		if _, exists := byID[key]; exists {
			return nil, fmt.Errorf("duplicate task %q", key)
		}
		byID[key] = task
	}

	ordered := make([]domain.Task, 0, len(tasks))
	temporary := make(map[string]bool, len(tasks))
	permanent := make(map[string]bool, len(tasks))

	var visit func(string) error
	visit = func(id string) error {
		if permanent[id] {
			return nil
		}
		if temporary[id] {
			return fmt.Errorf("circular task dependency involving %q", id)
		}
		task, ok := byID[id]
		if !ok {
			return fmt.Errorf("unknown task dependency %q", id)
		}
		temporary[id] = true
		for _, dep := range task.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		temporary[id] = false
		permanent[id] = true
		ordered = append(ordered, task)
		return nil
	}

	for _, task := range tasks {
		id := task.ID
		if id == "" {
			id = task.Name
		}
		if err := visit(id); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}
