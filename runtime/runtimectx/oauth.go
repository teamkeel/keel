package runtimectx

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/config"
)

const (
	oauthContextKey contextKey = "oauthConfig"
)

func WithOAuthConfig(ctx context.Context, config *config.AuthConfig) context.Context {
	ctx = context.WithValue(ctx, oauthContextKey, config)
	return ctx
}

func GetOAuthConfig(ctx context.Context) (*config.AuthConfig, error) {
	v := ctx.Value(oauthContextKey)
	if v == nil {
		return &config.AuthConfig{}, nil
	}

	config, ok := v.(*config.AuthConfig)

	if !ok {
		return nil, errors.New("auth config in the context has wrong value type")
	}
	return config, nil
}
