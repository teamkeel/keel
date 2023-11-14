package authapi

import (
	"mime"
	"net/http"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime")

// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
// https://datatracker.ietf.org/doc/html/rfc7009#section-2.1
const (
	ArgGrantType         = "grant_type"
	ArgSubjectToken      = "subject_token"
	ArgSubjectTokenType  = "subject_token_type"
	ArgRequestedTokeType = "requested_token_type"
	ArgRefreshToken      = "refresh_token"
	ArgToken             = "token"
)

const (
	TokenType = "bearer"
)

// https://openid.net/specs/openid-connect-standard-1_0-21_orig.html#AccessTokenResponse
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// https://openid.net/specs/openid-connect-standard-1_0-21_orig.html#AccessTokenErrorResponse
// https://datatracker.ietf.org/doc/html/rfc7009#section-2.2
type ErrorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
const (
	UnsupportedGrantType = "unsupported_grant_type"
	InvalidClient        = "invalid_client"
	InvalidRequest       = "invalid_request"
)

const (
	GrantTypeImplicit          = "implicit"
	GrantTypePassword          = "password"
	GrantTypeClientCredentials = "client_credentials"
	GrantTypeAuthCode          = "authorization_code"
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeTokenExchange     = "token_exchange"
)

// TokenEndpointHandler handles requests to the token endpoint for the various grant types we support.
// OAuth2.0 specification: https://datatracker.ietf.org/doc/html/rfc6749#section-3.2
// OpenID Connect specification for Token Endpoint: https://openid.net/specs/openid-connect-standard-1_0-21_orig.html#token_ep
func TokenEndpointHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Token Endpoint")
		defer span.End()

		var identityId string
		var refreshToken string

		config, err := runtimectx.GetOAuthConfig(ctx)
		if err != nil {
			span.RecordError(err)
			return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
		}

		if r.Method != http.MethodPost {
			return common.NewJsonResponse(http.StatusMethodNotAllowed, &ErrorResponse{
				Error:            InvalidRequest,
				ErrorDescription: "the token endpoint only accepts POST",
			}, nil)
		}

		if !HasContentType(r.Header, "application/x-www-form-urlencoded") {
			return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
				Error:            InvalidRequest,
				ErrorDescription: "the request must be an encoded form with Content-Type application/x-www-form-urlencoded",
			}, nil)
		}

		grantType := r.FormValue(ArgGrantType)

		if grantType == "" {
			return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
				Error:            InvalidRequest,
				ErrorDescription: "the grant-type field is required with either 'refresh_token' or 'token_exchange'",
			}, nil)
		}

		span.SetAttributes(
			attribute.String(ArgGrantType, grantType),
		)

		switch grantType {
		case GrantTypeRefreshToken:
			if !r.Form.Has(ArgRefreshToken) {
				return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
					Error:            InvalidRequest,
					ErrorDescription: "the refresh token must be provided in the refresh_token field",
				}, nil)
			}

			refreshTokenRaw := r.FormValue(ArgRefreshToken)

			if refreshTokenRaw == "" {
				return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
					Error:            InvalidRequest,
					ErrorDescription: "the refresh token in the refresh_token field cannot be an empty string",
				}, nil)
			}

			var isValid bool
			if config.RefreshTokenRotationEnabled() {
				// Rotate and revoke this refresh token, and mint a new one.
				isValid, refreshToken, identityId, err = oauth.RotateRefreshToken(ctx, refreshTokenRaw)
				if err != nil {
					span.RecordError(err)
					return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
				}
			} else {
				// Response with the same refresh token when refresh token rotation is disabled
				refreshToken = refreshTokenRaw

				// Check that the refresh token exists and has not expired.
				isValid, identityId, err = oauth.ValidateRefreshToken(ctx, refreshToken)
				if err != nil {
					span.RecordError(err)
					return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
				}
			}

			if !isValid {
				return common.NewJsonResponse(http.StatusUnauthorized, &ErrorResponse{
					Error:            InvalidClient,
					ErrorDescription: "possible causes may be that the refresh token has been revoked or has expired",
				}, nil)
			}

		case GrantTypeTokenExchange:
			if !r.Form.Has(ArgSubjectToken) {
				return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
					Error:            InvalidRequest,
					ErrorDescription: "the ID token must be provided in the subject_token field",
				}, nil)
			}

			// We do not require subject_token_type, but if provided we only support 'id_token'
			if r.Form.Has(ArgSubjectTokenType) && r.Form.Get(ArgSubjectTokenType) != "id_token" {
				return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
					Error:            InvalidRequest,
					ErrorDescription: "the only supported subject_token_type is 'id_token'",
				}, nil)
			}

			// We do not require requested_token_type, but if provided we only support 'access_token'
			if r.Form.Has(ArgRequestedTokeType) && (r.Form.Get(ArgRequestedTokeType) != "urn:ietf:params:oauth:token-type:access_token" && r.Form.Get("requested_token_type") != "access_token") {
				return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
					Error:            InvalidRequest,
					ErrorDescription: "the only supported requested_token_type is 'access_token'",
				}, nil)
			}

			idTokenRaw := r.Form.Get(ArgSubjectToken)

			if idTokenRaw == "" {
				return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
					Error:            InvalidRequest,
					ErrorDescription: "the ID token in the subject_token field cannot be an empty string",
				}, nil)
			}

			span.SetAttributes(
				attribute.String(ArgSubjectTokenType, r.Form.Get(ArgSubjectTokenType)),
				attribute.String(ArgRequestedTokeType, r.Form.Get(ArgRequestedTokeType)),
			)

			// Verify the ID token with the OIDC provider
			idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)
			if err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusUnauthorized, &ErrorResponse{
					Error:            InvalidClient,
					ErrorDescription: "possible causes may be that the id token is invalid, has expired, or has insufficient claims",
				}, nil)
			}

			// Extract claims
			var claims oauth.IdTokenClaims
			if err := idToken.Claims(&claims); err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
					Error:            InvalidRequest,
					ErrorDescription: "insufficient claims on id_token",
				}, nil)
			}

			identity, err := actions.FindIdentityByExternalId(ctx, schema, idToken.Subject, idToken.Issuer)
			if err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
			}

			if identity == nil {
				identity, err = actions.CreateIdentityWithIdTokenClaims(ctx, schema, idToken.Subject, idToken.Issuer, claims)
				if err != nil {
					span.RecordError(err)
					return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
				}
			} else {
				identity, err = actions.UpdateIdentityWithIdTokenClaims(ctx, schema, idToken.Subject, idToken.Issuer, claims)
				if err != nil {
					span.RecordError(err)
					return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
				}
			}

			// Generate a refresh token.
			refreshToken, err = oauth.NewRefreshToken(ctx, identity.Id)
			if err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
			}

			identityId = identity.Id

		default:
			return common.NewJsonResponse(http.StatusBadRequest, &ErrorResponse{
				Error:            UnsupportedGrantType,
				ErrorDescription: "the only supported grants are 'refresh_token' and 'token_exchange'",
			}, nil)
		}

		// Generate a new access token for this identity.
		accessTokenRaw, expiresIn, err := oauth.GenerateAccessToken(ctx, identityId)
		if err != nil {
			span.RecordError(err)
			return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
		}

		response := &TokenResponse{
			AccessToken:  accessTokenRaw,
			TokenType:    TokenType,
			ExpiresIn:    int(expiresIn.Seconds()),
			RefreshToken: refreshToken,
		}

		return common.NewJsonResponse(http.StatusOK, response, nil)
	}
}

func HasContentType(headers http.Header, mimetype string) bool {
	contentType := headers.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}

		if t == mimetype {
			return true
		}
	}
	return false
}
