package flows

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
)

type orchestratorContextKey string

var contextKey orchestratorContextKey = "flowOrchestrator"

// WithOrchestrator sets the given orchestrator in the context
func WithOrchestrator(ctx context.Context, o *Orchestrator) (context.Context, error) {
	return context.WithValue(ctx, contextKey, o), nil
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

func WithDirectInvocation() OrchestratorOpt {
	return func(o *Orchestrator) {
		o.directInvocation = true
	}
}

type Orchestrator struct {
	schema           *proto.Schema
	directInvocation bool // if this orchestrator should directly invoke the flows runtime
}

func NewOrchestrator(ctx context.Context, s *proto.Schema, opts ...OrchestratorOpt) *Orchestrator {
	o := &Orchestrator{
		schema: s,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// orchestrateRun will decide based on the db state if the flow should be ran or not
func (o *Orchestrator) orchestrateRun(ctx context.Context, runID string) error {
	run, err := GetFlowRun(ctx, runID)
	if err != nil {
		return err
	}

	switch run.Status {
	case StatusNew:
		// this is a new run, set it to running and trigger the flows runtime
		run, err := UpdateRun(ctx, run.ID, StatusRunning)
		if err != nil {
			return err
		}

		flow := o.schema.FindFlow(run.Name)
		if flow == nil {
			return fmt.Errorf("invalid flow run")
		}

		// call the flow runtime to execute this flow run
		if o.directInvocation {
			if err = functions.CallFlow(
				ctx,
				flow,
				*run.Input,
			); err != nil {
				return err
			}
		}

		// TODO: handle sqs messages
	case StatusFailed, StatusCompleted, StatusRunning, StatusWaiting:
		// Do nothing
		return nil
	}

	return nil
}
