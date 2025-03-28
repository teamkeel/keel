package flows

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/db"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartRun will start the scope's flow with the given input
func StartRun(ctx context.Context, scope *Scope, inputs any) (*Run, error) {
	if scope.Flow == nil {
		return nil, fmt.Errorf("invalid flow")
	}
	ctx, span := tracer.Start(ctx, "StartRun")
	defer span.End()

	span.SetAttributes(
		attribute.String("flow", scope.Flow.Name),
	)

	var jsonInputs JSONB
	if inputsMap, ok := inputs.(map[string]any); ok {
		jsonInputs = inputsMap
	}

	run := Run{
		Status: StatusNew,
		Input:  &jsonInputs,
		Name:   scope.Flow.Name,
	}

	database, err := db.GetDatabase(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	result := database.GetDB().Create(&run)
	if result.Error != nil {
		span.RecordError(result.Error, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, result.Error.Error())
		return nil, result.Error
	}

	span.SetAttributes(attribute.String("flowRunID", run.ID))
	return &run, nil
}
