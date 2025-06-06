package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

const (
	KeelIssuer                          = "https://keel.so"
	resetPasswordAudClaim               = "password-reset"
	ResetTokenExpiry      time.Duration = time.Minute * 15
)

var (
	ErrInvalidToken     = common.NewAuthenticationFailedMessageErr("cannot be parsed or verified as a valid JWT")
	ErrTokenExpired     = common.NewAuthenticationFailedMessageErr("token has expired")
	ErrIdentityNotFound = common.NewAuthenticationFailedMessageErr("identity not found")
)

type AccessTokenClaims struct {
	jwt.RegisteredClaims // https://pkg.go.dev/github.com/golang-jwt/jwt/v4#RegisteredClaims
}

func GenerateAccessToken(ctx context.Context, identityId string) (string, time.Duration, error) {
	if identityId == "" {
		return "", 0, errors.New("cannot generate access token with an empty identityId intended for the sub claim")
	}

	config, err := runtimectx.GetOAuthConfig(ctx)
	if err != nil {
		return "", 0, err
	}

	expiry := config.AccessTokenExpiry()

	token, err := generateToken(ctx, identityId, []string{}, expiry)
	if err != nil {
		return "", 0, err
	}

	return token, expiry, nil
}

func ValidateAccessToken(ctx context.Context, tokenString string) (string, error) {
	return validateToken(ctx, tokenString, "")
}

func GenerateResetToken(ctx context.Context, identityId string) (string, error) {
	if identityId == "" {
		return "", errors.New("cannot generate access token with an empty identityId intended for the sub claim")
	}

	return generateToken(ctx, identityId, []string{resetPasswordAudClaim}, ResetTokenExpiry)
}

func ValidateResetToken(ctx context.Context, tokenString string) (string, error) {
	subject, err := validateToken(ctx, tokenString, resetPasswordAudClaim)

	return subject, err
}

func generateToken(ctx context.Context, sub string, aud []string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Audience:  aud,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    KeelIssuer,
		},
	}

	privateKey, err := runtimectx.GetPrivateKey(ctx)
	if err != nil {
		return "", err
	}

	if privateKey == nil {
		return "", fmt.Errorf("no private key set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("cannot create signed jwt: %w", err)
	}
	return tokenString, nil
}

func validateToken(ctx context.Context, tokenString string, audienceClaim string) (string, error) {
	ctx, span := tracer.Start(ctx, "Validate access token")
	defer span.End()

	privateKey, err := runtimectx.GetPrivateKey(ctx)
	if err != nil {
		return "", err
	}

	if privateKey == nil {
		return "", errors.New("no private key set")
	}

	var token *jwt.Token
	claims := &AccessTokenClaims{}

	token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	var validationErr *jwt.ValidationError
	if errors.As(err, &validationErr) && validationErr.Errors == jwt.ValidationErrorExpired {
		return "", ErrTokenExpired
	}

	if err != nil {
		return "", ErrInvalidToken
	}

	if !claims.VerifyExpiresAt(time.Now().UTC(), true) {
		return "", ErrTokenExpired
	}

	if audienceClaim != "" {
		if !lo.Contains(claims.Audience, audienceClaim) {
			return "", ErrInvalidToken
		}
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	if claims.Subject == "" {
		return "", errors.New("subject claim cannot be empty")
	}

	if claims.Issuer != KeelIssuer {
		return "", errors.New("invalid issuer")
	}

	return claims.Subject, nil
}
