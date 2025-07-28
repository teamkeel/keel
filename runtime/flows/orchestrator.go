package flows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type orchestratorContextKey string

var contextKey orchestratorContextKey = "flowOrchestrator"

// WithOrchestrator sets the given orchestrator in the context.
func WithOrchestrator(ctx context.Context, o *Orchestrator) context.Context {
	return context.WithValue(ctx, contextKey, o)
}

// HasOrchestrator checks the given context if a flow orchestrator has been set.
func HasOrchestrator(ctx context.Context) bool {
	return ctx.Value(contextKey) != nil
}

// GetOrchestrator retrieves the flow orchestrator from the context.
func GetOrchestrator(ctx context.Context) (*Orchestrator, error) {
	v, ok := ctx.Value(contextKey).(*Orchestrator)
	if !ok {
		return nil, fmt.Errorf("context does not have key or is not Orchestrator: %s", contextKey)
	}
	return v, nil
}

type OrchestratorOpt func(o *Orchestrator)

// WithEventSender initialises the orchestrator with the given EventSender.
func WithEventSender(es EventSender) OrchestratorOpt {
	return func(o *Orchestrator) {
		o.eventSender = es
	}
}

// WithAsyncQueue sets a SQSEventSender on the orchestrator with the given options (queue URL and sqs Client).
func WithAsyncQueue(queueURL string, sqsClient *sqs.Client) OrchestratorOpt {
	es := NewSQSEventSender(queueURL, sqsClient)
	return WithEventSender(es)
}

// WithNoQueueEventSender initialises the orchestrator with a simulated async queue.
func WithNoQueueEventSender() OrchestratorOpt {
	return func(o *Orchestrator) {
		o.eventSender = NewNoQueueEventSender(o)
	}
}

type Orchestrator struct {
	schema      *proto.Schema
	eventSender EventSender
}

