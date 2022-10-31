package runtime

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const tracerName = "graphql"

type Tracer struct{}

func (t *Tracer) Init(ctx context.Context, p *graphql.Params) context.Context {
	return ctx
}

func (t *Tracer) Name() string {
	return "OpenTelemetry"
}

func (t *Tracer) HasResult() bool {
	return false
}

func (t *Tracer) GetResult(ctx context.Context) interface{} {
	return nil
}

func (t *Tracer) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "parse")
	defer span.End()

	if !span.IsRecording() {
		return ctx, func(err error) {}
	}

	return ctx, func(err error) {
		span.SetAttributes(
			attribute.String("error", fmt.Sprintf("%s", err)),
		)
	}
}

func (t *Tracer) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "validation")
	defer span.End()

	if !span.IsRecording() {
		return ctx, func(errs []gqlerrors.FormattedError) {}
	}

	return ctx, func([]gqlerrors.FormattedError) {}
}

func (t *Tracer) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "execution")
	defer span.End()

	if !span.IsRecording() {
		return ctx, func(res *graphql.Result) {}
	}

	return ctx, func(*graphql.Result) {}
}

func (t *Tracer) ResolveFieldDidStart(ctx context.Context, i *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, i.FieldName)
	defer span.End()

	if !span.IsRecording() {
		return ctx, func(interface{}, error) {}
	}

	span.SetAttributes(
		attribute.String("service.name", "graphql"),
		attribute.String("fieldName", i.FieldName),
		attribute.String("parent", i.ParentType.String()),
		attribute.String("query", string(i.Operation.GetLoc().Source.Body)),
	)
	return ctx, func(v interface{}, err error) {
		span.SetAttributes(
			attribute.String("error", fmt.Sprintf("%s", err)),
		)
	}
}
