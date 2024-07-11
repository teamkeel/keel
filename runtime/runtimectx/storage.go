package runtimectx

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/storage"
)

type storageContextKey string

var storageKey storageContextKey = "storageService"

// WithStorage adds the given storage service to the context
func WithStorage(ctx context.Context, svc storage.Storer) context.Context {
	return context.WithValue(ctx, storageKey, svc)
}

// HasStorage checks if the context has a Storer service
func HasStorage(ctx context.Context) bool {
	return ctx.Value(storageKey) != nil
}

// GetStorage will return the Storer service from the context
func GetStorage(ctx context.Context) (storage.Storer, error) {
	v, ok := ctx.Value(storageKey).(storage.Storer)
	if !ok {
		return nil, fmt.Errorf("context does not have key or is not a Storer: %s", storageKey)
	}
	return v, nil
}