func NewOrchestrator(s *proto.Schema, opts ...OrchestratorOpt) *Orchestrator {
	o := &Orchestrator{
		schema: s,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type FunctionsResponsePayload struct {
	RunID        string `json:"runId"`
	RunCompleted bool   `json:"runCompleted"`
	Data         JSON   `json:"data"`
	Config       JSON   `json:"config"`
	UI           JSON   `json:"ui"` // UI component for the current step, if applicable
	Error        string `json:"error"`
}

func (r *FunctionsResponsePayload) GetUIComponents() *FlowUIComponents {
	if r.Config != nil || r.UI != nil {
		return &FlowUIComponents{
			Config: r.Config,
			UI:     r.UI,
		}
	}

	return nil
}

// FlowUIComponents contains data returned from the functions runtime which is used for frontend rendering.
type FlowUIComponents struct {
	Config JSON `json:"config"`
	UI     JSON `json:"ui"`
}

// orchestrateRun will decide based on the db state if the flow should be ran or not.
//
// * inputs represents the flow run inputs.
// * data represents the input for the current step execution.
// * action is optional.
func (o *Orchestrator) orchestrateRun(ctx context.Context, runID string, inputs map[string]any, data map[string]any, action string) (error, *FlowUIComponents) {
	run, err := getRun(ctx, runID)
	if err != nil {
		return err, nil
	}
	if run == nil {
		return fmt.Errorf("invalid run ID: %s", runID), nil
	}

	switch run.Status {
	case StatusNew, StatusRunning, StatusAwaitingInput:
		if run.Status == StatusNew {
			// this is a new run, set it to running and trigger the flows runtime
			run, err = updateRun(ctx, run.ID, StatusRunning, run.Config)
			if err != nil {
				return err, nil
			}
		}

		if run.Status == StatusAwaitingInput {
			// we have been awaiting input, we're now going to continue running
			run, err = updateRun(ctx, run.ID, StatusRunning, run.Config)
			if err != nil {
				return err, nil
			}
		}

		// call the flow runtime
		resp, err := o.CallFlow(ctx, run, inputs, data, action)
		if err != nil {
			cfg := JSON(nil)
			if resp != nil {
				cfg = resp.Config
			}
			// failed orchestrating, mark the run as failed and return the error
			_, _ = updateRun(ctx, run.ID, StatusFailed, cfg)
			return err, nil
		}

		if resp.RunCompleted {
			if resp.Error != "" {
				// run was orchestrated and completed successfully, but with an error (e.g. exhaused retries)
				_, err = updateRun(ctx, run.ID, StatusFailed, resp.Config)
				return err, resp.GetUIComponents()
			}

			_, err = completeRun(ctx, run.ID, resp.Config, resp.Data)
			return err, resp.GetUIComponents()
		}

		// reload state from db
		run, err := getRun(ctx, run.ID)
		if err != nil {
			return err, nil
		}

		// Check to see if we're in a Pending UI step, break orchestration
		if run.HasPendingUIStep() {
			_, err = updateRun(ctx, run.ID, StatusAwaitingInput, resp.Config)
			return err, resp.GetUIComponents()
		}

		// Set the config
		_, err = updateRun(ctx, run.ID, run.Status, resp.Config)
		if err != nil {
			return err, nil
		}

		// Continue running the flow
		payload := FlowRunUpdated{RunID: resp.RunID}
		wrap, err := payload.Wrap()
		if err != nil {
			return err, nil
		}

		return o.sendEvent(ctx, wrap), resp.GetUIComponents()
	case StatusFailed, StatusCompleted, StatusCancelled:
		// Do nothing
		return nil, nil
	}

	return nil, nil
}

// HandleEvent will handle an event received by the orchestrator from the flows runtime; The only events handled at the
// moment are FlowRunUpdated.
func (o *Orchestrator) HandleEvent(ctx context.Context, event *EventWrapper) error {
	switch event.EventName {
	case EventNameFlowRunUpdated:
		var ev FlowRunUpdated
		if err := ev.ReadPayload(event); err != nil {
			return err
		}

		run, err := getRun(ctx, ev.RunID)
		if err != nil {
			return err
		}

		err, _ = o.orchestrateRun(ctx, run.ID, run.Input.(map[string]any), ev.Data, ev.Action)

		return err
	case EventNameFlowRunStarted:
		// a new flow has started; create the run and start orchestrating it
		var ev FlowRunStarted
		if err := ev.ReadPayload(event); err != nil {
			return err
		}

		flow := o.schema.FindFlow(ev.Name)
		if flow == nil {
			return fmt.Errorf("unknown flow: %s", ev.Name)
		}

		traceparent := event.Traceparent
		if traceparent == "" {
			var span trace.Span
			ctx, span = tracer.Start(ctx, "StartFlow")
			defer span.End()
			traceparent = util.GetTraceparent(span.SpanContext())
		}

		run, err := createRun(ctx, flow, ev.Inputs, traceparent, nil)
		if err != nil {
			return err
		}
		err, _ = o.orchestrateRun(ctx, run.ID, ev.Inputs, nil, "")

		return err
	}
	return nil
}

// sendEvent sends the given event to the flow runtime's queue or directly invokes the function depending on the
// orchestrator's settings.
func (o *Orchestrator) sendEvent(ctx context.Context, payload *EventWrapper) error {
	if payload == nil {
		return fmt.Errorf("invalid event payload")
	}

	if o.eventSender == nil {
		return fmt.Errorf("no event sender available for orchestrator")
	}

	// get the traceparent from context, and pass it through to the event if applicable
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	if traceparent := util.GetTraceparent(spanContext); traceparent != "" {
		payload.Traceparent = traceparent
	}

	// send event
	return o.eventSender.Send(ctx, payload)
}

// CallFlow is a helper function to call the flows runtime and retrieve the Response Payload.
func (o *Orchestrator) CallFlow(ctx context.Context, run *Run, inputs map[string]any, data map[string]any, action string) (*FunctionsResponsePayload, error) {
	ctx, span := tracer.Start(ctx, "CallFlow")
	defer span.End()

	if run == nil {
		return nil, fmt.Errorf("invalid run")
	}

	flow := o.schema.FindFlow(run.Name)
	if flow == nil {
		return nil, fmt.Errorf("invalid flow run")
	}

	if flow.GetInputMessageName() != "" {
		message := o.schema.FindMessage(flow.GetInputMessageName())
		var err error
		inputs, err = actions.TransformInputs(o.schema, message, inputs, true)
		if err != nil {
			return nil, err
		}
	}

	if !auth.IsAuthenticated(ctx) && run.StartedBy != nil {
		if identity, err := actions.FindIdentityById(ctx, o.schema, *run.StartedBy); err == nil {
			ctx = auth.WithIdentity(ctx, identity)
		}
	}

	resp, meta, err := functions.CallFlow(
		ctx,
		flow,
		run.ID,
		inputs,
		data,
		action,
	)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	var respBody FunctionsResponsePayload
	if err := json.Unmarshal(b, &respBody); err != nil {
		return nil, err
	}

	if meta != nil {
		span.SetAttributes(attribute.Int("response.code", meta.Status))
	}

	if respBody.RunID == "" {
		err := fmt.Errorf("invalid response from flows runtime")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("response.body", string(b)))
		return nil, err
	}

	return &respBody, nil
}
