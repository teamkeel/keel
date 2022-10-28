package actions

import (
	"context"
	"errors"
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
func Authenticate(ctx context.Context, schema *proto.Schema, args *AuthenticateArgs) (*ksuid.KSUID, bool, string, error) {
	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, false, "", err
	}

	identity, err := find(ctx, args.Email)

	if err != nil {
		return nil, false, "", err
	}

	if identity != nil {
		authenticated := bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(args.Password)) == nil

		if authenticated {
			id, err := ksuid.Parse(identity.Id)

			if err != nil {
				return nil, false, "", err
			}

			token, err := generateBearerToken(&id)

			return &id, false, token, nil
		} else {
			return nil, false, "", nil
		}
	} else if args.CreateIfNotExists {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)

		if err != nil {
			return nil, false, "", err
		}

		identityModel := proto.FindModel(schema.Models, parser.ImplicitIdentityModelName)

		modelMap, err := initialValueForModel(identityModel, schema)
		if err != nil {
			return nil, false, "", err
		}

		modelMap[strcase.ToSnake(EmailColumnName)] = args.Email
		modelMap[strcase.ToSnake(PasswordColumnName)] = string(hashedBytes)

		if err := db.Table(strcase.ToSnake(identityModel.Name)).Create(modelMap).Error; err != nil {
			return nil, false, "", err
		}

		id := modelMap[IdColumnName].(ksuid.KSUID)

		token, err := generateBearerToken(&id)

		return &id, true, token, nil
	}

	return nil, false, "", nil
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

func generateBearerToken(id *ksuid.KSUID) (string, error) {
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
