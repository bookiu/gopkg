package context

import stdctx "context"

type requestIdKeyType struct{}
type traceIdKeyType struct{}

var (
	requestIdKey requestIdKeyType
	traceIdKey   traceIdKeyType
)

func WithRequestId(ctx stdctx.Context, requestId string) stdctx.Context {
	return stdctx.WithValue(ctx, requestIdKey, requestId)
}

func GetRequestId(ctx stdctx.Context) string {
	l, ok := ctx.Value(requestIdKey).(string)
	if !ok {
		return ""
	}
	return l
}

func WithTraceId(ctx stdctx.Context, traceId string) stdctx.Context {
	return stdctx.WithValue(ctx, traceIdKey, traceId)
}

func GetTraceId(ctx stdctx.Context) string {
	l, ok := ctx.Value(traceIdKey).(string)
	if !ok {
		return ""
	}
	return l
}
