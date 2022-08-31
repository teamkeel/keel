package runtime

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/proto"
)

type FunctionsClient interface {
	Request(ctx context.Context, actionName string, opType proto.OperationType, body map[string]any) (any, error)
	ToGraphQL(ctx context.Context, response any, opType proto.OperationType) (interface{}, error)
}

type functionsClientContextKey string

var functionsClientKey functionsClientContextKey = "client"

func WithFunctionsClient(ctx context.Context, client FunctionsClient) context.Context {
	return context.WithValue(ctx, functionsClientKey, client)
}

func CallFunction(ctx context.Context, actionName string, opType proto.OperationType, body map[string]any) (any, error) {
	client, ok := ctx.Value(functionsClientKey).(FunctionsClient)
	if !ok {
		return nil, errors.New("no functions client in context")
	}

	return client.Request(ctx, actionName, opType, body)
}

func ToGraphQL(ctx context.Context, response any, opType proto.OperationType) (interface{}, error) {
	client, ok := ctx.Value(functionsClientKey).(FunctionsClient)
	if !ok {
		return nil, errors.New("no functions client in context")
	}

	return client.ToGraphQL(ctx, response, opType)
}
