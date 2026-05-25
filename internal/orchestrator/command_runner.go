package orchestrator

import (
	"context"
	"os/exec"
)

type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
	CombinedOutput(ctx context.Context, name string, args ...string) (string, error)
}

type shellCommandRunner struct {
	service *Service
}

func (r shellCommandRunner) Run(ctx context.Context, name string, args ...string) error {
	return r.service.runCommand(ctx, name, args...)
}

func (r shellCommandRunner) CombinedOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
