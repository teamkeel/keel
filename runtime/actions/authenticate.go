package actions

import (
	"context"
	"errors"
	"fmt"
	email "net/mail"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/karlseguin/typed"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"

	"github.com/teamkeel/keel/mail"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"

	"golang.org/x/crypto/bcrypt"
)

type AuthenticateResult struct {
	Token           string `json:"token"`
	IdentityCreated bool   `json:"identityCreated"`
}

var (
	ErrInvalidToken     = errors.New("cannot be parsed or vertified as a valid JWT")
	ErrTokenExpired     = errors.New("token has expired")
	ErrIdentityNotFound = errors.New("identity does not exist")
)

const (
	BearerTokenExpiry time.Duration = time.Hour * 24
	ResetTokenExpiry  time.Duration = time.Minute * 15
)

const (
	resetPasswordAudClaim = "password-reset"
	keelIssuerClaim       = "keel"
)

// https://pkg.go.dev/github.com/golang-jwt/jwt/v4#RegisteredClaims
type Claims struct {
	jwt.RegisteredClaims
}

// Authenticate will return the identity ID if it is successfully authenticated or when a new identity is created.
func Authenticate(scope *Scope, input map[string]any) (*AuthenticateResult, error) {
	typedInput := typed.New(input)

	emailPassword := typedInput.Object("emailPassword")
	if _, err := email.ParseAddress(emailPassword.String("email")); err != nil {
		return nil, common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid email address"}
	}

	if emailPassword.String("password") == "" {
		return nil, common.RuntimeError{Code: common.ErrInvalidInput, Message: "password cannot be empty"}
	}

	identity, err := FindIdentityByEmail(scope.Context, scope.Schema, emailPassword.String("email"))
	if err != nil {
		return nil, err
	}

	if identity != nil {
		authenticated := bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(emailPassword.String("password"))) == nil
		if !authenticated {
			return nil, common.RuntimeError{Code: common.ErrInvalidInput, Message: "failed to authenticate"}
		}

		id, err := ksuid.Parse(identity.Id)
		if err != nil {
			return nil, err
		}

		token, err := GenerateBearerToken(scope.Context, id.String())
		if err != nil {
			return nil, err
		}

		return &AuthenticateResult{
			Token:           token,
			IdentityCreated: false,
		}, nil
	}

	if !typedInput.Bool("createIfNotExists") {
		return nil, common.RuntimeError{Code: common.ErrInvalidInput, Message: "failed to authenticate"}
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(emailPassword.String("password")), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	identity, err = CreateIdentity(scope.Context, scope.Schema, emailPassword.String("email"), string(hashedBytes))
	if err != nil {
		return nil, err
	}

	token, err := GenerateBearerToken(scope.Context, identity.Id)
	if err != nil {
		return nil, err
	}

	return &AuthenticateResult{
		Token:           token,
		IdentityCreated: true,
	}, nil
}

func ResetRequestPassword(scope *Scope, input map[string]any) error {
	var err error
	typedInput := typed.New(input)

	emailString := typedInput.String("email")
	if _, err = email.ParseAddress(emailString); err != nil {
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid email address"}
	}

	var redirectUrl *url.URL
	if redirectUrl, err = url.ParseRequestURI(typedInput.String("redirectUrl")); err != nil {
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid redirect URL"}
	}

	var identity *runtimectx.Identity
	identity, err = FindIdentityByEmail(scope.Context, scope.Schema, emailString)
	if err != nil {
		return err
	}
	if identity == nil {
		return nil
	}

	token, err := GenerateResetToken(scope.Context, identity.Id)
	if err != nil {
		return err
	}

	q := redirectUrl.Query()
	q.Add("token", token)
	redirectUrl.RawQuery = q.Encode()

	client, err := runtimectx.GetMailClient(scope.Context)
	if err != nil {
		return err
	}

	err = client.Send(scope.Context, &mail.SendEmailRequest{
		To:        identity.Email,
		From:      "hi@keel.xyz",
		Subject:   "[Keel] Reset password request",
		PlainText: fmt.Sprintf("Please follow this link to reset your password: %s", redirectUrl),
	})

	return err
}

