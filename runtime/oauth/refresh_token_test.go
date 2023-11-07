package oauth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	keeltesting "github.com/teamkeel/keel/testing"
)

var authTestSchema = `model Post{}`

func TestNewRefreshToken_NotEmpty(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)
}

func TestNewRefreshToken_ErrorOnEmptyIdentityId(t *testing.T) {
	ctx := context.Background()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	_, err := oauth.NewRefreshToken(ctx, "")
	require.Error(t, err)
}

func TestRotateRefreshToken_Valid(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	isValid1, newRefreshToken1, identityId1, err := oauth.RotateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.True(t, isValid1)
	require.Equal(t, "identity_id", identityId1)
	require.NotEmpty(t, newRefreshToken1)

	isValid2, newRefreshToken2, identityId2, err := oauth.RotateRefreshToken(ctx, newRefreshToken1)
	require.NoError(t, err)
	require.True(t, isValid2)
	require.Equal(t, "identity_id", identityId2)
	require.NotEmpty(t, newRefreshToken2)
	require.NotEqual(t, newRefreshToken2, newRefreshToken1)
}

func TestRotateRefreshToken_ReuseRefreshTokenNotValid(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	isValid, newRefreshToken, identityId, err := oauth.RotateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.True(t, isValid)
	require.Equal(t, "identity_id", identityId)
	require.NotEmpty(t, newRefreshToken)

	isValid2, newRefreshToken2, identityId2, err := oauth.RotateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.False(t, isValid2)
	require.Empty(t, identityId2)
	require.Empty(t, newRefreshToken2)
}

func TestRevokeRefreshToken_Unauthorised(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	err = oauth.RevokeRefreshToken(ctx, refreshToken)
	require.NoError(t, err)

	isValid, _, _, err := oauth.RotateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.False(t, isValid)
}

func TestRevokeRefreshToken_MultipleForIdentity(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	refreshToken1, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	refreshToken2, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	err = oauth.RevokeRefreshToken(ctx, refreshToken1)
	require.NoError(t, err)

	isValid1, _, _, err := oauth.RotateRefreshToken(ctx, refreshToken1)
	require.NoError(t, err)
	require.False(t, isValid1)

	isValid2, _, _, err := oauth.RotateRefreshToken(ctx, refreshToken2)
	require.NoError(t, err)
	require.True(t, isValid2)
}
