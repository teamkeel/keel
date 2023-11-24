package authapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
	keeltesting "github.com/teamkeel/keel/testing"
)

var authTestSchema = `model Post{}`

func TestTokenExchange_ValidNewIdentity(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.Equal(t, "bearer", validResponse.TokenType)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.NotEmpty(t, validResponse.RefreshToken)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	sub, iss, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)
	require.Equal(t, "https://keel.so", iss)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)
}

func TestTokenExchange_ValidNewIdentityAllUserInfo(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:               "keelson@keel.so",
		EmailVerified:       true,
		Name:                "Keely",
		GivenName:           "Keel",
		FamilyName:          "Keelson",
		MiddleName:          "Kool",
		NickName:            "Koolio",
		PreferredUsername:   "keel",
		Profile:             "https://github.com/teamkeel",
		Picture:             "https://avatars.githubusercontent.com/u/102726482?s=200&v=4",
		Website:             "https://keel.so",
		Gender:              "Unknown",
		ZoneInfo:            "Europe/Paris",
		Locale:              "fr-CA",
		PhoneNumber:         "+99 (999) 999-9999",
		PhoneNumberVerified: false,
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	sub, iss, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)
	require.Equal(t, "https://keel.so", iss)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)

	// TODO: test all the user info
}

func TestTokenExchange_ValidUpdatedIdentity(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	var inserted []map[string]any
	database.GetDB().Raw(fmt.Sprintf("INSERT INTO identity (external_id, issuer, email) VALUES ('id|285620','%s','weaveton@keel.so') RETURNING *", server.Issuer)).Scan(&inserted)
	require.Len(t, inserted, 1)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	sub, iss, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)
	require.Equal(t, "https://keel.so", iss)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)
	require.Equal(t, id, inserted[0]["id"].(string))

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)
}

func TestTokenEndpoint_HttpGet(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")

	request.Method = http.MethodGet

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusMethodNotAllowed, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the token endpoint only accepts POST", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_ApplicationJsonRequest(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")
	request.Header = http.Header{}
	request.Header.Add("Content-Type", "application/json")

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the request must be an encoded form with Content-Type application/x-www-form-urlencoded", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_MissingGrantType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("subject_token", "mock_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the grant-type field is required with either 'refresh_token', 'token_exchange' or 'authorization_code'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_WrongGrantType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("subject_token", "mock_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "unsupported_grant_type", errorResponse.Error)
	require.Equal(t, "the only supported grants are 'refresh_token', 'token_exchange' or 'authorization_code'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_NoSubjectToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the ID token must be provided in the subject_token field", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_EmptySubjectToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the ID token in the subject_token field cannot be an empty string", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_WrongSubjectTokenType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "mock_token")
	form.Add("subject_token_type", "access_token")
	form.Add("requested_token_type", "access_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the only supported subject_token_type is 'id_token'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_WrongRequestedTokenType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "mock_token")
	form.Add("subject_token_type", "id_token")
	form.Add("requested_token_type", "id_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the only supported requested_token_type is 'access_token'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_BadIdToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "this is not a jwt token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "possible causes may be that the id token is invalid, has expired, or has insufficient claims", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestRefreshTokenGrantRotationEnabled_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)

	// Set up auth config
	refreshTokenRotation := true
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Tokens: config.TokensConfig{
			RefreshTokenRotationEnabled: &refreshTokenRotation,
		},
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	tokenExchangeResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// We need 1 second to pass in order to get a different access token
	time.Sleep(1000 * time.Millisecond)

	// Make a refresh token grant request
	request = makeRefreshTokenRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, refreshGrantResponse.AccessToken)
	require.Equal(t, "bearer", refreshGrantResponse.TokenType)
	require.NotEmpty(t, refreshGrantResponse.ExpiresIn)
	require.NotEmpty(t, refreshGrantResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.RefreshToken, tokenExchangeResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.AccessToken, tokenExchangeResponse.AccessToken)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	accessToken1Issuer, err := auth.ExtractClaimFromToken(tokenExchangeResponse.AccessToken, "iss")
	require.NoError(t, err)
	accessToken2Issuer, err := auth.ExtractClaimFromToken(refreshGrantResponse.AccessToken, "iss")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Issuer)
	require.Equal(t, accessToken1Issuer, accessToken2Issuer)

	accessToken1Sub, err := auth.ExtractClaimFromToken(tokenExchangeResponse.AccessToken, "sub")
	require.NoError(t, err)
	accessToken2Sub, err := auth.ExtractClaimFromToken(refreshGrantResponse.AccessToken, "sub")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Sub)
	require.Equal(t, accessToken1Sub, accessToken2Sub)

	// Make a refresh token grant request using the original refresh token
	request = makeRefreshTokenRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenErrorResponse
	secondRefreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "possible causes may be that the refresh token has been revoked or has expired", secondRefreshGrantResponse.ErrorDescription)
}

