package app

import (
	"context"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

func clockOrDefault(clock ports.Clock) ports.Clock {
	if clock != nil {
		return clock
	}
	return systemClock{}
}

func reportProgress(ctx context.Context, sink ports.ProgressSink, progress domain.Progress) error {
	if sink == nil {
		return nil
	}
	return sink.ReportProgress(ctx, progress)
}
