package oauth_test

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
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/testhelpers"
)

func newContextWithPK() context.Context {
	ctx := context.Background()

	pk, _ := testhelpers.GetEmbeddedPrivateKey()
	ctx = runtimectx.WithPrivateKey(ctx, pk)

	return ctx
}

func TestAccessTokenGeneration(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	bearerJwt, _, err := oauth.GenerateAccessToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := oauth.ValidateAccessToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestAccessTokenValidationNoPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	bearerJwt, _, err := oauth.GenerateAccessToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	ctx = runtimectx.WithPrivateKey(ctx, nil)

	parsedId, err := oauth.ValidateAccessToken(ctx, bearerJwt)
	require.Error(t, err, "no private key set")
	require.Empty(t, parsedId)
}

func TestAccessTokenGenerationAndParsingWithSamePrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	bearerJwt, _, err := oauth.GenerateAccessToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := oauth.ValidateAccessToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestAccessTokenValidationDifferentPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	bearerJwt, _, err := oauth.GenerateAccessToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	ctx = newContextWithPK()
	parsedId, err := oauth.ValidateAccessToken(ctx, bearerJwt)
	require.ErrorIs(t, oauth.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestAccessTokenIsRSAMethodWithPrivateKey(t *testing.T) {
	ctx := t.Context()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, _, err := oauth.GenerateAccessToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	_, err = jwt.ParseWithClaims(jwtToken, &oauth.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			assert.Fail(t, "Invalid signing method. Expected RSA.")
		}
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
}

func TestAccessTokenClaims(t *testing.T) {
	ctx := t.Context()
	identityId := ksuid.New()

	seconds := 360
	config := config.AuthConfig{
		Tokens: config.TokensConfig{
			AccessTokenExpiry: &seconds,
		},
	}
	ctx = runtimectx.WithOAuthConfig(ctx, &config)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, lifespan, err := oauth.GenerateAccessToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	claims := &oauth.IdTokenClaims{}
	_, err = jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	issuedAt := claims.IssuedAt.Time
	expiry := claims.ExpiresAt.Time
	subject := claims.Subject
	issuer := claims.Issuer

	require.Greater(t, expiry, time.Now())
	require.Equal(t, issuedAt.Add(time.Second*360), expiry)
	require.Equal(t, time.Second*360, lifespan)
	require.Equal(t, config.AccessTokenExpiry(), time.Second*360)
	require.Equal(t, subject, identityId.String())
	require.Equal(t, issuer, "https://keel.so")
}

func TestShortExpiredAccessTokenIsInvalid(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	seconds := 1
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Tokens: config.TokensConfig{
			AccessTokenExpiry: &seconds,
		},
	})

	bearerJwt, _, err := oauth.GenerateAccessToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	time.Sleep(1100 * time.Millisecond)

	parsedId, err := oauth.ValidateAccessToken(ctx, bearerJwt)
	require.ErrorIs(t, oauth.ErrTokenExpired, err)
	require.Empty(t, parsedId)
}

func TestExpiredAccessTokenIsInvalid(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	// Create the jwt 1 second expired.
	now := time.Now().UTC().Add(-config.DefaultAccessTokenExpiry).Add(time.Second * -1)
	claims := oauth.AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identityId.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedId, err := oauth.ValidateAccessToken(ctx, tokenString)
	require.ErrorIs(t, oauth.ErrTokenExpired, err)
	require.Empty(t, parsedId)
}

func TestResetTokenGenerationAndParsingWithoutPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	ctx = runtimectx.WithEnv(ctx, runtimectx.KeelEnvTest)
	identityId := ksuid.New()

	bearerJwt, err := oauth.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := oauth.ValidateResetToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestResetTokenGenerationAndParsingWithSamePrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	bearerJwt, err := oauth.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := oauth.ValidateResetToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestResetTokenGenerationAndParsingWithDifferentPrivateKeys(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey1)
	require.NoError(t, err)

	bearerJwt, err := oauth.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey2)
	require.NoError(t, err)

	parsedId, err := oauth.ValidateResetToken(ctx, bearerJwt)
	require.ErrorIs(t, oauth.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestResetTokenIsRSAMethodWithPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, err := oauth.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	_, err = jwt.ParseWithClaims(jwtToken, &oauth.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			assert.Fail(t, "Invalid signing method. Expected RSA.")
		}
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
}

func TestResetTokenHasExpiryClaims(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, err := oauth.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	claims := &oauth.AccessTokenClaims{}
	_, err = jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	issuedAt := claims.IssuedAt.Time
	expiry := claims.ExpiresAt.Time

	require.Greater(t, expiry, time.Now().UTC())
	require.Equal(t, issuedAt.Add(oauth.ResetTokenExpiry), expiry)
}

func TestExpiredResetTokenIsInvalid(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	// Create the jwt 1 second expired.
	now := time.Now().UTC().Add(-oauth.ResetTokenExpiry).Add(time.Second * -1)
	claims := oauth.AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identityId.String(),
			Audience:  jwt.ClaimStrings{"password-reset"},
			ExpiresAt: jwt.NewNumericDate(now.Add(oauth.ResetTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedId, err := oauth.ValidateResetToken(ctx, tokenString)
	require.ErrorIs(t, oauth.ErrTokenExpired, err)
	require.Empty(t, parsedId)
}

func TestResetTokenMissingAudIsInvalid(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	// Create the jwt with missing aud claim.
	now := time.Now().UTC()
	claims := oauth.AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identityId.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(oauth.ResetTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedId, err := oauth.ValidateResetToken(ctx, tokenString)
	require.ErrorIs(t, oauth.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}
