package retry

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Func func(ctx context.Context) error

func Retry(f Func, retries int) Func {
	return func(ctx context.Context) error {
		for r := 0; ; r++ {
			err := f(ctx)
			if err == nil || r >= retries {
				return err
			}

			// Exponential increase in latency.
			shouldRetryAt := time.Second * 2 << r //nolint:mnd
			slog.Warn(fmt.Sprintf("Attempt %d failed; retrying in %v", r+1, shouldRetryAt))

			select {
			case <-time.After(shouldRetryAt):
			case <-ctx.Done():
				return fmt.Errorf("retry: %w", ctx.Err())
			}
		}
	}
}
