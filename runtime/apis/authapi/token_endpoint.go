package authapi

import (
	"mime"
	"net/http"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime")

// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
const (
	ArgGrantType         = "grant_type"
	ArgSubjectToken      = "subject_token"
	ArgSubjectTokenType  = "subject_token_type"
	ArgRequestedTokeType = "requested_token_type"
	ArgRefreshToken      = "refresh_token"
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
type TokenErrorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
const (
	TokenEndpointUnsupportedGrantType = "unsupported_grant_type"
	TokenEndpointInvalidClient        = "invalid_client"
	TokenEndpointInvalidRequest       = "invalid_request"
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

		if r.Method != http.MethodPost {
			return common.NewJsonResponse(http.StatusMethodNotAllowed, &TokenErrorResponse{
				Error:            TokenEndpointInvalidRequest,
				ErrorDescription: "the token endpoint only accepts POST",
			}, nil)
		}

		if !HasContentType(r.Header, "application/x-www-form-urlencoded") {
			return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
				Error:            TokenEndpointInvalidRequest,
				ErrorDescription: "the request must be an encoded form with Content-Type application/x-www-form-urlencoded",
			}, nil)
		}

		grantType := r.FormValue(ArgGrantType)

		if grantType == "" {
			return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
				Error:            TokenEndpointInvalidRequest,
				ErrorDescription: "the grant-type field is required with either 'refresh_token' or 'token_exchange'",
			}, nil)
		}

		span.SetAttributes(
			attribute.String(ArgGrantType, grantType),
		)

		switch grantType {
		case GrantTypeRefreshToken:
			if !r.Form.Has(ArgRefreshToken) {
				return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
					Error:            TokenEndpointInvalidRequest,
					ErrorDescription: "the refresh token must be provided in the refresh_token field",
				}, nil)
			}

			refreshTokenRaw := r.Form.Get(ArgRefreshToken)

			if refreshTokenRaw == "" {
				return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
					Error:            TokenEndpointInvalidRequest,
					ErrorDescription: "the refresh token in the refresh_token field cannot be an empty string",
				}, nil)
			}

			isValid, newRefreshToken, identityId, err := oauth.RotateRefreshToken(ctx, refreshTokenRaw)
			if err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
			}

			if !isValid {
				return common.NewJsonResponse(http.StatusUnauthorized, &TokenErrorResponse{
					Error:            TokenEndpointInvalidClient,
					ErrorDescription: "possible causes may be that the refresh token has been revoked or has expired",
				}, nil)
			}

			// Generate an access token for this identity.
			accessTokenRaw, expiresIn, err := oauth.GenerateAccessToken(ctx, identityId)
			if err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
			}

			response := &TokenResponse{
				AccessToken:  accessTokenRaw,
				TokenType:    TokenType,
				ExpiresIn:    int(expiresIn.Seconds()),
				RefreshToken: newRefreshToken,
			}

			return common.NewJsonResponse(http.StatusOK, response, nil)
		case GrantTypeTokenExchange:
			if !r.Form.Has(ArgSubjectToken) {
				return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
					Error:            TokenEndpointInvalidRequest,
					ErrorDescription: "the ID token must be provided in the subject_token field",
				}, nil)
			}

			// We do not require subject_token_type, but if provided we only support 'id_token'
			if r.Form.Has(ArgSubjectTokenType) && r.Form.Get(ArgSubjectTokenType) != "id_token" {
				return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
					Error:            TokenEndpointInvalidRequest,
					ErrorDescription: "the only supported subject_token_type is 'id_token'",
				}, nil)
			}

			// We do not require requested_token_type, but if provided we only support 'access_token'
			if r.Form.Has(ArgRequestedTokeType) && (r.Form.Get(ArgRequestedTokeType) != "urn:ietf:params:oauth:token-type:access_token" && r.Form.Get("requested_token_type") != "access_token") {
				return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
					Error:            TokenEndpointInvalidRequest,
					ErrorDescription: "the only supported requested_token_type is 'access_token'",
				}, nil)
			}

			idTokenRaw := r.Form.Get(ArgSubjectToken)

			if idTokenRaw == "" {
				return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
					Error:            TokenEndpointInvalidRequest,
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
				return common.NewJsonResponse(http.StatusUnauthorized, &TokenErrorResponse{
					Error:            TokenEndpointInvalidClient,
					ErrorDescription: "possible causes may be that the id token is invalid, has expired, or has insufficient claims",
				}, nil)
			}

			// Extract claims
			var claims oauth.IdTokenClaims
			if err := idToken.Claims(&claims); err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
					Error:            TokenEndpointInvalidRequest,
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

			// Generate an access token for this identity.
			accessTokenRaw, expiresIn, err := oauth.GenerateAccessToken(ctx, identity.Id)
			if err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
			}

			// Generate a refresh token.
			refreshTokenRaw, err := oauth.NewRefreshToken(ctx, identity.Id)
			if err != nil {
				span.RecordError(err)
				return common.NewJsonResponse(http.StatusInternalServerError, nil, nil)
			}

			response := &TokenResponse{
				AccessToken:  accessTokenRaw,
				TokenType:    TokenType,
				ExpiresIn:    int(expiresIn.Seconds()),
				RefreshToken: refreshTokenRaw,
			}

			return common.NewJsonResponse(http.StatusOK, response, nil)

		default:
			return common.NewJsonResponse(http.StatusBadRequest, &TokenErrorResponse{
				Error:            TokenEndpointUnsupportedGrantType,
				ErrorDescription: "the only supported grants are 'refresh_token' and 'token_exchange'",
			}, nil)
		}
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
