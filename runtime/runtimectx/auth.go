package runtimectx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type contextKey string

const (
	authContextKey        contextKey = "authConfig"
	ExternalIssuersEnvKey string     = "KEEL_EXTERNAL_ISSUERS"
)

type AuthConfig struct {
	// If enabled, will verify tokens using any OIDC compatible issuer
	AllowAnyIssuers bool             `json:"AllowAllIssuers"`
	Issuers         []ExternalIssuer `json:"issuers"`
	Keel            *KeelAuthConfig  `json:"keel"`
}

type KeelAuthConfig struct {
	// Allow new identities to be created through the authenticate endpoint
	AllowCreate bool `json:"allowCreate"`
	// In seconds
	TokenDuration int `json:"tokenDuration"`
}

type ExternalIssuer struct {
	Iss      string  `json:"iss"`
	Audience *string `json:"audience"`
}

func WithAuthConfig(ctx context.Context, config AuthConfig) context.Context {
	ctx = context.WithValue(ctx, authContextKey, config)
	return ctx
}

func GetAuthConfig(ctx context.Context) (*AuthConfig, error) {

	v := ctx.Value(authContextKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", authContextKey)
	}

	config, ok := v.(AuthConfig)

	if !ok {
		return nil, errors.New("auth config in the context has wrong value type")
	}
	return &config, nil
}

// Backwards compatibility with the previous env var config.
func WithIssuersFromEnv(ctx context.Context) context.Context {
	envVar := os.Getenv(ExternalIssuersEnvKey)

	if envVar == "" {
		return ctx
	}

	authConfig, _ := GetAuthConfig(ctx)

	if authConfig != nil && len(authConfig.Issuers) > 0 {
		// Already have known issuers
		return ctx
	}

	issuers := []ExternalIssuer{}

	for _, uri := range strings.Split(envVar, ",") {
		issuers = append(issuers, ExternalIssuer{
			Iss: uri,
		})
	}

	if authConfig == nil {
		authConfig = &AuthConfig{}
	}

	authConfig.Issuers = issuers

	ctx = WithAuthConfig(ctx, *authConfig)

	return ctx
}
