package obs

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// FieldsFromContext extracts trace and span IDs from context as zap fields.
func FieldsFromContext(ctx context.Context) []zap.Field {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	return []zap.Field{
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	}
}

// WithContext adds trace context fields to a logger.
func WithContext(ctx context.Context, l *zap.Logger) *zap.Logger {

	if l == nil {
		return l
	}
	fields := FieldsFromContext(ctx)
	if len(fields) == 0 {
		return l
	}

	return l.With(fields...)
}
