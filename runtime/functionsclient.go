package runtime

import (
	"context"
	"errors"
)

type FunctionsClient interface {
	Request(ctx context.Context, actionName string, body map[string]any) (any, error)
}

type functionsClientContextKey string

var functionsClientKey functionsClientContextKey = "client"

func WithFunctionsClient(ctx context.Context, client FunctionsClient) context.Context {
	return context.WithValue(ctx, functionsClientKey, client)
}

func CallFunction(ctx context.Context, actionName string, body map[string]any) (any, error) {
	client, ok := ctx.Value(functionsClientKey).(FunctionsClient)
	if !ok {
		return nil, errors.New("no functions client in context")
	}

	return client.Request(ctx, actionName, body)
}
