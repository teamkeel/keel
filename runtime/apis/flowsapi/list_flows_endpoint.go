package flowsapi

import (
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/runtime/locale"
	"github.com/teamkeel/keel/runtime/openapi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/flowsapi")

// ListFlowsHandler handles a request to /flows/json and returns data about all flows defined in the schema.
func ListFlowsHandler(p *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "FlowsAPI")
		defer span.End()
		span.SetAttributes(
			attribute.String("api.protocol", "HTTP JSON"),
		)

		identity, err := actions.HandleAuthorizationHeader(ctx, p, r.Header)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}
		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
		}

		// handle any Time-Zone headers
		location, err := locale.HandleTimezoneHeader(ctx, r.Header)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError(err.Error()), nil)
		}
		ctx = locale.WithTimeLocation(ctx, location)

		if r.Method != http.MethodGet {
			return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
		}

		authorisedFlows, err := flows.AuthorisedFlows(ctx, p)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}

		flowsData := []map[string]any{}
		for _, f := range authorisedFlows {
			inputFields := []map[string]any{}
			if inputMsg := p.FindMessage(f.GetInputMessageName()); inputMsg != nil {
				for _, field := range inputMsg.GetFields() {
					inputFields = append(inputFields, map[string]any{
						"name": field.GetName(),
						"type": field.GetType().GetType().String(),
					})
				}
			}
			flowsData = append(flowsData, map[string]any{
				"name":   f.GetName(),
				"inputs": inputFields,
			})
		}
		return common.NewJsonResponse(http.StatusOK, map[string]any{"flows": flowsData}, nil)
	}
}

func OpenAPISchemaHandler(p *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "FlowsAPI")
		defer span.End()
		span.SetAttributes(
			attribute.String("api.protocol", "HTTP JSON"),
		)

		if r.Method != http.MethodGet {
			return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
		}

		sch := openapi.GenerateFlows(ctx, p)
		return common.NewJsonResponse(http.StatusOK, sch, nil)
	}
}
