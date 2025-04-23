package flows

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartFlow will start a new run for the given flow with the given input
func StartFlow(ctx context.Context, flow *proto.Flow, inputs any) (run *Run, err error) {
	ctx, span := tracer.Start(ctx, "StartFlow")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	var o *Orchestrator
	o, err = GetOrchestrator(ctx)
	if err != nil {
		err = fmt.Errorf("retrieving context flow orchestrator: %w", err)
		return
	}

	run, err = createRun(ctx, flow, inputs)
	if err != nil {
		err = fmt.Errorf("creating flow run: %w", err)
		return
	}

	span.SetAttributes(
		attribute.String("flow", flow.Name),
		attribute.String("flowRun.id", run.ID),
	)

	if err = o.orchestrateRun(ctx, run.ID, nil); err != nil {
		err = fmt.Errorf("orchestrating flow run: %w", err)
		return
	}

	// load fresh state
	run, err = getRun(ctx, run.ID)
	if err != nil {
		err = fmt.Errorf("retrieving flow run: %w", err)
		return
	}

	return run, nil
}

// GetFlowRunState retrieves the state of the given flow run. If the run has a pending UI step, the UI component will be
// injected into the step before returning it
func GetFlowRunState(ctx context.Context, runID string) (run *Run, err error) {
	ctx, span := tracer.Start(ctx, "GetFlowRunState")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	run, err = getRun(ctx, runID)
	if err != nil {
		err = fmt.Errorf("retrieving flow run: %w", err)
		return
	}

	span.SetAttributes(
		attribute.String("flowRun.id", run.ID),
		attribute.String("flowRun.status", string(run.Status)),
	)

	// if we're not waiting for a UI step, return
	if !run.PendingUI() {
		return
	}

	var o *Orchestrator
	o, err = GetOrchestrator(ctx)
	if err != nil {
		err = fmt.Errorf("retrieving context flow orchestrator: %w", err)
		return
	}

	// retrieving the step component from the flow runtime
	resp, err := o.CallFlow(ctx, run, nil)
	if err != nil {
		err = fmt.Errorf("retrieving ui component: %w", err)
		return
	}

	// setting the ui component on the pending UI step
	run.SetUIComponent(resp.UI)
	return
}

// UpdateStep sets the given input on the given pending UI step, updating it's status to COMPLETED. It then returs the
// updated run state
func UpdateStep(ctx context.Context, runID string, stepID string, inputs map[string]any) (run *Run, err error) {
	ctx, span := tracer.Start(ctx, "UpdateStep")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	// trigger the orchestrator to continue running the flow
	var o *Orchestrator
	o, err = GetOrchestrator(ctx)
	if err != nil {
		err = fmt.Errorf("retrieving context flow orchestrator: %w", err)
		return
	}

	payload := FlowRunUpdated{RunID: runID, Data: inputs}
	wrap, err := payload.Wrap()
	if err != nil {
		err = fmt.Errorf("creating FlowRunUpdated event: %w", err)
		return
	}

	err = o.SendEvent(ctx, wrap)
	if err != nil {
		err = fmt.Errorf("sending FlowRunUpdated event: %w", err)
		return
	}

	// return the new run state
	return GetFlowRunState(ctx, runID)
}
