package runtimectx

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/storage"
)

type storageContextKey string
type storageServerContextKey string

var storageKey storageContextKey = "storageService"
var storageServerKey storageServerContextKey = "storageServer"

// WithStorage adds the given storage service to the context.
func WithStorage(ctx context.Context, svc storage.Storer) context.Context {
	return context.WithValue(ctx, storageKey, svc)
}

// HasStorage checks if the context has a Storer service.
func HasStorage(ctx context.Context) bool {
	return ctx.Value(storageKey) != nil
}

// WithStorageServer marks that the runtime ctx should have a storage server (used to retrieve storage files via HTTP).
func WithStorageServer(ctx context.Context) context.Context {
	return context.WithValue(ctx, storageServerKey, true)
}

// HasStorageServer checks if the context has a storage server.
func HasStorageServer(ctx context.Context) bool {
	return ctx.Value(storageServerKey) != nil
}

// GetStorage will return the Storer service from the context.
func GetStorage(ctx context.Context) (storage.Storer, error) {
	v, ok := ctx.Value(storageKey).(storage.Storer)
	if !ok {
		return nil, fmt.Errorf("context does not have key or is not a Storer: %s", storageKey)
	}
	return v, nil
}
