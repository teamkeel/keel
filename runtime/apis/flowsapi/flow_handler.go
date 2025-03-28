package flowsapi

import (
	"net/http"
	"slices"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/flows"
	"go.opentelemetry.io/otel/attribute"
)

func FlowHandler(s *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "FlowsAPI")
		defer span.End()
		span.SetAttributes(
			attribute.String("api.protocol", "HTTP JSON"),
		)
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/flows/json/"), "/")
		flowName := pathParts[0]

		if !slices.ContainsFunc(s.FlowNames(), func(f string) bool {
			return strings.ToLower(f) == flowName
		}) {
			return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
		}
		scope := flows.NewScope(s.FindFlow(flowName), s)

		switch len(pathParts) {
		case 1:
			// Start flow - POST flows/json/[flowName]
			if r.Method != http.MethodPost {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP POST accepted"), nil)
			}

			inputs, err := common.ParseRequestData(r)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
			}

			run, err := flows.StartRun(ctx, scope, inputs)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}

			return common.NewJsonResponse(http.StatusOK, run, nil)
		case 2:
			// Check for progress - GET flows/json/[flowName]/[runID]
			if r.Method != http.MethodGet {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
			}

			run, err := flows.GetFlowRun(ctx, scope, pathParts[1])
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}
			if run == nil {
				return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
			}

			return common.NewJsonResponse(http.StatusOK, run, nil)
		case 3:
			// TODO: # Send step updates
			// PUT flows/json/[flowName]/[runID]/[stepID]

			// TODO: # Cancel run
			// POST flows/json/[flowName]/[runID]/cancel
			return common.NewJsonResponse(http.StatusNotImplemented, pathParts, nil)
		}
		return common.Response{
			Status: http.StatusNotFound,
			Body:   []byte("Not found"),
		}
	}
}
