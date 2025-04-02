package flows

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// CreateRun will start the scope's flow with the given input
func CreateRun(ctx context.Context, flow *proto.Flow, inputs any) (*Run, error) {
	if flow == nil {
		return nil, fmt.Errorf("invalid flow")
	}
	ctx, span := tracer.Start(ctx, "CreateRun")
	defer span.End()

	span.SetAttributes(
		attribute.String("flow", flow.Name),
	)

	var jsonInputs JSONB
	if inputsMap, ok := inputs.(map[string]any); ok {
		jsonInputs = inputsMap
	}

	run := Run{
		Status: StatusNew,
		Input:  &jsonInputs,
		Name:   flow.Name,
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

	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &run, nil
}
