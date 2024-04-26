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
	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Redirect handler
	redirectHandler := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	require.NoError(t, err)

	// Set up auth config
	redirectUrl := redirectHandler.URL + "/signedup"
	ctx := runtimectx.WithOAuthConfig(context.TODO(), &config.AuthConfig{
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
		Claims: []config.IdentityClaim{
			{Key: "https://slack.com/#teamID", Field: "teamId"},
			{Key: "custom_claim", Field: "customClaim"},
			{Key: "not_exists", Field: "notExists"},
		},
	})

	ctx, database, schema := keeltesting.MakeContext(t, ctx, authTestSchema, true)
	defer database.Close()

	// Set secret for client
	ctx = runtimectx.WithSecrets(ctx, map[string]string{
		fmt.Sprintf("AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
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
		Name:          "name claim",
		GivenName:     "given name claim",
		FamilyName:    "family name claim",
		MiddleName:    "middle name claim",
		NickName:      "nick name claim",
		Profile:       "profile claim",
		Picture:       "picture claim",
		Website:       "website claim",
		Gender:        "gender claim",
		ZoneInfo:      "zoneInfo claim",
		Locale:        "locale claim",
	})

	// Make an SSO login request
	request, err := http.NewRequest(http.MethodPost, runtime.URL+"/auth/authorize/myoidc", nil)
	require.NoError(t, err)

	httpResponse, err := runtime.Client().Do(request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.Equal(t, http.StatusFound, httpResponse.Request.Response.StatusCode)
	require.Contains(t, httpResponse.Request.Response.Header["Location"][0], redirectUrl+"?code=")

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

	emailVerified, ok := identities[0]["email_verified"].(bool)
	require.True(t, ok)
	require.Equal(t, true, emailVerified)

	name, ok := identities[0]["name"].(string)
	require.True(t, ok)
	require.Equal(t, "name claim", name)

	givenName, ok := identities[0]["given_name"].(string)
	require.True(t, ok)
	require.Equal(t, "given name claim", givenName)

	familyName, ok := identities[0]["family_name"].(string)
	require.True(t, ok)
	require.Equal(t, "family name claim", familyName)

	middleName, ok := identities[0]["middle_name"].(string)
	require.True(t, ok)
	require.Equal(t, "middle name claim", middleName)

	nickName, ok := identities[0]["nick_name"].(string)
	require.True(t, ok)
	require.Equal(t, "nick name claim", nickName)

	profile, ok := identities[0]["profile"].(string)
	require.True(t, ok)
	require.Equal(t, "profile claim", profile)

	picture, ok := identities[0]["picture"].(string)
	require.True(t, ok)
	require.Equal(t, "picture claim", picture)

	website, ok := identities[0]["website"].(string)
	require.True(t, ok)
	require.Equal(t, "website claim", website)

	gender, ok := identities[0]["gender"].(string)
	require.True(t, ok)
	require.Equal(t, "gender claim", gender)

	zoneInfo, ok := identities[0]["zone_info"].(string)
	require.True(t, ok)
	require.Equal(t, "zoneInfo claim", zoneInfo)

	locale, ok := identities[0]["locale"].(string)
	require.True(t, ok)
	require.Equal(t, "locale claim", locale)

	teamId, ok := identities[0]["team_id"].(string)
	require.True(t, ok)
	require.Equal(t, "342352392354", teamId)

	customClaim, ok := identities[0]["custom_claim"].(string)
	require.True(t, ok)
	require.Equal(t, "custom value", customClaim)

	notExists := identities[0]["not_exists"]
	require.Nil(t, notExists)
}

func TestSsoLogin_WrongSecret(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Redirect handler
	redirectHandler := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	require.NoError(t, err)

	// Set up auth config
	redirectUrl := redirectHandler.URL + "/signedup"
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
		fmt.Sprintf("AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "wrong-secret",
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
	require.Equal(t, http.StatusFound, httpResponse.Request.Response.StatusCode)
	require.Contains(t, httpResponse.Request.Response.Header["Location"][0], redirectUrl+"?error=access_denied&error_description=failed+to+exchange+code+at+provider+token+endpoint")

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 0)
}

func TestSsoLogin_InvalidLoginUrl(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
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
		fmt.Sprintf("AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
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
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
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
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Redirect handler
	redirectHandler := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	require.NoError(t, err)

	// Set up auth config
	redirectUrl := redirectHandler.URL + "/signedup"
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
		fmt.Sprintf("AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
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
	require.Equal(t, http.StatusFound, httpResponse.Request.Response.StatusCode)
	require.Contains(t, httpResponse.Request.Response.Header["Location"][0], redirectUrl+"?error=access_denied&error_description=provider+error%3A+invalid_request.+client+id+not+registered+on+server")

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 0)
}

func TestSsoLogin_RedirectUrlMismatch(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
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
		fmt.Sprintf("AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
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
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
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
		fmt.Sprintf("AUTH_PROVIDER_SECRET_%s", strings.ToUpper("myoidc")): "secret",
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
		"AUTH_PROVIDER_SECRET_GOOGLE": "secret",
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
		"AUTH_PROVIDER_SECRET_GOOGLE_CLIENT": "secret",
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
		"AUTH_PROVIDER_SECRET_GOOGLE_CLIENT": "secret",
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
		"AUTH_PROVIDER_SECRET_CLIENT_2": "secret",
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
