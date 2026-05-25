package adapters

import (
	"context"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

func (SystemClock) Sleep(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type NoopProgressSink struct{}

func (NoopProgressSink) ReportProgress(ctx context.Context, progress domain.Progress) error {
	return nil
}
