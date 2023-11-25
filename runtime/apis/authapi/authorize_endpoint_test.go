package authapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	defer server.Close()

	// Set up auth config
	redirectUrl := "https://myapp.com/signedup"
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		RedirectUrl: &redirectUrl,
		Providers: []config.Provider{
			{
				Type:             config.OpenIdConnectProvider,
				Name:             "myoidc",
				ClientId:         "oidc-client-id",
				IssuerUrl:        server.Issuer,
				TokenUrl:         server.TokenUrl,
				AuthorizationUrl: server.AuthorizeUrl,
			},
		},
	})

	// Set secret for client
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
	})

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)
	defer runtime.Close()

	t.Setenv("KEEL_API_URL", runtime.URL)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/myoidc",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.Contains(t, httpResponse.Request.Header["Referer"][0], "https://myapp.com/signedup?code=")
	require.Equal(t, http.StatusMovedPermanently, httpResponse.Request.Response.StatusCode)

	require.Contains(t, httpResponse.Request.Response.Request.Header["Referer"][0], runtime.URL+"/auth/callback/myoidc?code=")
	require.Equal(t, http.StatusFound, httpResponse.Request.Response.Request.Response.StatusCode)

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

func TestSsoLogin_WrongSecret(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	redirectUrl := "https://myapp.com/signedup"
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		RedirectUrl: &redirectUrl,
		Providers: []config.Provider{
			{
				Type:             config.OpenIdConnectProvider,
				Name:             "myoidc",
				ClientId:         "oidc-client-id",
				IssuerUrl:        server.Issuer,
				TokenUrl:         server.TokenUrl,
				AuthorizationUrl: server.AuthorizeUrl,
			},
		},
	})

	// Set secret for client
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "wrong-secret",
	})

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)
	defer runtime.Close()

	t.Setenv("KEEL_API_URL", runtime.URL)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/myoidc",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.Contains(t, httpResponse.Request.Header["Referer"][0], "https://myapp.com/signedup?error=access_denied&error_description=failed+to+exchange+code+at+provider+token+endpoint")
	require.Equal(t, http.StatusMovedPermanently, httpResponse.Request.Response.StatusCode)

	require.Contains(t, httpResponse.Request.Response.Request.Header["Referer"][0], runtime.URL+"/auth/callback/myoidc?code=")
	require.Equal(t, http.StatusFound, httpResponse.Request.Response.Request.Response.StatusCode)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 0)
}

func TestSsoLogin_InvalidLoginUrl(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	redirectUrl := "https://myapp.com/signedup"
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		RedirectUrl: &redirectUrl,
		Providers: []config.Provider{
			{
				Type:             config.OpenIdConnectProvider,
				Name:             "myoidc",
				ClientId:         "oidc-client-id",
				IssuerUrl:        server.Issuer,
				TokenUrl:         server.TokenUrl,
				AuthorizationUrl: server.AuthorizeUrl,
			},
		},
	})

	// Set secret for client
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
	})

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)
	defer runtime.Close()

	t.Setenv("KEEL_API_URL", runtime.URL)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/myoidc",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request with an unknown provider
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/unknown-oidc", nil)
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
	request, err = http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc/oops", nil)
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
	request, err = http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/", nil)
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

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 0)
}

func TestSsoLogin_MissingSecret(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	redirectUrl := "https://myapp.com/signedup"
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		RedirectUrl: &redirectUrl,
		Providers: []config.Provider{
			{
				Type:             config.OpenIdConnectProvider,
				Name:             "myoidc",
				ClientId:         "oidc-client-id",
				IssuerUrl:        server.Issuer,
				TokenUrl:         server.TokenUrl,
				AuthorizationUrl: server.AuthorizeUrl,
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
	defer runtime.Close()

	t.Setenv("KEEL_API_URL", runtime.URL)

	server.WithOAuthClient(&oauthtest.OAuthClient{
		ClientId:     "oidc-client-id",
		ClientSecret: "secret",
		RedirectUrl:  runtime.URL + "/auth/callback/myoidc",
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc", nil)
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
	require.Equal(t, "client secret not configured for provider: myoidc", errorResponse.ErrorDescription)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 0)
}

func TestSsoLogin_ClientIdNotRegistered(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	redirectUrl := "https://myapp.com/signedup"
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		RedirectUrl: &redirectUrl,
		Providers: []config.Provider{
			{
				Type:             config.OpenIdConnectProvider,
				Name:             "myoidc",
				ClientId:         "oidc-client-id",
				IssuerUrl:        server.Issuer,
				TokenUrl:         server.TokenUrl,
				AuthorizationUrl: server.AuthorizeUrl,
			},
		},
	})

	// Set secret for client
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
	})

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)
	defer runtime.Close()

	t.Setenv("KEEL_API_URL", runtime.URL)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.Contains(t, httpResponse.Request.Header["Referer"][0], "https://myapp.com/signedup?error=access_denied&error_description=provider+error%3A+invalid_request.+client+id+not+registered+on+server")
	require.Equal(t, http.StatusMovedPermanently, httpResponse.Request.Response.StatusCode)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 0)
}

