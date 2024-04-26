package oauth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	keeltesting "github.com/teamkeel/keel/testing"
)

var authTestSchema = `model Post{}`

func TestNewRefreshToken_NotEmpty(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)
}

func TestNewRefreshToken_ErrorOnEmptyIdentityId(t *testing.T) {
	ctx := context.Background()

	_, err := oauth.NewRefreshToken(ctx, "")
	require.Error(t, err)
}

func TestRotateRefreshToken_Valid(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

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

func TestRotateRefreshToken_Expired(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

	// Set up auth config
	seconds := 1
	config := config.AuthConfig{
		Tokens: config.TokensConfig{
			RefreshTokenExpiry: &seconds,
		},
	}
	ctx = runtimectx.WithOAuthConfig(ctx, &config)

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)

	isValid, newRefreshToken, identityId, err := oauth.RotateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.False(t, isValid)
	require.Empty(t, identityId)
	require.Empty(t, newRefreshToken)
}

func TestRotateRefreshToken_ReuseRefreshTokenNotValid(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

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

func TestValidateRefreshToken_Valid(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	isValid, identityId, err := oauth.ValidateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.True(t, isValid)
	require.Equal(t, "identity_id", identityId)
}

func TestValidateRefreshToken_Expired(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

	// Set up auth config
	seconds := 1
	config := config.AuthConfig{
		Tokens: config.TokensConfig{
			RefreshTokenExpiry: &seconds,
		},
	}
	ctx = runtimectx.WithOAuthConfig(ctx, &config)

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)

	isValid, identityId, err := oauth.ValidateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.False(t, isValid)
	require.Empty(t, identityId)
}

func TestRevokeRefreshToken_Unauthorised(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

	refreshToken, err := oauth.NewRefreshToken(ctx, "identity_id")
	require.NoError(t, err)

	err = oauth.RevokeRefreshToken(ctx, refreshToken)
	require.NoError(t, err)

	isValid, _, _, err := oauth.RotateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.False(t, isValid)
}

func TestRevokeRefreshToken_MultipleForIdentity(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, context.TODO(), authTestSchema, true)
	defer database.Close()

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
