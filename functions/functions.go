package functions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

type Transport func(ctx context.Context, req *FunctionsRuntimeRequest) (*FunctionsRuntimeResponse, error)

type FunctionsRuntimeRequest struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
	Meta   map[string]any `json:"meta"`
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

	requestHeaders, err := runtimectx.GetRequestHeaders(ctx)
	if err != nil {
		return nil, err
	}

	joinedHeaders := map[string]string{}
	for k, v := range requestHeaders {
		joinedHeaders[k] = strings.Join(v, ", ")
	}

	var identity *runtimectx.Identity
	if runtimectx.IsAuthenticated(ctx) {
		identity, err = runtimectx.GetIdentity(ctx)
		if err != nil {
			return nil, err
		}
	}

	meta := map[string]any{
		"headers":  requestHeaders,
		"identity": identity,
	}

	req := &FunctionsRuntimeRequest{
		ID:     ksuid.New().String(),
		Method: actionName,
		Params: body,
		Meta:   meta,
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
