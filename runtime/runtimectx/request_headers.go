package runtimectx

import (
	"context"
	"fmt"
)

const (
	requestHeadersContextKey contextKey = "requestHeaders"
)

func WithRequestHeaders(ctx context.Context, headers map[string][]string) context.Context {
	if headers != nil {
		ctx = context.WithValue(ctx, requestHeadersContextKey, headers)
	}

	return ctx
}

func GetRequestHeaders(ctx context.Context) (map[string][]string, error) {
	v, ok := ctx.Value(requestHeadersContextKey).(map[string][]string)
	if !ok {
		return nil, fmt.Errorf("context does not have key or is not map[string]string: %s", requestHeadersContextKey)
	}
	return v, nil
}
