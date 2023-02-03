package actions

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/karlseguin/typed"
	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"

	"github.com/iancoleman/strcase"

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
	ErrInvalidToken         = errors.New("cannot be parsed or vertified as a valid JWT")
	ErrTokenExpired         = errors.New("token has expired")
	ErrInvalidIdentityClaim = errors.New("the identity claim is invalid and cannot be parsed")
	ErrIdentityNotFound     = errors.New("identity does not exist")
)

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

		token, err := GenerateBearerToken(id.String())
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

	modelMap[strcase.ToSnake(EmailColumnName)] = emailPassword.String("email")
	modelMap[strcase.ToSnake(PasswordColumnName)] = string(hashedBytes)

	query := NewQuery(identityModel)
	query.AddWriteValues(modelMap)
	query.AppendSelect(AllFields())
	query.AppendReturning(IdField())

	_, err = query.InsertStatement().Execute(scope.context)
	if err != nil {
		return nil, err
	}

	id := modelMap[IdColumnName].(string)

	token, err := GenerateBearerToken(id)
	if err != nil {
		return nil, err
	}

	return &AuthenticateResult{
		Token:           token,
		IdentityCreated: true,
	}, nil
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
type claims struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}

func GenerateBearerToken(id string) (string, error) {
	now := time.Now()

	claims := claims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(getSigningKey())

	return tokenString, err
}

func ParseBearerToken(jwtToken string) (string, error) {
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

	ksuid, err := ksuid.Parse(claims.Id)

	if err != nil {
		return "", ErrInvalidIdentityClaim
	}

	return ksuid.String(), nil
}

func getSigningKey() []byte {
	// TODO: make this a configuration to the runtime
	return []byte("test")
}
