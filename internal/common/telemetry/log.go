package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// withTrace prefixes the provided keyvals with trace/span correlation fields when available.
func withTrace(ctx context.Context, kv []any) []any {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return kv
	}
	prefixed := make([]any, 0, len(kv)+4)
	prefixed = append(prefixed, "trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
	prefixed = append(prefixed, kv...)
	return prefixed
}

// Info logs a message using slog, including trace/span correlation fields when available.
func Info(ctx context.Context, msg string, kv ...any) {
	slog.InfoContext(ctx, msg, withTrace(ctx, kv)...)
}

// Error logs an error using slog, including trace/span correlation fields when available.
func Error(ctx context.Context, err error, msg string, kv ...any) {
	attrs := append([]any{"error", err}, kv...)
	slog.ErrorContext(ctx, msg, withTrace(ctx, attrs)...)
}
