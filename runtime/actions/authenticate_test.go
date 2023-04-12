package actions_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

func TestGenerationAndParsingWithoutPrivateKey(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := actions.ParseBearerToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestGenerationAndParsingWithSamePrivateKey(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := actions.ParseBearerToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestGenerationWithPrivateKeyAndParsingWithoutPrivateKey(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	ctx = context.Background()
	parsedId, err := actions.ParseBearerToken(ctx, bearerJwt)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestGenerationWithoutPrivateKeyAndParsingWithPrivateKey(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	parsedId, err := actions.ParseBearerToken(ctx, bearerJwt)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestGenerationAndParsingWithDifferentPrivateKeys(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	privateKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey1)
	require.NoError(t, err)

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey2)
	require.NoError(t, err)

	parsedId, err := actions.ParseBearerToken(ctx, bearerJwt)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestBearerTokenIsRSAMethodWithPrivateKey(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	_, err = jwt.ParseWithClaims(jwtToken, &actions.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			assert.Fail(t, "Invalid signing method. Expected RSA.")
		}
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
}

func TestBearerTokenIsNoneMethodWithoutPrivateKey(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	jwtToken, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	_, err = jwt.ParseWithClaims(jwtToken, &actions.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Header["alg"] != "none" {
			assert.Fail(t, "Invalid signing method. Expected none.")
		}
		return jwt.UnsafeAllowNoneSignatureType, nil
	})
	require.NoError(t, err)
}

func TestBearerTokenHasExpiryClaims(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	claims := &actions.Claims{}

	_, err = jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	issuedAt := claims.IssuedAt.Time
	expiry := claims.ExpiresAt.Time

	require.Greater(t, expiry, time.Now())
	require.Equal(t, issuedAt.Add(24*time.Hour), expiry)
}

func TestExpiredBearerTokenIsInvalid(t *testing.T) {
	ctx := context.Background()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	// Create the jwt 25 hours in past, which means it is 1 hour expired.
	now := time.Now().UTC().Add(time.Hour * -25)
	claims := actions.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identityId.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedId, err := actions.ParseBearerToken(ctx, tokenString)
	require.ErrorIs(t, actions.ErrTokenExpired, err)
	require.Empty(t, parsedId)
}
