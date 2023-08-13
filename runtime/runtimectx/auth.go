package runtimectx

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/runtime/auth"
)

const (
	authContextKey contextKey = "authConfig"
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
