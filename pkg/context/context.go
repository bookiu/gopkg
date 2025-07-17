package context

import (
	stdctx "context"

	"github.com/bookiu/gopkg/pkg/infra/logger"
	"go.uber.org/zap"
)

type loggerKeyType struct{}

var loggerKey loggerKeyType

func WithLogger(ctx stdctx.Context, logger *zap.Logger) stdctx.Context {
	return stdctx.WithValue(ctx, loggerKey, logger)
}

func GetLogger(ctx stdctx.Context) *zap.Logger {
	l, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok {
		return logger.GetLogger()
	}
	return l
}
