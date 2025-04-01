package flows

import (
	"context"
	"fmt"
)

// RunFlow runs the given flow run using the flow orchestrator from the context
func RunFlow(ctx context.Context, runID string) error {
	if !HasOrchestrator(ctx) {
		return fmt.Errorf("flow orchestrator not set")
	}

	o, err := GetOrchestrator(ctx)
	if err != nil {
		return err
	}

	return o.orchestrateRun(ctx, runID)
}
