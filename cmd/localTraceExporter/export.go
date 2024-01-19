package localTraceExporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
)

// New constructs a new Exporter and starts it.
func New(ctx context.Context) (*otlptrace.Exporter, error) {
	return otlptrace.New(ctx, NewClient())
}
