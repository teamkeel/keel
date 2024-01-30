package authapi

import (
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
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
			return jsonErrResponse(ctx, http.StatusMethodNotAllowed, TokenErrInvalidRequest, "the revoke endpoint only accepts POST", nil)
		}

		if !common.HasContentType(r.Header, "application/x-www-form-urlencoded") && !common.HasContentType(r.Header, "application/json") {
			return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the request body must either be an encoded form (Content-Type: application/x-www-form-urlencoded) or JSON (Content-Type: application/json)", nil)
		}

		data, err := common.ParseRequestData(r)
		if err != nil {
			return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "request payload is malformed", err)
		}

		inputs, ok := data.(map[string]any)
		if !ok {
			return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "request payload is malformed", err)
		}

		refreshTokenRaw, hasRefreshTokenRaw := inputs[ArgToken].(string)
		if !hasRefreshTokenRaw || refreshTokenRaw == "" {
			return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the refresh token must be provided in the token field", nil)
		}

		// Revoke the refresh token
		err = oauth.RevokeRefreshToken(ctx, refreshTokenRaw)
		if err != nil {
			return jsonErrResponse(ctx, http.StatusUnauthorized, TokenErrInvalidClient, "possible causes may be that the id token is invalid, has expired, or has insufficient claims", err)
		}

		return common.NewJsonResponse(http.StatusOK, nil, nil)
	}
}