func ResetPassword(scope *Scope, input map[string]any) error {
	typedInput := typed.New(input)

	token := typedInput.String("token")
	password := typedInput.String("password")

	identityId, err := ValidateResetToken(scope.Context, token)
	switch {
	case errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired):
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: err.Error()}
	case err != nil:
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	identityModel := proto.FindModel(scope.Schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)
	err = query.Where(Field("id"), Equals, Value(identityId))
	if err != nil {
		return err
	}

	query.AddWriteValue("password", string(hashedPassword))

	affected, err := query.UpdateStatement().Execute(scope.Context)
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("expected 1 row to be updated, but %v rows were updated", affected)
	}

	return nil
}

func GenerateBearerToken(ctx context.Context, identityId string) (string, error) {
	return generateToken(ctx, identityId, []string{}, BearerTokenExpiry)
}

func GenerateResetToken(ctx context.Context, identityId string) (string, error) {
	return generateToken(ctx, identityId, []string{resetPasswordAudClaim}, ResetTokenExpiry)
}

func generateToken(ctx context.Context, sub string, aud []string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Audience:  aud,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    keelIssuerClaim,
		},
	}

	privateKey, err := runtimectx.GetPrivateKey(ctx)
	if err != nil {
		return "", err
	}

	if privateKey != nil {
		// If the private key is set, sign the token with the private key.
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(privateKey)
		if err != nil {
			return "", fmt.Errorf("cannot create signed jwt: %w", err)
		}
		return tokenString, nil
	} else {
		// If the private key is not set, do not sign the token.
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			return "", fmt.Errorf("cannot create unsecured jwt: %w", err)
		}
		return tokenString, nil
	}
}

// Verifies the bearer token and returns the JWT subject and issuer.
func ValidateBearerToken(ctx context.Context, tokenString string) (string, string, error) {
	return validateToken(ctx, tokenString, "")
}

// Verifies the reset token and returns the JWT subject.
func ValidateResetToken(ctx context.Context, tokenString string) (string, error) {
	subject, issuer, err := validateToken(ctx, tokenString, resetPasswordAudClaim)
	if issuer != keelIssuerClaim && issuer != "" {
		return "", fmt.Errorf("can only accept reset token from %s issuer, not: %s", keelIssuerClaim, issuer)
	}
	return subject, err
}

func validateToken(ctx context.Context, tokenString string, audienceClaim string) (string, string, error) {
	privateKey, err := runtimectx.GetPrivateKey(ctx)
	if err != nil {
		return "", "", err
	}

	var token *jwt.Token
	claims := &Claims{}
	if privateKey != nil {
		// If the private key is set, parse the token with the public key.
		token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
			}

			return &privateKey.PublicKey, nil
		})
	} else {
		// If the private key is not set, parse the token without the signature.
		token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if t.Header["alg"] != "none" {
				return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
			}

			return jwt.UnsafeAllowNoneSignatureType, nil
		})
	}

	if !claims.VerifyExpiresAt(time.Now().UTC(), true) {
		return "", "", ErrTokenExpired
	}

	if audienceClaim != "" {
		if !lo.Contains(claims.Audience, audienceClaim) {
			return "", "", ErrInvalidToken
		}
	}

	if err != nil || !token.Valid {
		return "", "", ErrInvalidToken
	}

	if claims.Issuer == keelIssuerClaim || claims.Issuer == "" {
		identifier, err := ksuid.Parse(claims.Subject)
		if err != nil {
			return "", "", err
		}
		return identifier.String(), claims.Issuer, nil
	} else {
		if claims.Subject == "" {
			return "", "", errors.New("subject claim cannot be empty")
		}
		return claims.Subject, claims.Issuer, nil
	}
}