func TestRefreshTokenGrantRotationDisabled_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)

	// Set up auth config
	refreshTokenRotation := false
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Tokens: config.TokensConfig{
			RefreshTokenRotationEnabled: &refreshTokenRotation,
		},
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	tokenExchangeResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// We need 1 second to pass in order to get a different access token
	time.Sleep(1000 * time.Millisecond)

	// Make a refresh token grant request
	request = makeRefreshTokenRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, refreshGrantResponse.AccessToken)
	require.Equal(t, "bearer", refreshGrantResponse.TokenType)
	require.NotEmpty(t, refreshGrantResponse.ExpiresIn)
	require.NotEmpty(t, refreshGrantResponse.RefreshToken)
	require.Equal(t, refreshGrantResponse.RefreshToken, tokenExchangeResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.AccessToken, tokenExchangeResponse.AccessToken)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	accessToken1Issuer, err := auth.ExtractClaimFromToken(tokenExchangeResponse.AccessToken, "iss")
	require.NoError(t, err)
	accessToken2Issuer, err := auth.ExtractClaimFromToken(refreshGrantResponse.AccessToken, "iss")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Issuer)
	require.Equal(t, accessToken1Issuer, accessToken2Issuer)

	accessToken1Sub, err := auth.ExtractClaimFromToken(tokenExchangeResponse.AccessToken, "sub")
	require.NoError(t, err)
	accessToken2Sub, err := auth.ExtractClaimFromToken(refreshGrantResponse.AccessToken, "sub")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Sub)
	require.Equal(t, accessToken1Sub, accessToken2Sub)

	// Make a refresh token grant request using the original refresh token
	request = makeRefreshTokenRequest(ctx, tokenExchangeResponse.RefreshToken)

	secondRefreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.Equal(t, tokenExchangeResponse.RefreshToken, secondRefreshGrantResponse.RefreshToken)
}

func TestRefreshTokenGrant_NoRefreshToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a refresh token grant request
	request := makeRefreshTokenRequest(ctx, "")
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token must be provided in the refresh_token field", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestRefreshTokenGrant_EmptyRefreshToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a refresh token grant request
	request := makeRefreshTokenRequest(ctx, "")

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token in the refresh_token field cannot be an empty string", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestAuthorizationCodeGrant_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	code, err := oauth.NewAuthCode(ctx, "identity_id")
	require.NoError(t, err)

	// Make a auth code grant request
	request := makeAuthorizationCodeRequest(ctx, code)

	// Handle runtime request, expecting TokenResponse
	response, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, response.AccessToken)
	require.Equal(t, "bearer", response.TokenType)
	require.NotEmpty(t, response.ExpiresIn)
	require.NotEmpty(t, response.RefreshToken)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	accessToken1Issuer, err := auth.ExtractClaimFromToken(response.AccessToken, "iss")
	require.NoError(t, err)
	require.Equal(t, accessToken1Issuer, "https://keel.so")

	accessToken1Sub, err := auth.ExtractClaimFromToken(response.AccessToken, "sub")
	require.NoError(t, err)
	require.Equal(t, accessToken1Sub, "identity_id")
}

func TestAuthorizationCodeGrant_InvalidCode(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a auth code grant request
	request := makeAuthorizationCodeRequest(ctx, "whoops")

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "possible causes may be that the auth code has been consumed or has expired", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestAuthorizationCodeGrant_NoCode(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a auth code grant request
	request := makeAuthorizationCodeRequest(ctx, "")
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the authorization code must be provided in the code field", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func handleRuntimeRequest[T any](schema *proto.Schema, req *http.Request) (T, *http.Response, error) {
	var response T
	handler := runtime.NewHttpHandler(schema)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	httpResponse := w.Result()

	if httpResponse.StatusCode == http.StatusInternalServerError {
		return response, nil, errors.New("internal server response from oidc server")
	}

	data, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return response, nil, err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return response, nil, err
	}

	return response, httpResponse, nil
}

func makeTokenExchangeRequest(ctx context.Context, token string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", token)
	form.Add("subject_token_type", "id_token")
	form.Add("requested_token_type", "access_token")
	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}

func makeRefreshTokenRequest(ctx context.Context, token string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", token)
	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}

func makeAuthorizationCodeRequest(ctx context.Context, code string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}
