package oauth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

func ExtractClaimFromJwt(token string, claim string) (string, error) {
	// Parse the JWT without verifying the signature
	t, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("error parsing JWT: %s", err)
	}

	// Extract the claim
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("claims not found")

	}

	value, ok := claims[claim].(string)
	if !ok {
		return "", fmt.Errorf("%s claim not found or not a string", claim)
	}

	return value, nil
}
