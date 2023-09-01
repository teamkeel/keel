package util

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

// Retrieves traceparent, or empty string if no span exists.
// Uses standard format: {version}-{trace_id}-{span_id}-{trace_flags}
// See specification: https://www.w3.org/TR/trace-context/#traceparent-header-field-values
func GetTraceparent(ctx context.Context) string {
	span := trace.SpanContextFromContext(ctx)
	if !span.IsValid() {
		return ""
	}

	// Clear all flags other than the trace-context supported sampling bit.
	flags := span.TraceFlags() & trace.FlagsSampled

	return fmt.Sprintf("%s-%s-%s-%s",
		"00",
		span.TraceID(),
		span.SpanID(),
		flags)
}
