package oauth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/oauth"
	keeltesting "github.com/teamkeel/keel/testing"
)

func TestNewAuthCode_NotEmpty(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	code, err := oauth.NewAuthCode(ctx, "identity_id")
	require.NoError(t, err)
	require.Len(t, code, 32)
}

func TestNewAuthCode_ErrorOnEmptyIdentityId(t *testing.T) {
	ctx := context.Background()

	_, err := oauth.NewAuthCode(ctx, "")
	require.Error(t, err)
}

func TestConsumeAuthCode_Success(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	code, err := oauth.NewAuthCode(ctx, "identity_id")
	require.NoError(t, err)

	isValid, identityId, err := oauth.ConsumeAuthCode(ctx, code)
	require.NoError(t, err)
	require.True(t, isValid)
	require.Equal(t, "identity_id", identityId)
}

func TestConsumeAuthCode_DoesNotExist(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	isValid, identityId, err := oauth.ConsumeAuthCode(ctx, "notexists")
	require.NoError(t, err)
	require.False(t, isValid)
	require.Empty(t, identityId)
}

func TestConsumeAuthCode_AlreadyConsumed(t *testing.T) {
	ctx, database, _ := keeltesting.MakeContext(t, authTestSchema, true)
	defer database.Close()

	code, err := oauth.NewAuthCode(ctx, "identity_id")
	require.NoError(t, err)

	isValid, identityId, err := oauth.ConsumeAuthCode(ctx, code)
	require.NoError(t, err)
	require.True(t, isValid)
	require.Equal(t, "identity_id", identityId)

	isValid, identityId, err = oauth.ConsumeAuthCode(ctx, code)
	require.NoError(t, err)
	require.False(t, isValid)
	require.Empty(t, identityId)
}
