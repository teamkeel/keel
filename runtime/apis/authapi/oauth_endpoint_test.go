package authapi_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
	keeltesting "github.com/teamkeel/keel/testing"
)

func TestSsoLogin_Success(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)

	redirectUrl := "https://myapp.com/signedup"
	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		RedirectUrl: &redirectUrl,
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	// Set secret for client
	t.Setenv(fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("my-oidc")), "secret")

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/my-oidc",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/login/my-oidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	_, ok := identities[0]["id"].(string)
	require.True(t, ok)

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

func TestSsoLogin_InvalidLoginUrl(t *testing.T) {
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

	// Set secret for client
	t.Setenv(fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("my-oidc")), "secret")

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/my-oidc",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request with an unknown provider
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/login/unknown-oidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	data, err := io.ReadAll(httpResponse.Body)
	require.NoError(t, err)

	var errorResponse authapi.ErrorResponse
	err = json.Unmarshal(data, &errorResponse)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "login url malformed or provider not found", errorResponse.ErrorDescription)

	// Make an SSO login request with additional fragment
	request, err = http.NewRequest(http.MethodPost, runtime.URL+"/auth/login/my-oidc/oops", nil)
	require.NoError(t, err)

	httpResponse, err = runtime.Client().Do(request)
	require.NoError(t, err)

	data, err = io.ReadAll(httpResponse.Body)
	require.NoError(t, err)

	err = json.Unmarshal(data, &errorResponse)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "login url malformed or provider not found", errorResponse.ErrorDescription)

	// Make an SSO login request without a provider
	request, err = http.NewRequest(http.MethodPost, runtime.URL+"/auth/login/", nil)
	require.NoError(t, err)

	httpResponse, err = runtime.Client().Do(request)
	require.NoError(t, err)

	data, err = io.ReadAll(httpResponse.Body)
	require.NoError(t, err)

	err = json.Unmarshal(data, &errorResponse)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "login url malformed or provider not found", errorResponse.ErrorDescription)
}

func TestSsoLogin_MissingSecret(t *testing.T) {
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

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/my-oidc",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/login/my-oidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	data, err := io.ReadAll(httpResponse.Body)
	require.NoError(t, err)

	var errorResponse authapi.ErrorResponse
	err = json.Unmarshal(data, &errorResponse)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "client secret not configured for provider: my-oidc", errorResponse.ErrorDescription)
}

func TestSsoLogin_ClientIdNotRegistered(t *testing.T) {
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

	// Set secret for client
	t.Setenv(fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("my-oidc")), "secret")

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/login/my-oidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	data, err := io.ReadAll(httpResponse.Body)
	require.NoError(t, err)

	var errorResponse authapi.ErrorResponse
	err = json.Unmarshal(data, &errorResponse)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "provider could not authenticate due to invalid_request: client id not registered on server", errorResponse.ErrorDescription)
}

func TestSsoLogin_RedirectUrlMismatch(t *testing.T) {
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

	// Set secret for client
	t.Setenv(fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("my-oidc")), "secret")

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/mismatch",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/login/my-oidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	data, err := io.ReadAll(httpResponse.Body)
	require.NoError(t, err)

	var errorResponse authapi.ErrorResponse
	err = json.Unmarshal(data, &errorResponse)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "redirect uri does not match", errorResponse.ErrorDescription)
}
