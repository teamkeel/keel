package runtimectx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/teamkeel/keel/runtime/auth"
)

const (
	authContextKey        contextKey = "authConfig"
	ExternalIssuersEnvKey string     = "KEEL_EXTERNAL_ISSUERS"
)

func WithAuthConfig(ctx context.Context, config auth.AuthConfig) context.Context {
	validIssuers := auth.CheckIssuers(ctx, config.Issuers)
	config.Issuers = validIssuers

	ctx = context.WithValue(ctx, authContextKey, config)

	return ctx
}

func GetAuthConfig(ctx context.Context) (*auth.AuthConfig, error) {

	v := ctx.Value(authContextKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", authContextKey)
	}

	config, ok := v.(auth.AuthConfig)

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

	newCtx, err := GetAuthConfig(ctx)
	if err != nil {
		return ctx
	}

	if newCtx != nil && len(newCtx.Issuers) > 0 {
		// Already have known issuers
		return ctx
	}

	issuers := []auth.ExternalIssuer{}

	for _, uri := range strings.Split(envVar, ",") {
		issuers = append(issuers, auth.ExternalIssuer{
			Iss: uri,
		})
	}

	newCtx.Issuers = issuers

	ctx = WithAuthConfig(ctx, *newCtx)

	return ctx
}
