package functions

import (
	"context"
	"errors"
	"fmt"

	"github.com/segmentio/ksuid"
)

type Transport func(ctx context.Context, req *FunctionsRuntimeRequest) (*FunctionsRuntimeResponse, error)

type FunctionsRuntimeRequest struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

type FunctionsRuntimeResponse struct {
	ID     string                 `json:"id"`
	Result any                    `json:"result"`
	Error  *FunctionsRuntimeError `json:"error"`
}

type FunctionsRuntimeError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type transportContextKey string

var contextKey transportContextKey = "transport"

func WithFunctionsTransport(ctx context.Context, transport Transport) context.Context {
	return context.WithValue(ctx, contextKey, transport)
}

func CallFunction(ctx context.Context, actionName string, body map[string]any) (any, error) {
	transport, ok := ctx.Value(contextKey).(Transport)
	if !ok {
		return nil, errors.New("no functions client in context")
	}

	req := &FunctionsRuntimeRequest{
		ID:     ksuid.New().String(),
		Method: actionName,
		Params: body,
	}

	resp, err := transport(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("%s function returned error (%d) - %s", actionName, resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}
