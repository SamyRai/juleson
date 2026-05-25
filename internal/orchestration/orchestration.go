package orchestration

import "github.com/SamyRai/juleson/internal/orchestration/app"

type Runtime = app.Runtime
type RuntimeDeps = app.RuntimeDeps

func NewRuntime(deps RuntimeDeps) *Runtime {
	return app.NewRuntime(deps)
}
