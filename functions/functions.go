package functions

import (
	"context"
	"errors"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

// Custom error codes returned from custom
// function runtime
// See packages/functions-runtime for original definition and more info.
const (
	UnknownError              = -32001
	DatabaseError             = -32002
	NoResultError             = -32003
	RecordNotFoundError       = -32004
	ForeignKeyConstraintError = -32005
	NotNullConstraintError    = -32006
	UniqueConstraintError     = -32007
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
	Meta   *FunctionsRuntimeMeta  `json:"meta"`
	Error  *FunctionsRuntimeError `json:"error"`
}

type FunctionsRuntimeMeta struct {
	Headers map[string][]string `json:"headers"`
}

// FunctionsRuntimeError follows the error object specification
// from the JSONRPC spec: https://www.jsonrpc.org/specification#error_object
type FunctionsRuntimeError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

type transportContextKey string

var contextKey transportContextKey = "transport"

func WithFunctionsTransport(ctx context.Context, transport Transport) context.Context {
	return context.WithValue(ctx, contextKey, transport)
}

func CallFunction(ctx context.Context, actionName string, body map[string]any) (any, map[string][]string, error) {
	transport, ok := ctx.Value(contextKey).(Transport)
	if !ok {
		return nil, nil, errors.New("no functions client in context")
	}

	requestHeaders, err := runtimectx.GetRequestHeaders(ctx)
	if err != nil {
		return nil, nil, err
	}

	joinedHeaders := map[string]string{}
	for k, v := range requestHeaders {
		joinedHeaders[k] = strings.Join(v, ", ")
	}

	var identity *runtimectx.Identity
	if runtimectx.IsAuthenticated(ctx) {
		identity, err = runtimectx.GetIdentity(ctx)
		if err != nil {
			return nil, nil, err
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
		return nil, nil, err
	}

	if resp.Error != nil {
		data := resp.Error.Data

		switch resp.Error.Code {
		case ForeignKeyConstraintError:
			return nil, common.NewForeignKeyConstraintError(data["column"].(string))
		case NoResultError:
			return nil, common.RuntimeError{
				Code:    common.ErrInternal,
				Message: "custom function returned no result",
			}
		case NotNullConstraintError:
			return nil, common.NewNotNullError(data["column"].(string))
		case UniqueConstraintError:
			return nil, common.NewUniquenessError(data["column"].(string))
		case RecordNotFoundError:
			return nil, common.NewNotFoundError()
		default:
			// includes generic errors thrown by custom functions during execution, plus other types of DatabaseError's that aren't fk/uniqueness/null constraint errors:
			// https://www.postgresql.org/docs/current/errcodes-appendix.html
			return nil, common.RuntimeError{
				Code:    common.ErrInternal,
				Message: resp.Error.Message,
			}
		}
	}

	return resp.Result, resp.Meta.Headers, nil
}
