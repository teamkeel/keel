package runtimectx

import (
	"context"
	"fmt"
)

const (
	environmentVariablesContextKey contextKey = "environmentVariables"
)

func GetEnvironmentVariables(ctx context.Context) (map[string][]string, error) {
	v, ok := ctx.Value(environmentVariablesContextKey).(map[string][]string)
	if !ok {
		return nil, fmt.Errorf("context does not have key or is not map[string]string: %s", environmentVariablesContextKey)
	}
	return v, nil
}

func WithEnvironmentVariables(ctx context.Context, env map[string]string) context.Context {
	if env != nil {
		ctx = context.WithValue(ctx, environmentVariablesContextKey, env)
	}

	return ctx
}
