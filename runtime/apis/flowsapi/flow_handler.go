package flowsapi

import (
	"net/http"
	"path"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func FlowHandler(s *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "FlowsAPI")
		defer span.End()

		span.SetAttributes(
			attribute.String("api.protocol", "HTTP JSON"),
		)

		identity, err := actions.HandleAuthorizationHeader(ctx, s, r.Header)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}
		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
		}

		path := path.Clean(r.URL.EscapedPath())
		pathParts := strings.Split(strings.TrimPrefix(path, "/flows/json/"), "/")

		flow := s.FindFlow(pathParts[0])
		if flow == nil {
			return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
		}

		// authorise that the user is allowed to access this flow
		authorised, err := flows.AuthoriseFlow(ctx, s, flow)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}
		if !authorised {
			return httpjson.NewErrorResponse(ctx, common.NewPermissionError(), nil)
		}

		switch len(pathParts) {
		case 1:
			switch r.Method {
			case http.MethodPost:
				// Start flow - POST flows/json/[flowName]
				inputs, err := common.ParseRequestData(r)
				if err != nil {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
				}

				inputsMap, ok := inputs.(map[string]any)
				if !ok && inputs != nil {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
				}

				run, err := flows.StartFlow(ctx, flow, inputsMap)
				if err != nil {
					return httpjson.NewErrorResponse(ctx, err, nil)
				}

				return common.NewJsonResponse(http.StatusOK, run, nil)
			case http.MethodGet:
				// List Flow runs - GET flows/json/[flowName]
				runs, err := flows.ListFlowRuns(ctx, flow, common.ParseQueryParams(r))
				if err != nil {
					return httpjson.NewErrorResponse(ctx, err, nil)
				}

				return common.NewJsonResponse(http.StatusOK, runs, nil)
			}

			return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP POST or GET accepted"), nil)
		case 2:
			// Check for progress - GET flows/json/[flowName]/[runID]
			if r.Method != http.MethodGet {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
			}

			run, err := flows.GetFlowRunState(ctx, pathParts[1])
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}
			if run == nil || run.Name != flow.GetName() {
				return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
			}

			return common.NewJsonResponse(http.StatusOK, run, nil)
		case 3:
			// we're operating on a flow run (cancel/put values), we now need to set the tracing span context to the flow's trace
			if traceparent, err := flows.GetTraceparent(ctx, pathParts[1]); err == nil {
				sc := util.ParseTraceparent(traceparent)
				ctx = trace.ContextWithSpanContext(ctx, sc)
			}

			if pathParts[2] == "cancel" {
				// Cancel run: POST flows/json/[flowName]/[runID]/cancel
				if r.Method != http.MethodPost {
					return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP POST accepted"), nil)
				}

				run, err := flows.CancelFlowRun(ctx, pathParts[1])
				if err != nil {
					return httpjson.NewErrorResponse(ctx, err, nil)
				}

				if run == nil {
					return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
				}

				return common.NewJsonResponse(http.StatusOK, run, nil)
			}

			// Send step updates: PUT flows/json/[flowName]/[runID]/[stepID]?[action=xxxx]
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

			run, err := flows.UpdateStep(ctx, pathParts[1], pathParts[2], data, r.URL.Query().Get("action"))
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}

			return common.NewJsonResponse(http.StatusOK, run, nil)
		case 4:
			// UI Callback: POST flows/json/[flowName]/[runID]/[stepID]/callback?name={callbackName}}&element={elementName}

			// we're operating on a flow run (cancel/put values), we now need to set the tracing span context to the flow's trace
			if traceparent, err := flows.GetTraceparent(ctx, pathParts[1]); err == nil {
				sc := util.ParseTraceparent(traceparent)
				ctx = trace.ContextWithSpanContext(ctx, sc)
			}

			if r.Method != http.MethodPost {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP POST accepted"), nil)
			}

			inputs, err := common.ParseRequestData(r)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
			}

			data, ok := inputs.(map[string]any)
			if !ok {
				return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
			}

			response, err := flows.Callback(ctx, pathParts[1], pathParts[2], data, r.URL.Query().Get("element"), r.URL.Query().Get("callback"))
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}

			return common.NewJsonResponse(http.StatusOK, response, nil)
		}
		return common.Response{
			Status: http.StatusNotFound,
			Body:   []byte("Not found"),
		}
	}
}
