package build

import "context"

type ModuleManager struct{}

func NewModuleManager() *ModuleManager {
	return &ModuleManager{}
}

func (m *ModuleManager) Tidy(ctx context.Context) error {
	return runGo(ctx, "mod", "tidy")
}

func (m *ModuleManager) Download(ctx context.Context) error {
	return runGo(ctx, "mod", "download")
}

func (m *ModuleManager) Verify(ctx context.Context) error {
	return runGo(ctx, "mod", "verify")
}

func (m *ModuleManager) Vendor(ctx context.Context) error {
	return runGo(ctx, "mod", "vendor")
}

func (m *ModuleManager) Graph(ctx context.Context) error {
	return runGo(ctx, "mod", "graph")
}

func (m *ModuleManager) Why(ctx context.Context, packages ...string) error {
	args := append([]string{"mod", "why"}, packages...)
	return runGo(ctx, args...)
}
