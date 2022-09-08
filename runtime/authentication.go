package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/segmentio/ksuid"
)

type contextKey string

const (
	identityIdContextKey    contextKey = "identityId"
	authorizationHeaderName string     = "Authorization"
)

var (
	ErrNoAuthorizationHeader = errors.New("no authentication header set")
	ErrNoBearerPrefix        = errors.New("no 'Bearer' prefix in the authentication header")
	ErrInvalidToken          = errors.New("cannot be parsed or vertified as a valid JWT")
	ErrTokenExpired          = errors.New("token has expired")
	ErrInvalidIdentityClaim  = errors.New("the identity claim is invalid and cannot be parsed")
)

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

func RetrieveIdentityClaim(request *http.Request) (*ksuid.KSUID, error) {
	header := request.Header.Get(authorizationHeaderName)
	if header == "" {
		return nil, ErrNoAuthorizationHeader
	}

	headerSplit := strings.Split(header, "Bearer ")
	if len(headerSplit) != 2 {
		return nil, ErrNoBearerPrefix
	}

	jwtToken := headerSplit[1]

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

func WithIdentity(ctx context.Context, id *ksuid.KSUID) context.Context {
	if id != nil {
		ctx = context.WithValue(ctx, identityIdContextKey, id)
	}

	return ctx
}

func GetIdentity(ctx context.Context) (*ksuid.KSUID, error) {
	v := ctx.Value(identityIdContextKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", identityIdContextKey)
	}

	id, ok := v.(*ksuid.KSUID)
	if !ok {
		return nil, errors.New("identity id on the context is not of type ksuid.KSUID")
	}
	return id, nil
}

func getSigningKey() []byte {
	return []byte("PLACEHOLDER_PRIVATE_KEY")
}
