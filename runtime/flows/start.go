package flows

import (
	"fmt"

	"github.com/teamkeel/keel/db"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Start will start the scope's flow with the given input
func Start(scope *Scope, inputs any) (*FlowRun, error) {
	if scope.Flow == nil {
		return nil, fmt.Errorf("invalid flow")
	}
	ctx, span := tracer.Start(scope.Context, "StartFlow")
	defer span.End()

	span.SetAttributes(
		attribute.String("flow", scope.Flow.Name),
	)

	scope = scope.WithContext(ctx)
	var jsonInputs JSONB
	if inputsMap, ok := inputs.(map[string]any); ok {
		jsonInputs = inputsMap
	}

	run := FlowRun{
		Status: StatusNew,
		Input:  &jsonInputs,
		Name:   scope.Flow.Name,
	}

	database, err := db.GetDatabase(scope.Context)
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
