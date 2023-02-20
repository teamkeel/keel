package runtimectx

import (
	"context"
	"errors"
	"fmt"
)

type secretContextKey string

const (
	SecretContextKey secretContextKey = "secret"
)

func GetSecret(ctx context.Context, secret string) (string, error) {
	v := ctx.Value(SecretContextKey)
	if v == nil {
		return "", fmt.Errorf("context does not have a :%s key", SecretContextKey)
	}

	secretFromCtx, ok := v.(map[string]string)[secret]
	if !ok {
		return "", errors.New("secret in the context has wrong value type")
	}
	return secretFromCtx, nil
}

func GetSecrets(ctx context.Context) map[string]string {
	return ctx.Value(SecretContextKey).(map[string]string)
}

func WithSecrets(ctx context.Context, secrets map[string]string) context.Context {
	return context.WithValue(ctx, SecretContextKey, secrets)
}
