package httpclient

import (
	"context"
	"time"

	pkgctx "github.com/bookiu/gopkg/pkg/context"
	"go.uber.org/zap"
)

// ObserveProvider is an interface for observing HTTP requests.
type ObserveProvider interface {
	RecordRequest(ctx context.Context, method, url string, statusCode int, duration time.Duration, err error)
}

type ObserveRequest struct {
}

func NewObserveRequest() ObserveProvider {
	return &ObserveRequest{}
}

func (o *ObserveRequest) RecordRequest(ctx context.Context, method, url string, statusCode int, duration time.Duration, err error) {
	log := pkgctx.GetLogger(ctx)
	if err != nil {
		log.Error("Request failed",
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return
	}
	log.Info("Rcv response",
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
	)
}

type NoopObserve struct {
}

func (o *NoopObserve) RecordRequest(ctx context.Context, method, url string, statusCode int, duration time.Duration, err error) {
}
