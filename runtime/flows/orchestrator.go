package flows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type orchestratorContextKey string

var contextKey orchestratorContextKey = "flowOrchestrator"

// WithOrchestrator sets the given orchestrator in the context
func WithOrchestrator(ctx context.Context, o *Orchestrator) context.Context {
	return context.WithValue(ctx, contextKey, o)
}

// HasOrchestrator checks the given context if a flow orchestrator has been set
func HasOrchestrator(ctx context.Context) bool {
	return ctx.Value(contextKey) != nil
}

// GetOrchestrator retrieves the flow orchestrator from the context
func GetOrchestrator(ctx context.Context) (*Orchestrator, error) {
	v, ok := ctx.Value(contextKey).(*Orchestrator)
	if !ok {
		return nil, fmt.Errorf("context does not have key or is not Orchestrator: %s", contextKey)
	}
	return v, nil
}

type OrchestratorOpt func(o *Orchestrator)

func WithAsyncQueue(queueURL string, client *sqs.Client) OrchestratorOpt {
	return func(o *Orchestrator) {
		o.sqsClient = client
		o.sqsQueueURL = queueURL
	}
}

type Orchestrator struct {
	schema *proto.Schema
	// Client for sqs messages sent to the flows runtime
	sqsClient *sqs.Client
	// The Flows runtime queue used to trigger the execution of a flow
	sqsQueueURL string
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
	Config       *JSONB `json:"config"`
	UI           *JSONB `json:"ui"` // UI component for the current step, if applicable
}

// orchestrateRun will decide based on the db state if the flow should be ran or not
func (o *Orchestrator) orchestrateRun(ctx context.Context, runID string, data map[string]any) error {
	run, err := getRun(ctx, runID)
	if err != nil {
		return err
	}
	if run == nil {
		return fmt.Errorf("invalid run ID: %s", runID)
	}

	switch run.Status {
	case StatusNew, StatusRunning:
		if run.Status == StatusNew {
			// this is a new run, set it to running and trigger the flows runtime
			run, err = updateRun(ctx, run.ID, StatusRunning)
			if err != nil {
				return err
			}
		}

		// call the flow runtime
		resp, err := o.CallFlow(ctx, run, data)
		if err != nil {
			_, _ = updateRun(ctx, run.ID, StatusFailed)

			return err
		}

		if resp.RunCompleted {
			_, err = updateRun(ctx, run.ID, StatusCompleted)
			return err
		}

		// reload state from db
		run, err := getRun(ctx, run.ID)
		if err != nil {
			return err
		}

		stepsMap := map[string][]Step{}
		for _, step := range run.Steps {
			stepsMap[step.Name] = append(stepsMap[step.Name], step)
		}

		// Check to see if we're in a Pending UI step, break orchestration
		if run.PendingUI() {
			return nil
		}

		// Continue running the flow
		payload := FlowRunUpdated{RunID: resp.RunID, Data: data}
		wrap, err := payload.Wrap()
		if err != nil {
			return err
		}

		return o.SendEvent(ctx, wrap)
	case StatusFailed, StatusCompleted, StatusCancelled:
		// Do nothing
		return nil
	}

	return nil
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

		return o.orchestrateRun(ctx, ev.RunID, ev.Data)
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

		traceID := util.ParseTraceparent(event.Traceparent).TraceID().String()
		run, err := createRun(ctx, flow, ev.Inputs, traceID)
		if err != nil {
			return err
		}
		return o.orchestrateRun(ctx, run.ID, nil)
	}
	return nil
}

// SendEvent sends the given event to the flow runtime's queue or directly invokes the function depending on the
// orchestrator's settings
func (o *Orchestrator) SendEvent(ctx context.Context, payload *EventWrapper) error {
	if payload == nil {
		return fmt.Errorf("invalid event payload")
	}

	// get the traceparent from context, and pass it through to the event if applicable
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	if traceparent := util.GetTraceparent(spanContext); traceparent != "" {
		payload.Traceparent = traceparent
	}

	// if a sqs queue hasn't been set, we continue executing
	if o.sqsClient == nil || o.sqsQueueURL == "" {
		go o.HandleEvent(ctx, payload) //nolint we're "simulating" an async queue
		return nil
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(bodyBytes)),
		QueueUrl:    aws.String(o.sqsQueueURL),
	}

	_, err = o.sqsClient.SendMessage(ctx, input)
	return err
}

// CallFlow is a helper function to call the flows runtime and retrieve the Response Payload
func (o *Orchestrator) CallFlow(ctx context.Context, run *Run, data map[string]any) (*FunctionsResponsePayload, error) {
	ctx, span := tracer.Start(ctx, "CallFlow")
	defer span.End()

	if run == nil {
		return nil, fmt.Errorf("invalid run")
	}

	flow := o.schema.FindFlow(run.Name)
	if flow == nil {
		return nil, fmt.Errorf("invalid flow run")
	}

	resp, meta, err := functions.CallFlow(
		ctx,
		flow,
		run.ID,
		data,
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
