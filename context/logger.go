package context

import (
	stdctx "context"

	"github.com/bookiu/gopkg/infra/logger"
	"go.uber.org/zap"
)

type loggerKeyType struct{}

var loggerKey loggerKeyType

func WithLogger(ctx stdctx.Context, logger *zap.Logger) stdctx.Context {
	return stdctx.WithValue(ctx, loggerKey, logger)
}

func LoggerWithFields(ctx stdctx.Context, fields ...zap.Field) stdctx.Context {
	l := GetLogger(ctx)
	nl := l.With(fields...)
	return WithLogger(ctx, nl)
}

func GetLogger(ctx stdctx.Context) *zap.Logger {
	l, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok {
		return logger.GetLogger()
	}
	return l
}