func TestSsoLogin_RedirectUrlMismatch(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	redirectUrl := "https://myapp.com/signedup"
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		RedirectUrl: &redirectUrl,
		Providers: []config.Provider{
			{
				Type:             config.OpenIdConnectProvider,
				Name:             "myoidc",
				ClientId:         "oidc-client-id",
				IssuerUrl:        server.Issuer,
				TokenUrl:         server.TokenUrl,
				AuthorizationUrl: server.AuthorizeUrl,
			},
		},
	})

	// Set secret for client
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
	})

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)
	defer runtime.Close()

	t.Setenv("KEEL_API_URL", runtime.URL)

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
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc", nil)
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

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 0)
}

func TestSsoLogin_NoRedirectUrlInConfig(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:             config.OpenIdConnectProvider,
				Name:             "myoidc",
				ClientId:         "oidc-client-id",
				IssuerUrl:        server.Issuer,
				TokenUrl:         server.TokenUrl,
				AuthorizationUrl: server.AuthorizeUrl,
			},
		},
	})

	// Set secret for client
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		fmt.Sprintf("KEEL_AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
	})

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		h := runtime.NewHttpHandler(schema)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	runtime := httptest.NewServer(http.HandlerFunc(httpHandler))
	require.NoError(t, err)
	defer runtime.Close()

	t.Setenv("KEEL_API_URL", runtime.URL)

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
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc", nil)
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
	require.Equal(t, "redirectUrl must be specified in keelconfig.yaml", errorResponse.ErrorDescription)

}

func TestGetClientSecret(t *testing.T) {
	provider := &config.Provider{
		Name: "google",
	}

	ctx := context.Background()
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		"KEEL_AUTH_PROVIDER_SECRET_GOOGLE": "secret",
	})

	secret, hasSecret := authapi.GetClientSecret(ctx, provider)
	assert.True(t, hasSecret)
	assert.Equal(t, "secret", secret)
}

func TestGetClientSecret_WithUnderscore(t *testing.T) {
	provider := &config.Provider{
		Name: "google_client",
	}

	ctx := context.Background()
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		"KEEL_AUTH_PROVIDER_SECRET_GOOGLE_CLIENT": "secret",
	})

	secret, hasSecret := authapi.GetClientSecret(ctx, provider)
	assert.True(t, hasSecret)
	assert.Equal(t, "secret", secret)
}

func TestGetClientSecret_WithCapitals(t *testing.T) {
	provider := &config.Provider{
		Name: "GOOGLE_Client",
	}

	ctx := context.Background()
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		"KEEL_AUTH_PROVIDER_SECRET_GOOGLE": "secret",
	})

	secret, hasSecret := authapi.GetClientSecret(ctx, provider)
	assert.True(t, hasSecret)
	assert.Equal(t, "secret", secret)
}

func TestGetClientSecret_WithNumbers(t *testing.T) {
	provider := &config.Provider{
		Name: "client_2",
	}

	ctx := context.Background()
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		"KEEL_AUTH_PROVIDER_SECRET_CLIENT_2": "secret",
	})

	secret, hasSecret := authapi.GetClientSecret(ctx, provider)
	assert.True(t, hasSecret)
	assert.Equal(t, "secret", secret)
}

func TestGetClientSecret_NotExists(t *testing.T) {
	provider := &config.Provider{
		Name: "google",
	}

	secret, hasSecret := authapi.GetClientSecret(context.Background(), provider)
	assert.False(t, hasSecret)
	assert.Empty(t, secret)
}
