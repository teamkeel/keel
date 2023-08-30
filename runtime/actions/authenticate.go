package actions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	email "net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/karlseguin/typed"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/teamkeel/keel/mail"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
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
	ErrInvalidToken     = common.NewAuthenticationFailedMessageErr("cannot be parsed or verified as a valid JWT")
	ErrTokenExpired     = common.NewAuthenticationFailedMessageErr("token has expired")
	ErrIdentityNotFound = common.NewAuthenticationFailedMessageErr("identity not found")
)

const (
	DefaultBearerTokenExpiry time.Duration = time.Hour * 24
	ResetTokenExpiry         time.Duration = time.Minute * 15
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

	config, err := runtimectx.GetAuthConfig(scope.Context)
	if err == nil {
		if config != nil && config.Keel != nil && !config.Keel.AllowCreate {
			// Creating new identities is disabled
			return nil, common.RuntimeError{Code: common.ErrInvalidInput, Message: "registration disabled"}
		}
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
	expiry := DefaultBearerTokenExpiry
	config, err := runtimectx.GetAuthConfig(ctx)
	if err == nil {
		if config != nil && config.Keel != nil {
			expiry = time.Duration(config.Keel.TokenDuration) * time.Second
		}
	}

	return generateToken(ctx, identityId, []string{}, expiry)
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

	if privateKey == nil {
		return "", "", errors.New("no private key set")
	}

	ctx, span := tracer.Start(ctx, "Validate token")
	defer span.End()

	authConfig, _ := runtimectx.GetAuthConfig(ctx)

	var token *jwt.Token
	claims := &Claims{}

	// try to decode the token and validate using our private key as the signing method.
	// this supports external issued tokens but which are signed with our private key (such as clerk)
	span.AddEvent("Validating with Keel pk")
	token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	// if unsuccessful using our private key, try to validate against any of the known external issuers if there are any
	if err != nil {
		token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			iss := t.Claims.(*Claims).Issuer

			if authConfig == nil {
				// If no auth config skip all this
				return nil, nil
			}

			issuers := authConfig.Issuers

			if authConfig.AllowAnyIssuers {
				span.AddEvent("Accepting any issuer as AllowAnyIssuers is enabled")
				// In this mode we allow any issuer that is openID connect compatible
				// So if this issuer is new add it to the know issuers and verify

				match := false
				for _, extIssuers := range issuers {
					if extIssuers.Iss == iss {
						match = true
					}
				}
				if !match {
					issuers = append(issuers, auth.ExternalIssuer{
						Iss: iss,
					})
				}

				authConfig.Issuers = issuers
				ctx = runtimectx.WithAuthConfig(ctx, *authConfig)

			}

			if len(issuers) > 0 {
				span.AddEvent("Validating with external issuers")
			}

			for _, issuer := range issuers {
				if issuer.Iss != iss {
					continue
				}
				iss := t.Claims.(*Claims).Issuer

				kid := ""
				if header, ok := t.Header["kid"]; ok {
					if kidStr, ok := header.(string); ok {
						kid = kidStr
					}
				}

				publicKey, err := auth.PublicKeyForIssuer(ctx, iss, kid)

				// Check the audience matches if set
				match := true
				if issuer.Audience != nil {
					match = false
					aud := t.Claims.(*Claims).Audience
					for _, audience := range aud {
						if audience == *issuer.Audience {
							match = true
							break
						}
					}
					span.AddEvent("Validating audience claims", trace.WithAttributes(attribute.Bool("match", match)))
				}

				if !match {
					return nil, fmt.Errorf("invalid token")
				}

				if err == nil {
					return publicKey, nil
				}
			}

			return nil, fmt.Errorf("unexpected issuer in token: %s", iss)
		})
	}

	if err != nil {
		return "", "", ErrInvalidToken
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

func HandleAuthorizationHeader(ctx context.Context, schema *proto.Schema, headers http.Header) (*runtimectx.Identity, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return nil, nil
	}

	headerSplit := strings.Split(header, "Bearer ")
	if len(headerSplit) != 2 {
		return nil, common.NewAuthenticationFailedMessageErr("no 'Bearer' prefix in the Authorization header")
	}

	token := headerSplit[1]

	if token != "" {
		identity, err := HandleBearerToken(ctx, schema, token)
		if err != nil {
			return nil, err
		}
		return identity, nil
	}

	return nil, nil
}

func HandleBearerToken(ctx context.Context, schema *proto.Schema, token string) (*runtimectx.Identity, error) {
	ctx, span := tracer.Start(ctx, "Authorization")
	defer span.End()

	span.SetAttributes(attribute.String("token", token))

	subject, issuer, err := ValidateBearerToken(ctx, token)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Check that identity actually does exist as it could
	// have been deleted after the bearer token was generated.
	var identity *runtimectx.Identity
	if issuer == "keel" || issuer == "" {
		identity, err = FindIdentityById(ctx, schema, subject)
	} else {
		identity, err = FindIdentityByExternalId(ctx, schema, subject, issuer)
		if identity == nil {
			identity, err = CreateExternalIdentity(ctx, schema, subject, issuer, token)
		}
	}

	if err != nil {
		return nil, err
	}

	if identity == nil {
		return nil, ErrIdentityNotFound
	}

	span.SetAttributes(attribute.String("identity.id", identity.Id))

	return identity, nil
}
