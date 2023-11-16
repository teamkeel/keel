package authapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
	keeltesting "github.com/teamkeel/keel/testing"
)

func TestRevokeToken_Success(t *testing.T) {
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
	requestToken := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, requestToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// Make a token exchange grant request
	requestRevoke := makeRevokeTokenRequest(ctx, validResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	_, revokeHttpResponse, err := handleRuntimeRequest[any](schema, requestRevoke)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, revokeHttpResponse.StatusCode)

	// Make a token exchange grant request
	refreshToken := makeRefreshTokenRequest(ctx, validResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshResponse, refreshHttpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, refreshToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, refreshHttpResponse.StatusCode)
	require.Equal(t, "invalid_client", refreshResponse.Error)
}

func TestRevokeEndpoint_HttpGet(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeRevokeTokenRequest(ctx, "mock_token")

	request.Method = http.MethodGet

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusMethodNotAllowed, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the revoke endpoint only accepts POST", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestRevokeEndpoint_EmptyToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a revoke request
	request := makeRevokeTokenRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("mock_token", "")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token must be provided in the token field", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestRevokeEndpoint_NoToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a revoke request
	request := makeRevokeTokenRequest(ctx, "mock_token")
	form := url.Values{}
	form.Del("token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token must be provided in the token field", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestRevokeEndpoint_UnknownToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Make a revoke request
	request := makeRevokeTokenRequest(ctx, "mock_token")

	// Handle runtime request, expecting TokenErrorResponse
	_, httpResponse, err := handleRuntimeRequest[any](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func makeRevokeTokenRequest(ctx context.Context, token string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/revoke", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("token", token)
	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}
