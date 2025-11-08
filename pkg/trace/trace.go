package trace

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

const RequestIdKey = "X-Request-ID"

// TracerFromContext 从上下文获取 tracer
func TracerFromContext(ctx context.Context) trace.Tracer {
	// 返回一个默认的 tracer
	return trace.NewNoopTracerProvider().Tracer("default")
}

func SpanIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}

	return ""
}

func TraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}
