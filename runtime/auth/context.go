package auth

import (
	"context"
	"fmt"
)

type contextKey string

const (
	identityContextKey contextKey = "identityId"
)

type Identity map[string]any

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	if identity != nil {
		ctx = context.WithValue(ctx, identityContextKey, identity)
	}

	return ctx
}

func GetIdentity(ctx context.Context) (Identity, error) {
	v, ok := ctx.Value(identityContextKey).(Identity)
	if !ok {
		return nil, fmt.Errorf("context does not have a key or is not Identity: %s", identityContextKey)
	}
	return v, nil
}

func IsAuthenticated(ctx context.Context) bool {
	return ctx.Value(identityContextKey) != nil
}
