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
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"

	"github.com/iancoleman/strcase"

	"golang.org/x/crypto/bcrypt"
)

type Identity struct {
	Id       string `gorm:"column:id"`
	Email    string `gorm:"column:email"`
	Password string `gorm:"column:password"`
}

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
)

// Authenticate will return the identity ID if it is successfully authenticated or when a new identity is created.
func Authenticate(scope *Scope, input map[string]any) (*AuthenticateResult, error) {
	typedInput := typed.New(input)
	ctx := scope.context

	emailPassword := typedInput.Object("emailPassword")
	if _, err := mail.ParseAddress(emailPassword.String("email")); err != nil {
		return nil, errors.New("invalid email address")
	}

	if emailPassword.String("password") == "" {
		return nil, errors.New("password cannot be empty")
	}

	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	identity, err := find(ctx, emailPassword.String("email"))
	if err != nil {
		return nil, err
	}

	if identity != nil {
		authenticated := bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(emailPassword.String("password"))) == nil
		if !authenticated {
			return nil, errors.New("failed to authenticate")
		}

		id, err := ksuid.Parse(identity.Id)
		if err != nil {
			return nil, err
		}

		token, err := GenerateBearerToken(&id)
		if err != nil {
			return nil, err
		}

		return &AuthenticateResult{
			Token:           token,
			IdentityCreated: false,
		}, nil
	}

	if typedInput.Bool("createIfNotExists") {
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

		err = db.Table(strcase.ToSnake(identityModel.Name)).Create(modelMap).Error
		if err != nil {
			return nil, err
		}

		id := modelMap[IdColumnName].(ksuid.KSUID)

		token, err := GenerateBearerToken(&id)
		if err != nil {
			return nil, err
		}

		return &AuthenticateResult{
			Token:           token,
			IdentityCreated: true,
		}, nil
	} else {
		return nil, errors.New("failed to authenticate")
	}
}

func find(ctx context.Context, email string) (*Identity, error) {
	db, _ := runtimectx.GetDatabase(ctx)

	var identity Identity

	result := db.
		Table(strcase.ToSnake(parser.ImplicitIdentityModelName)).
		Limit(1).
		Where(EmailColumnName, email).
		Find(&identity)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &identity, nil
}

// https://pkg.go.dev/github.com/golang-jwt/jwt/v4#RegisteredClaims
type claims struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}

func GenerateBearerToken(id *ksuid.KSUID) (string, error) {
	now := time.Now()

	claims := claims{
		Id: id.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(getSigningKey())

	return tokenString, err
}

func ParseBearerToken(jwtToken string) (*ksuid.KSUID, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(getSigningKey()), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims := token.Claims.(*claims)

	if !claims.VerifyExpiresAt(time.Now(), true) {
		return nil, ErrTokenExpired
	}

	ksuid, err := ksuid.Parse(claims.Id)

	if err != nil {
		return nil, ErrInvalidIdentityClaim
	}

	return &ksuid, nil
}

func getSigningKey() []byte {
	return []byte("PLACEHOLDER_PRIVATE_KEY")
}
