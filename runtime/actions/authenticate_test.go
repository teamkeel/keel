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
	"github.com/teamkeel/keel/testhelpers"
)

func newContextWithPK() context.Context {
	ctx := context.Background()

	pk, _ := testhelpers.GetEmbeddedPrivateKey()
	ctx = runtimectx.WithPrivateKey(ctx, pk)

	return ctx
}

func TestBearerTokenGenerationAndParsingWithoutPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	ctx = runtimectx.WithEnv(ctx, runtimectx.KeelEnvTest)
	identityId := ksuid.New()

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, _, err := actions.ValidateBearerToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestBearerTokenGenerationAndParsingWithSamePrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, _, err := actions.ValidateBearerToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, identityId.String(), parsedId)
}

func TestBearerTokenGenerationWithPrivateKeyAndParsingWithoutPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	ctx = newContextWithPK()
	parsedId, _, err := actions.ValidateBearerToken(ctx, bearerJwt)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestBearerTokenGenerationWithoutPrivateKeyAndParsingWithPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	parsedId, _, err := actions.ValidateBearerToken(ctx, bearerJwt)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestBearerTokenGenerationAndParsingWithDifferentPrivateKeys(t *testing.T) {
	ctx := newContextWithPK()
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

	parsedId, _, err := actions.ValidateBearerToken(ctx, bearerJwt)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestBearerTokenIsRSAMethodWithPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
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

func TestBearerTokenHasExpiryClaims(t *testing.T) {
	ctx := newContextWithPK()
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
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	// Create the jwt 1 second expired.
	now := time.Now().UTC().Add(-actions.DefaultBearerTokenExpiry).Add(time.Second * -1)
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

	parsedId, _, err := actions.ValidateBearerToken(ctx, tokenString)
	require.ErrorIs(t, actions.ErrTokenExpired, err)
	require.Empty(t, parsedId)
}

func TestResetTokenGenerationAndParsingWithoutPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	ctx = runtimectx.WithEnv(ctx, runtimectx.KeelEnvTest)
	identityId := ksuid.New()

	bearerJwt, err := actions.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := actions.ValidateResetToken(ctx, bearerJwt)
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

	bearerJwt, err := actions.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	parsedId, err := actions.ValidateResetToken(ctx, bearerJwt)
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

	bearerJwt, err := actions.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey2)
	require.NoError(t, err)

	parsedId, err := actions.ValidateResetToken(ctx, bearerJwt)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestResetTokenIsRSAMethodWithPrivateKey(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, err := actions.GenerateResetToken(ctx, identityId.String())
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

func TestResetTokenHasExpiryClaims(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	jwtToken, err := actions.GenerateResetToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	claims := &actions.Claims{}
	_, err = jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	issuedAt := claims.IssuedAt.Time
	expiry := claims.ExpiresAt.Time

	require.Greater(t, expiry, time.Now().UTC())
	require.Equal(t, issuedAt.Add(actions.ResetTokenExpiry), expiry)
}

func TestExpiredResetTokenIsInvalid(t *testing.T) {
	ctx := newContextWithPK()
	identityId := ksuid.New()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	// Create the jwt 1 second expired.
	now := time.Now().UTC().Add(-actions.ResetTokenExpiry).Add(time.Second * -1)
	claims := actions.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identityId.String(),
			Audience:  jwt.ClaimStrings{"password-reset"},
			ExpiresAt: jwt.NewNumericDate(now.Add(actions.ResetTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedId, err := actions.ValidateResetToken(ctx, tokenString)
	require.ErrorIs(t, actions.ErrTokenExpired, err)
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
	claims := actions.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identityId.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(actions.ResetTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedId, err := actions.ValidateResetToken(ctx, tokenString)
	require.ErrorIs(t, actions.ErrInvalidToken, err)
	require.Empty(t, parsedId)
}

func TestBearerTokenIssueClaimIsKeel(t *testing.T) {
	ctx := newContextWithPK()
	ctx = runtimectx.WithEnv(ctx, runtimectx.KeelEnvTest)

	identityId := ksuid.New()

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	bearerJwt, err := actions.GenerateBearerToken(ctx, identityId.String())
	require.NoError(t, err)
	require.NotEmpty(t, bearerJwt)

	_, issuer, err := actions.ValidateBearerToken(ctx, bearerJwt)
	require.NoError(t, err)
	require.Equal(t, "keel", issuer)
}

func TestBearerTokenFromThirdParty(t *testing.T) {
	issuer := "https://enhanced-osprey-20.clerk.accounts.dev"

	ctx := newContextWithPK()
	ctx = runtimectx.WithEnv(ctx, runtimectx.KeelEnvTest)

	identityId := "user_2OdykNxqHGHNtBA5Hcdu5Zm6vDp"

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	require.NoError(t, err)

	// Create the jwt with third party claims
	now := time.Now().UTC()
	claims := actions.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "a98554836cb9880557ce",
			Subject:   identityId,
			ExpiresAt: jwt.NewNumericDate(now.Add(actions.ResetTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedId, parsedIssuer, err := actions.ValidateBearerToken(ctx, tokenString)
	require.NoError(t, err)
	require.Equal(t, identityId, parsedId)
	require.Equal(t, issuer, parsedIssuer)
}
