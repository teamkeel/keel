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
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
)

// MyRunsHandler handles a request to /flows/json/myRuns and returns data about all flow runs ran by the current logged in user
func MyRunsHandler(p *proto.Schema) common.HandlerFunc {
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
		if identity == nil {
			return httpjson.NewErrorResponse(ctx, common.NewPermissionError(), nil)
		}
		ctx = auth.WithIdentity(ctx, identity)

		identityID := identity[parser.FieldNameId].(string)
		if identityID == "" {
			return httpjson.NewErrorResponse(ctx, common.NewPermissionError(), nil)
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

		runs, err := flows.ListUserFlowRuns(ctx, identityID, common.ParseQueryParams(r))
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}

		return common.NewJsonResponse(http.StatusOK, runs, nil)
	}
}
