package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tracerProvider *sdktrace.TracerProvider
var tracer = otel.Tracer("keel.xyz")

func initTracing() {
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(&NoopExporter{}),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

type NoopExporter struct {
}

func (n *NoopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (n *NoopExporter) Shutdown(ctx context.Context) error {
	return nil
}
