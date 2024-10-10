package runtime

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func initTracing(enabled bool) (trace.Tracer, *sdktrace.TracerProvider, error) {
	var provider *sdktrace.TracerProvider

	if !enabled {
		// We set a no-op exporter so that we still generate trace ID's as they are a key part
		// of how events / audit work
		provider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(&NoopExporter{}),
		)
	} else {
		// If enabled we export trace data to the default endpoint over grpc. In AWS deployments
		// this will send trace data to the OTEL collector
		exporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}

		provider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
		)
	}

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return otel.Tracer("keel.xyz"), provider, nil
}

type NoopExporter struct {
}

func (n *NoopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (n *NoopExporter) Shutdown(ctx context.Context) error {
	return nil
}
