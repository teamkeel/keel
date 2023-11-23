package authapi

import (
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"go.opentelemetry.io/otel/trace"
)

type RevokeEndpointErrorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func RevokeHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Revoke Token")
		defer span.End()

		if r.Method != http.MethodPost {
			return common.NewJsonResponse(http.StatusMethodNotAllowed, &ErrorResponse{
				Error:            TokenErrInvalidRequest,
				ErrorDescription: "the revoke endpoint only accepts POST",
			}, nil)
		}

		if !HasContentType(r.Header, "application/x-www-form-urlencoded") {
			return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
				Error:            TokenErrInvalidRequest,
				ErrorDescription: "the request must be an encoded form with Content-Type application/x-www-form-urlencoded",
			}, nil)
		}

		refreshTokenRaw := r.FormValue(ArgToken)

		if refreshTokenRaw == "" {
			return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
				Error:            TokenErrInvalidRequest,
				ErrorDescription: "the refresh token must be provided in the token field",
			}, nil)
		}

		// Revoke the refresh token
		err := oauth.RevokeRefreshToken(ctx, refreshTokenRaw)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			return common.NewJsonResponse(http.StatusUnauthorized, &ErrorResponse{
				Error:            TokenErrInvalidClient,
				ErrorDescription: "possible causes may be that the id token is invalid, has expired, or has insufficient claims",
			}, nil)
		}

		return common.NewJsonResponse(http.StatusOK, nil, nil)
	}
}
