package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/otel/trace"
)

func TracerStartSpan(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	ctx = trace.ContextWithSpan(ctx, span)
	return ctx
}

func TracerFinishSpan(ctx context.Context) {
	span := opentracing.SpanFromContext(ctx)
	span.Finish()
}
