package actions

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

type AuthenticateArgs struct {
	CreateIfNotExists bool
	Email             string
	Password          string
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
func Authenticate(ctx context.Context, schema *proto.Schema, args *AuthenticateArgs) (string, bool, error) {
	if _, err := mail.ParseAddress(args.Email); err != nil {
		return "", false, errors.New("invalid email address")
	}

	if args.Password == "" {
		return "", false, errors.New("password cannot be empty")
	}

	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return "", false, err
	}

	identity, err := find(ctx, args.Email)

	if err != nil {
		return "", false, err
	}

	if identity != nil {
		authenticated := bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(args.Password)) == nil

		if authenticated {
			id, err := ksuid.Parse(identity.Id)

			if err != nil {
				return "", false, err
			}

			token, err := GenerateBearerToken(&id) // todo: check this error

			return token, false, nil
		} else {
			return "", false, nil
		}
	} else if args.CreateIfNotExists {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)

		if err != nil {
			return "", false, err
		}

		identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

		modelMap, err := initialValueForModel(identityModel, schema)
		if err != nil {
			return "", false, err
		}

		modelMap[strcase.ToSnake(EmailColumnName)] = args.Email
		modelMap[strcase.ToSnake(PasswordColumnName)] = string(hashedBytes)

		if err := db.Table(strcase.ToSnake(identityModel.Name)).Create(modelMap).Error; err != nil {
			return "", false, err
		}

		id := modelMap[IdColumnName].(ksuid.KSUID)

		token, err := GenerateBearerToken(&id) // todo: check this error

		return token, true, nil
	}

	return "", false, nil
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
