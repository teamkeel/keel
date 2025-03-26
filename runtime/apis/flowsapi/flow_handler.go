package flowsapi

import (
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"go.opentelemetry.io/otel/attribute"
)

func FlowHandler(p *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		_, span := tracer.Start(r.Context(), "Flow")
		defer span.End()
		span.SetAttributes(
			attribute.String("api.protocol", "HTTP JSON"),
		)

		//TODO: implement handling of:
		// # Start flow
		// POST flows/json/[flowName]

		// # Check for progress
		// GET flows/json/[flowName]/[runID]

		// # Send step updates
		// PUT flows/json/[flowName]/[runID]/[stepID]

		// # Cancel run
		// POST flows/json/[flowName]/[runID]/cancel

		// # Pending runs started by me
		// GET flows/json/myRuns

		return common.NewJsonResponse(http.StatusNotImplemented, "TODO", nil)
	}
}
