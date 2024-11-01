package rpcApi

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
)

type contextKey string

var schemaKey contextKey = "schema"
var configKey contextKey = "config"
var traceVerbosityKey contextKey = "verboseTraces"

func GetSchema(ctx context.Context) (*proto.Schema, error) {
	v := ctx.Value(schemaKey)
	schema, ok := v.(*proto.Schema)

	if !ok {
		return nil, errors.New("database in the context has wrong value type")
	}
	return schema, nil
}

func WithSchema(ctx context.Context, schema *proto.Schema) context.Context {
	return context.WithValue(ctx, schemaKey, schema)
}

func GetConfig(ctx context.Context) (*config.ProjectConfig, error) {
	v := ctx.Value(configKey)
	schema, ok := v.(*config.ProjectConfig)

	if !ok {
		return nil, errors.New("database in the context has wrong value type")
	}
	return schema, nil
}

func WithConfig(ctx context.Context, schema *config.ProjectConfig) context.Context {
	return context.WithValue(ctx, configKey, schema)
}

func WithTraceVerbosity(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, traceVerbosityKey, verbose)
}

func GetTraceVerbosity(ctx context.Context) bool {
	v := ctx.Value(traceVerbosityKey)
	verbose, has := v.(bool)

	if !has {
		return false
	}
	return verbose
}
