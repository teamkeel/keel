package runtimectx

import (
	"context"
)

type KeelEnv string

const (
	KeelEnvTest    KeelEnv = "test"
	KeelEnvDefault KeelEnv = "default"
)

var envKeyContext string = "env"

func GetEnv(ctx context.Context) KeelEnv {
	v := ctx.Value(envKeyContext)

	if v == nil {
		return KeelEnvDefault
	}

	env, ok := v.(KeelEnv)

	if !ok {
		return KeelEnvDefault
	}

	return env
}

func WithEnv(ctx context.Context, env KeelEnv) context.Context {
	return context.WithValue(ctx, envKeyContext, env)
}
