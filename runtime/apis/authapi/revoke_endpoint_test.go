package authapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
	keeltesting "github.com/teamkeel/keel/testing"
)

func TestRevokeTokenForm_Success(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

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
	requestToken := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, requestToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// Make a token exchange grant request
	requestRevoke := makeRevokeTokenFormRequest(ctx, validResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	_, revokeHttpResponse, err := handleRuntimeRequest[any](schema, requestRevoke)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, revokeHttpResponse.StatusCode)

	// Make a token exchange grant request
	refreshToken := makeRefreshTokenFormRequest(ctx, validResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshResponse, refreshHttpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, refreshToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, refreshHttpResponse.StatusCode)
	require.Equal(t, "invalid_client", refreshResponse.Error)
}

func TestRevokeTokenJson_Success(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

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
	requestToken := makeTokenExchangeJsonRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, requestToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// Make a token exchange grant request
	requestRevoke := makeRevokeTokenJsonRequest(ctx, validResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	_, revokeHttpResponse, err := handleRuntimeRequest[any](schema, requestRevoke)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, revokeHttpResponse.StatusCode)

	// Make a token exchange grant request
	refreshToken := makeRefreshTokenJsonRequest(ctx, validResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshResponse, refreshHttpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, refreshToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, refreshHttpResponse.StatusCode)
	require.Equal(t, "invalid_client", refreshResponse.Error)
}

func TestRevokeEndpoint_HttpGet(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeRevokeTokenFormRequest(ctx, "mock_token")

	request.Method = http.MethodGet

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusMethodNotAllowed, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the revoke endpoint only accepts POST", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestRevokeEndpoint_EmptyToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a revoke request
	request := makeRevokeTokenFormRequest(ctx, "mock_token")
	form := url.Values{}
	form.Add("mock_token", "")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token must be provided in the token field", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestRevokeEndpoint_NoToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a revoke request
	request := makeRevokeTokenFormRequest(ctx, "mock_token")
	form := url.Values{}
	form.Del("token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token must be provided in the token field", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestRevokeEndpoint_UnknownToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a revoke request
	request := makeRevokeTokenFormRequest(ctx, "mock_token")

	// Handle runtime request, expecting TokenErrorResponse
	_, httpResponse, err := handleRuntimeRequest[any](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func makeRevokeTokenFormRequest(ctx context.Context, token string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/revoke", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("token", token)
	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}

func makeRevokeTokenJsonRequest(ctx context.Context, token string) *http.Request {
	values := map[string]string{
		"token": token,
	}

	jsonValue, _ := json.Marshal(values)
	responseBody := bytes.NewBuffer(jsonValue)

	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/revoke", responseBody)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	return request
}
