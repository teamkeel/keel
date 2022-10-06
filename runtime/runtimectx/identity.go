package runtimectx

import (
	"context"
	"errors"
	"fmt"

	"github.com/segmentio/ksuid"
)

type contextKey string

const (
	identityIdContextKey contextKey = "identityId"
)

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

func IsAuthenticated(ctx context.Context) (bool, error) {
	return ctx.Value(identityIdContextKey) != nil, nil
}
