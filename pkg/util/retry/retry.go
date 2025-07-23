package retry

import (
	"context"
	"errors"
)

// Retry is a util function to retry execute function
func Retry(ctx context.Context, fn func() error, retryTimes int) error {
	var err error
	for i := 0; i < retryTimes; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}
	return err
}
