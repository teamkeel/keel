package actions

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/karlseguin/typed"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
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

const (
	IdColumnName       string = "id"
	EmailColumnName    string = "email"
	PasswordColumnName string = "password"
)

var (
	ErrInvalidToken     = errors.New("cannot be parsed or vertified as a valid JWT")
	ErrTokenExpired     = errors.New("token has expired")
	ErrIdentityNotFound = errors.New("identity does not exist")
)

const (
	bearerTokenExpiry time.Duration = time.Hour * 24
	resetTokenExpiry  time.Duration = time.Minute * 15
)

const resetPasswordAudClaim = "password-reset"

// Authenticate will return the identity ID if it is successfully authenticated or when a new identity is created.
func Authenticate(scope *Scope, input map[string]any) (*AuthenticateResult, error) {
	typedInput := typed.New(input)

	emailPassword := typedInput.Object("emailPassword")
	if _, err := mail.ParseAddress(emailPassword.String("email")); err != nil {
		return nil, common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid email address"}
	}

	if emailPassword.String("password") == "" {
		return nil, common.RuntimeError{Code: common.ErrInvalidInput, Message: "password cannot be empty"}
	}

	identity, err := FindIdentityByEmail(scope.context, scope.schema, emailPassword.String("email"))
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

		token, err := GenerateToken(scope.context, id.String(), []string{}, bearerTokenExpiry)
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

	identityModel := proto.FindModel(scope.schema.Models, parser.ImplicitIdentityModelName)

	modelMap, err := initialValueForModel(identityModel, scope.schema)
	if err != nil {
		return nil, err
	}

	modelMap[EmailColumnName] = emailPassword.String("email")
	modelMap[PasswordColumnName] = string(hashedBytes)

	query := NewQuery(identityModel)
	query.AddWriteValues(modelMap)
	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	_, err = query.InsertStatement().Execute(scope.context)
	if err != nil {
		return nil, err
	}

	id := modelMap[IdColumnName].(string)

	token, err := GenerateToken(scope.context, id, []string{}, bearerTokenExpiry)
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

	email := typedInput.String("email")
	if _, err = mail.ParseAddress(email); err != nil {
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid email address"}
	}

	var redirectUrl *url.URL
	if redirectUrl, err = url.ParseRequestURI(typedInput.String("redirectUrl")); err != nil {
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: "invalid redirect URL"}
	}

	var identity *runtimectx.Identity
	identity, err = FindIdentityByEmail(scope.context, scope.schema, email)
	if err != nil {
		return err
	}
	if identity == nil {
		return nil
	}

	token, err := GenerateToken(identity.Id, []string{resetPasswordAudClaim}, resetTokenExpiry)
	if err != nil {
		return err
	}

	q := redirectUrl.Query()
	q.Add("token", token)
	redirectUrl.RawQuery = q.Encode()

	client, err := runtimectx.GetMailClient(scope.context)
	if err != nil {
		return err
	}

	err = client.SendResetPasswordMail(identity.Email, redirectUrl.String())
	if err != nil {
		return err
	}

	return nil
}

func ResetPassword(scope *Scope, input map[string]any) error {
	typedInput := typed.New(input)

	token := typedInput.String("token")
	password := typedInput.String("password")

	identityId, err := ParseResetToken(token)
	switch {
	case errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrInvalidIdentityClaim):
		return common.RuntimeError{Code: common.ErrInvalidInput, Message: err.Error()}
	case err != nil:
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	identityModel := proto.FindModel(scope.schema.Models, parser.ImplicitIdentityModelName)

	query := NewQuery(identityModel)
	err = query.Where(Field("id"), Equals, Value(identityId))
	if err != nil {
		return err
	}

	query.AddWriteValue("password", string(hashedPassword))

	affected, err := query.UpdateStatement().Execute(scope.context)
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("expected 1 row to be updated, but %v rows were updated", affected)
	}

	return nil
}

func FindIdentityById(ctx context.Context, schema *proto.Schema, id string) (*runtimectx.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)
	query := NewQuery(identityModel)
	err := query.Where(IdField(), Equals, Value(id))
	if err != nil {
		return nil, err
	}

	query.AppendSelect(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return &runtimectx.Identity{
		Id:        result["id"].(string),
		Email:     result["email"].(string),
		Password:  result["password"].(string),
		CreatedAt: result["createdAt"].(time.Time),
		UpdatedAt: result["updatedAt"].(time.Time),
	}, nil
}

func FindIdentityByEmail(ctx context.Context, schema *proto.Schema, email string) (*runtimectx.Identity, error) {
	identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)
	query := NewQuery(identityModel)
	err := query.Where(Field(EmailColumnName), Equals, Value(email))
	if err != nil {
		return nil, err
	}

	query.AppendSelect(AllFields())
	result, err := query.SelectStatement().ExecuteToSingle(ctx)

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return &runtimectx.Identity{
		Id:        result["id"].(string),
		Email:     result["email"].(string),
		Password:  result["password"].(string),
		CreatedAt: result["createdAt"].(time.Time),
		UpdatedAt: result["updatedAt"].(time.Time),
	}, nil
}

// https://pkg.go.dev/github.com/golang-jwt/jwt/v4#RegisteredClaims
type Claims struct {
	jwt.RegisteredClaims
}

func GenerateToken(ctx context.Context, sub string, aud []string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Audience:  aud,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
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

func ParseBearerToken(ctx context.Context, tokenString string) (string, error) {
	privateKey, err := runtimectx.GetPrivateKey(ctx)
	if err != nil {
		return "", err
	}

	var token *jwt.Token
	if privateKey != nil {
		// If the private key is set, parse the token with the public key.
		token, err = jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
			}

			return &privateKey.PublicKey, nil
		})
	} else {
		// If the private key is not set, parse the token without the signature.
		token, err = jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			if t.Header["alg"] != "none" {
				return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
			}

			return jwt.UnsafeAllowNoneSignatureType, nil
		})
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", ErrInvalidToken
	}

	if !claims.VerifyExpiresAt(time.Now(), true) {
		return "", ErrTokenExpired
	}

	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	ksuid, err := ksuid.Parse(claims.Subject)
	if err != nil {
		return "", errors.New("token does not contain a parsable subject claim")
	}

	return ksuid.String(), nil
}

func ParseResetToken(jwtToken string) (string, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return getSigningKey(), nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	claims := token.Claims.(*claims)

	if !claims.VerifyExpiresAt(time.Now(), true) {
		return "", ErrTokenExpired
	}

	if !lo.Contains(claims.Audience, resetPasswordAudClaim) {
		return "", ErrInvalidToken
	}

	ksuid, err := ksuid.Parse(claims.Subject)

	if err != nil {
		return "", ErrInvalidIdentityClaim
	}

	return ksuid.String(), nil
}

func getSigningKey() []byte {
	// TODO: make this a configuration to the runtime
	return []byte("test")
}
