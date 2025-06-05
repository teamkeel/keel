package runtimectx

import (
	"context"
)

type KeelEnv string

const (
	// The Test environment denotes any Keel environment that isn't production or staging. So this includes the
	// runtime inside of the Keel Test Framework runner, as well as the runtime locally (e.g 'keel run' in the CLI).
	KeelEnvTest       KeelEnv = "test"
	KeelEnvProduction KeelEnv = "production"
)

type EnvKeyContextType string

var envKeyContext EnvKeyContextType = "env"

func GetEnv(ctx context.Context) KeelEnv {
	v := ctx.Value(envKeyContext)

	// if there's nothing in the context, then we default to 'test'
	if v == nil {
		return KeelEnvTest
	}

	env, ok := v.(KeelEnv)

	// If no valid KeelEnv can be unmarshaled then we default to 'test'
	if !ok {
		return KeelEnvTest
	}

	return env
}

func WithEnv(ctx context.Context, env KeelEnv) context.Context {
	return context.WithValue(ctx, envKeyContext, env)
}
