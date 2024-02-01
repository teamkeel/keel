package authapi

import (
	"net/http"

	email "net/mail"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/bcrypt"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime")

// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
// https://datatracker.ietf.org/doc/html/rfc7009#section-2.1
const (
	ArgGrantType          = "grant_type"
	ArgSubjectToken       = "subject_token"
	ArgSubjectTokenType   = "subject_token_type"
	ArgRequestedTokenType = "requested_token_type"
	ArgCode               = "code"
	ArgRefreshToken       = "refresh_token"
	ArgToken              = "token"
	ArgUsername           = "username"
	ArgPassword           = "password"
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

// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
const (
	TokenErrUnsupportedGrantType = "unsupported_grant_type"
	TokenErrInvalidClient        = "invalid_client"
	TokenErrInvalidRequest       = "invalid_request"
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
			return common.InternalServerErrorResponse(ctx, err)
		}

		if r.Method != http.MethodPost {
			return jsonErrResponse(ctx, http.StatusMethodNotAllowed, TokenErrInvalidRequest, "the token endpoint only accepts POST", nil)
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

		grantType, hasGrantType := inputs[ArgGrantType].(string)
		if !hasGrantType || grantType == "" {
			return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the grant-type field is required with either 'refresh_token', 'token_exchange', 'authorization_code' or 'password'", nil)
		}

		span.SetAttributes(
			attribute.String(ArgGrantType, grantType),
		)

		switch grantType {
		case GrantTypeRefreshToken:
			refreshTokenRaw, hasRefreshTokenRaw := inputs[ArgRefreshToken].(string)
			if !hasRefreshTokenRaw || refreshTokenRaw == "" {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the refresh token in the 'refresh_token' field is required", nil)
			}

			var isValid bool
			if config.RefreshTokenRotationEnabled() {
				// Rotate and revoke this refresh token, and mint a new one.
				isValid, refreshToken, identityId, err = oauth.RotateRefreshToken(ctx, refreshTokenRaw)
				if err != nil {
					return common.InternalServerErrorResponse(ctx, err)
				}
			} else {
				// Response with the same refresh token when refresh token rotation is disabled
				refreshToken = refreshTokenRaw

				// Check that the refresh token exists and has not expired.
				isValid, identityId, err = oauth.ValidateRefreshToken(ctx, refreshToken)
				if err != nil {
					return common.InternalServerErrorResponse(ctx, err)
				}
			}

			if !isValid {
				return jsonErrResponse(ctx, http.StatusUnauthorized, TokenErrInvalidClient, "possible causes may be that the refresh token has been revoked or has expired", nil)
			}

		case GrantTypePassword:
			username, hasUsername := inputs[ArgUsername].(string)
			if !hasUsername || username == "" {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the identity's email in the 'username' field is required", nil)
			}

			if _, err := email.ParseAddress(username); err != nil {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "invalid email address", nil)
			}

			password, hasPassword := inputs[ArgPassword].(string)
			if !hasPassword || password == "" {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the identity's password in the 'password' field is required", nil)
			}

			identity, err := actions.FindIdentityByEmail(ctx, schema, username, oauth.KeelIssuer)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			if identity == nil {
				hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					return common.InternalServerErrorResponse(ctx, err)
				}

				identity, err = actions.CreateIdentity(ctx, schema, username, string(hashedBytes), oauth.KeelIssuer)
				if err != nil {
					return common.InternalServerErrorResponse(ctx, err)
				}
			} else {
				correct := bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(password)) == nil
				if !correct {
					return jsonErrResponse(ctx, http.StatusUnauthorized, TokenErrInvalidClient, "possible causes may be that the identity does not exist or the credentials are incorrect", nil)
				}
			}

			// Generate a refresh token.
			refreshToken, err = oauth.NewRefreshToken(ctx, identity.Id)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			identityId = identity.Id

		case GrantTypeAuthCode:
			authCode, hasAuthCode := inputs[ArgCode].(string)
			if !hasAuthCode || authCode == "" {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the authorization code in the 'code' field is required", nil)
			}

			// Consume the auth code
			var isValid bool
			isValid, identityId, err = oauth.ConsumeAuthCode(ctx, authCode)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			if !isValid {
				return jsonErrResponse(ctx, http.StatusUnauthorized, TokenErrInvalidClient, "possible causes may be that the auth code has been consumed or has expired", nil)
			}

			// Generate a refresh token.
			refreshToken, err = oauth.NewRefreshToken(ctx, identityId)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

		case GrantTypeTokenExchange:
			idTokenRaw, hasIdTokenRaw := inputs[ArgSubjectToken].(string)
			if !hasIdTokenRaw || idTokenRaw == "" {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the ID token must be provided in the 'subject_token' field", nil)
			}

			// We do not require subject_token_type, but if provided we only support 'id_token'
			if tokenType, hasTokenType := inputs[ArgSubjectTokenType]; hasTokenType && tokenType != "id_token" {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the only supported subject_token_type is 'id_token'", nil)
			} else if hasTokenType {
				span.SetAttributes(attribute.String(ArgSubjectTokenType, tokenType.(string)))
			}

			// We do not require requested_token_type, but if provided we only support 'access_token'
			if reqTokenType, hasReqTokenType := inputs[ArgRequestedTokenType]; hasReqTokenType && reqTokenType != "access_token" && reqTokenType != "urn:ietf:params:oauth:token-type:access_token" {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "the only supported requested_token_type is 'access_token'", nil)
			} else if hasReqTokenType {
				span.SetAttributes(attribute.String(ArgRequestedTokenType, reqTokenType.(string)))
			}

			// Verify the ID token with the OIDC provider
			idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)
			if err != nil {
				return jsonErrResponse(ctx, http.StatusUnauthorized, TokenErrInvalidClient, "possible causes may be that the id token is invalid, has expired, or has insufficient claims", err)
			}

			// Extract claims
			var claims oauth.IdTokenClaims
			if err := idToken.Claims(&claims); err != nil {
				return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrInvalidRequest, "insufficient claims on id_token", err)
			}

			identity, err := actions.FindIdentityByExternalId(ctx, schema, idToken.Subject, idToken.Issuer)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			if identity == nil {
				identity, err = actions.CreateIdentityWithIdTokenClaims(ctx, schema, idToken.Subject, idToken.Issuer, claims)
				if err != nil {
					return common.InternalServerErrorResponse(ctx, err)
				}
			} else {
				identity, err = actions.UpdateIdentityWithIdTokenClaims(ctx, schema, idToken.Subject, idToken.Issuer, claims)
				if err != nil {
					return common.InternalServerErrorResponse(ctx, err)
				}
			}

			// Generate a refresh token.
			refreshToken, err = oauth.NewRefreshToken(ctx, identity.Id)
			if err != nil {
				return common.InternalServerErrorResponse(ctx, err)
			}

			identityId = identity.Id

		default:
			return jsonErrResponse(ctx, http.StatusBadRequest, TokenErrUnsupportedGrantType, "the only supported grants are 'refresh_token', 'token_exchange', 'authorization_code' or 'password'", nil)
		}

		// Generate a new access token for this identity.
		accessTokenRaw, expiresIn, err := oauth.GenerateAccessToken(ctx, identityId)
		if err != nil {
			return common.InternalServerErrorResponse(ctx, err)
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
