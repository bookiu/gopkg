package retry

import (
	"context"
	"errors"
)

// Retry is a util function to retry execute function
func Retry(ctx context.Context, fn func() error, retryTimes int) (error, int) {
	var err error
	times := 0
	for i := 0; i < retryTimes; i++ {
		if ctx.Err() != nil {
			return ctx.Err(), times
		}
		times++
		err = fn()
		if err == nil {
			return nil, times
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err, times
		}
	}
	return err, times
}
