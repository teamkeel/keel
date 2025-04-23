package flowsapi

import (
	"net/http"
	"path"
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

		path := path.Clean(r.URL.EscapedPath())
		pathParts := strings.Split(strings.TrimPrefix(path, "/flows/json/"), "/")

		flow := s.FindFlow(pathParts[0])
		if flow == nil {
			return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
		}

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

			run, err := flows.StartFlow(ctx, flow, inputs)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}

			return common.NewJsonResponse(http.StatusOK, run, nil)
		case 2:
			// Check for progress - GET flows/json/[flowName]/[runID]
			if r.Method != http.MethodGet {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
			}

			run, err := flows.GetFlowRunState(ctx, pathParts[1])
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}
			if run == nil || run.Name != flow.Name {
				return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
			}

			return common.NewJsonResponse(http.StatusOK, run, nil)
		case 3:
			if pathParts[2] == "cancel" {
				// TODO: # Cancel run
				// POST flows/json/[flowName]/[runID]/cancel
				return common.NewJsonResponse(http.StatusNotImplemented, pathParts, nil)
			}
			// Send step updates: PUT flows/json/[flowName]/[runID]/[stepID]
			if r.Method != http.MethodPut {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP PUT accepted"), nil)
			}

			inputs, err := common.ParseRequestData(r)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
			}

			data, ok := inputs.(map[string]any)
			if !ok {
				return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
			}

			run, err := flows.UpdateStep(ctx, pathParts[1], pathParts[2], data)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}

			return common.NewJsonResponse(http.StatusOK, run, nil)
		}
		return common.Response{
			Status: http.StatusNotFound,
			Body:   []byte("Not found"),
		}
	}
}
