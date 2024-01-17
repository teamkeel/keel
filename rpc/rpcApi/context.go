package rpcApi

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/proto"
)

type schemaContextKey string

var schemaKey schemaContextKey = "schema"

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
