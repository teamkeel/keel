package oauth_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

func TestIdTokenAuth_Valid(t *testing.T) {
	ctx := t.Context()

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
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.NoError(t, err)
	require.NotNil(t, idToken)
}

func TestIdTokenAuthNoEmail_Valid(t *testing.T) {
	ctx := t.Context()

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

	server.SetUser("id|285620", &oauth.UserClaims{})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.NoError(t, err)
	require.NotNil(t, idToken)
}

func TestIdTokenAuthMultipleIssuers_Valid(t *testing.T) {
	ctx := t.Context()

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
				ClientId:  "oidc-client-id-1",
				IssuerUrl: server.Issuer,
			},
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id-2",
				IssuerUrl: server.Issuer,
			},
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id-3",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"oidc-client-id-3"})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.NoError(t, err)
	require.NotNil(t, idToken)
}

func TestIdTokenAuth_IncorrectlySigned(t *testing.T) {
	ctx := t.Context()

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
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Renew the public set, thus making the ID token invalid
	err = server.RenewPrivateKey()
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Equal(t, "failed to verify signature: failed to verify id token signature", err.Error())
	require.Nil(t, idToken)
}

func TestIdTokenAuth_IssuerMismatch(t *testing.T) {
	ctx := t.Context()

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

	issuer := server.Config["issuer"]

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	server.Config["issuer"] = "http://accounts.google.com"

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("oidc: issuer did not match the issuer returned by provider, expected \"%s\" got \"http://accounts.google.com\"", issuer), err.Error())
	require.Nil(t, idToken)
}

func TestIdTokenAuth_ClientIdMismatch(t *testing.T) {
	ctx := t.Context()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc-1",
				ClientId:  "oidc-client-id-1",
				IssuerUrl: server.Issuer,
			},
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc-2",
				ClientId:  "oidc-client-id-2",
				IssuerUrl: server.Issuer,
			},
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc-3",
				ClientId:  "oidc-client-id-3",
				IssuerUrl: server.Issuer,
			},
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc-4",
				ClientId:  "oidc-client-id-4",
				IssuerUrl: "https://someother.com",
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"different-client-id"})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Contains(t, err.Error(), "oidc: expected audience \"oidc-client-id-1\" got [\"different-client-id\"]")
	require.Contains(t, err.Error(), "oidc: expected audience \"oidc-client-id-2\" got [\"different-client-id\"]")
	require.Contains(t, err.Error(), "oidc: expected audience \"oidc-client-id-3\" got [\"different-client-id\"]")
	require.NotContains(t, err.Error(), "oidc: expected audience \"oidc-client-id-4\" got [\"different-client-id\"]")
	require.Nil(t, idToken)
}

func TestIdTokenAuth_IssuerNotConfigured(t *testing.T) {
	ctx := t.Context()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config with no issuer
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("issuer %s not registered to authenticate on this server", server.Issuer), err.Error())
	require.Nil(t, idToken)
}

func TestIdTokenAuth_ExpiredIdToken(t *testing.T) {
	ctx := t.Context()

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

	server.IdTokenLifespan = 0 * time.Second

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Contains(t, err.Error(), "oidc: token is expired")
	require.Nil(t, idToken)
}
