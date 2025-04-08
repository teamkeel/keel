package flows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
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
}

// orchestrateRun will decide based on the db state if the flow should be ran or not
func (o *Orchestrator) orchestrateRun(ctx context.Context, runID string) error {
	run, err := GetFlowRun(ctx, runID)
	if err != nil {
		return err
	}
	if run == nil {
		return fmt.Errorf("invalid run ID: %s", runID)
	}

	flow := o.schema.FindFlow(run.Name)
	if flow == nil {
		return fmt.Errorf("invalid flow run")
	}

	switch run.Status {
	case StatusNew, StatusRunning:
		if run.Status == StatusNew {
			// this is a new run, set it to running and trigger the flows runtime
			run, err = UpdateRun(ctx, run.ID, StatusRunning)
			if err != nil {
				return err
			}
		}

		resp, _, err := functions.CallFlow(
			ctx,
			flow,
			run.ID,
		)
		if err != nil {
			return err
		}

		b, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		var respBody FunctionsResponsePayload
		if err := json.Unmarshal(b, &respBody); err != nil {
			return err
		}

		if respBody.RunCompleted {
			_, err = UpdateRun(ctx, run.ID, StatusCompleted)
			return err
		}

		//if !respBody.RunCompleted {
		stepsMap := map[string][]Step{}
		for _, step := range run.Steps {
			stepsMap[step.Name] = append(stepsMap[step.Name], step)
		}

		// Check to see if the retries have been exceeded
		if len(run.Steps) > 0 {
			lastStep := run.Steps[len(run.Steps)-1]
			if lastStep.Status == "FAILED" && len(stepsMap[lastStep.Name]) >= lastStep.MaxRetries {
				_, err := UpdateRun(ctx, run.ID, StatusFailed)
				return err
			}
		}
		// } else {
		// 	_, err = UpdateRun(ctx, run.ID, StatusCompleted)
		// 	return err
		// }

		payload := FlowRunUpdated{RunID: respBody.RunID}
		wrap, err := payload.Wrap()
		if err != nil {
			return err
		}

		return o.sendEvent(ctx, wrap)
	case StatusFailed, StatusCompleted:
		// Do nothing
		return nil
	case StatusWaiting:
		return fmt.Errorf("not implemented")
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

		return o.orchestrateRun(ctx, ev.RunID)
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
		run, err := CreateRun(ctx, flow, ev.Inputs)
		if err != nil {
			return err
		}
		return o.orchestrateRun(ctx, run.ID)
	}
	return nil
}

// sendEvent sends the given event to the flow runtime's queue or directly invokes the function depending on the
// orchestrator's settings
func (o *Orchestrator) sendEvent(ctx context.Context, payload *EventWrapper) error {
	// if a sqs queue hasn't been set, we continue executing
	if o.sqsClient == nil || o.sqsQueueURL == "" {
		return o.HandleEvent(ctx, payload)
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
