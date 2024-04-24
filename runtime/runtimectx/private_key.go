package runtimectx

import (
	"context"
	"crypto/rsa"
	"errors"
)

type privateKeyContextKey string

var privateKeyContext privateKeyContextKey = "privateKey"

func GetPrivateKey(ctx context.Context) (*rsa.PrivateKey, error) {
	v := ctx.Value(privateKeyContext)
	if v == nil {
		return nil, nil
	}

	privateKey, ok := v.(*rsa.PrivateKey)

	if !ok {
		return nil, errors.New("private key in the context has wrong type")
	}
	return privateKey, nil
}

func WithPrivateKey(ctx context.Context, privateKey *rsa.PrivateKey) context.Context {
	return context.WithValue(ctx, privateKeyContext, privateKey)
}
