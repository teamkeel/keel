package flows

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartFlow will start a new run for the given flow with the given input.
func StartFlow(ctx context.Context, flow *proto.Flow, inputs map[string]any) (run *Run, err error) {
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

	var identityID *string

	if identity, err := auth.GetIdentity(ctx); err == nil {
		idenID := identity[parser.FieldNameId].(string)
		if idenID != "" {
			identityID = &idenID
		}
	}

	run, err = createRun(ctx, flow, inputs, util.GetTraceparent(span.SpanContext()), identityID)
	if err != nil {
		err = fmt.Errorf("creating flow run: %w", err)
		return
	}

	span.SetAttributes(
		attribute.String("flow", flow.GetName()),
		attribute.String("flowRun.id", run.ID),
	)

	err, uiComponents := o.orchestrateRun(ctx, run.ID, inputs, nil, "")
	if err != nil {
		err = fmt.Errorf("orchestrating flow run: %w", err)
		return
	}

	// load fresh state
	run, err = getRun(ctx, run.ID)
	if err != nil {
		err = fmt.Errorf("retrieving flow run: %w", err)
		return
	}

	run.SetUIComponents(uiComponents)

	return run, nil
}

// ListFlowRuns will return the runs for the given flow; with pagination.
func ListFlowRuns(ctx context.Context, flow *proto.Flow, pageInputs map[string]any) (runs []*Run, err error) {
	ctx, span := tracer.Start(ctx, "ListFlowRuns")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	pf := paginationFields{}
	pf.Parse(pageInputs)

	runs, err = listRuns(ctx, &filterFields{FlowName: &flow.Name}, &pf)
	return
}

// ListUserFlowRuns will return the runs initiated by the given identity ID.
func ListUserFlowRuns(ctx context.Context, identityID string, inputs map[string]any) (runs []*Run, err error) {
	ctx, span := tracer.Start(ctx, "ListUserFlowRuns")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	pf := paginationFields{}
	pf.Parse(inputs)

	ff := filterFields{}
	ff.Parse(inputs)
	ff.StartedBy = &identityID

	runs, err = listRuns(ctx, &ff, &pf)
	return
}

// GetFlowRunState retrieves the state of the given flow run. If the run has a pending UI step, the UI component will be
// injected into the step before returning it.
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

	// If no run found, return
	if run == nil {
		return
	}

	span.SetAttributes(
		attribute.String("flowRun.id", run.ID),
		attribute.String("flowRun.status", string(run.Status)),
	)

	// if we're not waiting for a UI step, return
	if !run.HasPendingUIStep() && !run.HasCompleteStep() {
		return
	}

	var o *Orchestrator
	o, err = GetOrchestrator(ctx)
	if err != nil {
		err = fmt.Errorf("retrieving context flow orchestrator: %w", err)
		return
	}

	// retrieving the step component from the flow runtime
	resp, err := o.CallFlow(ctx, run, run.Input.(map[string]any), nil, "")
	if err != nil {
		err = fmt.Errorf("retrieving ui component: %w", err)
		return
	}

	// setting the flow config and ui component on the pending UI step
	run.SetUIComponents(resp.GetUIComponents())

	return
}

// CancelFlowRun cancels the run with the given ID.
func CancelFlowRun(ctx context.Context, runID string) (run *Run, err error) {
	ctx, span := tracer.Start(ctx, "CancelFlowRun")
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

	if run == nil {
		// return nil run as it's not found
		return
	}

	span.SetAttributes(
		attribute.String("flowRun.id", run.ID),
	)

	// if the run cannot be cancelled, just return
	if run.Status != StatusNew && run.Status != StatusRunning {
		return
	}

	run, err = updateRun(ctx, run.ID, StatusCancelled, nil)
	if err != nil {
		err = fmt.Errorf("updating flow run: %w", err)
		return
	}

	// return fresh state
	run, err = getRun(ctx, runID)
	return
}

// UpdateStep sets the given input on the given pending UI step, updating it's status to COMPLETED. It then returs the
// updated run state.
func UpdateStep(ctx context.Context, runID string, stepID string, data map[string]any, action string) (run *Run, err error) {
	ctx, span := tracer.Start(ctx, "UpdateStep")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	run, err = getRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	if run == nil {
		return nil, fmt.Errorf("flow run not found")
	}

	var o *Orchestrator
	o, err = GetOrchestrator(ctx)
	if err != nil {
		err = fmt.Errorf("retrieving context flow orchestrator: %w", err)
		return
	}

	// Run the flow synchronously
	err, uiComponents := o.orchestrateRun(ctx, runID, run.Input.(map[string]any), data, action)
	if err != nil {
		err = fmt.Errorf("orchestrating flow run: %w", err)
		return
	}

	// return fresh state
	run, err = getRun(ctx, runID)
	if err == nil {
		// set the config & any ui components
		run.SetUIComponents(uiComponents)
	}

	return
}
