package runtime

import (
	"context"

	"github.com/teamkeel/keel/runtime"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type RunFlowPayload struct {
	ID string `json:"id"`
	// The name of the flow to run e.g. MySpecialFlow
	Name string `json:"name"`

	Inputs map[string]any `json:"inputs"`
}

func (h *Handler) FlowHandler(ctx context.Context, payload *RunFlowPayload) error {
	defer func() {
		if h.tracerProvider != nil {
			h.tracerProvider.ForceFlush(ctx)
		}
	}()

	ctx, span := h.tracer.Start(ctx, payload.Name)
	defer span.End()

	ctx, err := h.buildContext(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	err = runtime.NewFlowHandler(h.schema).RunFlow(ctx, payload.Name, payload.Inputs)
	if err != nil {
		return err
	}

	return nil
}
