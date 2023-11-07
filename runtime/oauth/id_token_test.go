package oauth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

func TestIdTokenAuth_Valid(t *testing.T) {
	ctx := context.Background()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.NoError(t, err)
	require.NotNil(t, idToken)
}

func TestIdTokenAuth_IncorrectlySigned(t *testing.T) {
	ctx := context.Background()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{})
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
	ctx := context.Background()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	issuer := server.Config["issuer"]

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	server.Config["issuer"] = "http://accounts.google.com"

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("oidc: issuer did not match the issuer returned by provider, expected \"%s\" got \"http://accounts.google.com\"", issuer), err.Error())
	require.Nil(t, idToken)
}

func TestIdTokenAuth_IssuerNotRegistered(t *testing.T) {
	ctx := context.Background()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: false,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("issuer %s not registered to authenticate on this server", server.Issuer), err.Error())
	require.Nil(t, idToken)
}

func TestIdTokenAuth_ExpiredIdToken(t *testing.T) {
	ctx := context.Background()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.IdTokenLifespan = 0 * time.Second

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idTokenRaw, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	idToken, err := oauth.VerifyIdToken(ctx, idTokenRaw)

	require.Error(t, err)
	require.Contains(t, err.Error(), "oidc: token is expired")
	require.Nil(t, idToken)
}
