package functions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/functions")

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
	PermissionError           = -32008
)

type Transport func(ctx context.Context, req *FunctionsRuntimeRequest) (*FunctionsRuntimeResponse, error)

type FunctionsRuntimeRequest struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Params any            `json:"params"`
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
	Code    int    `json:"code"`
	Message string `json:"message"`

	// Data represents any additional error metadata that the functions-runtime want to send in addition to the error code + message
	Data map[string]any `json:"data"`
}

type transportContextKey string

var contextKey transportContextKey = "transport"

func WithFunctionsTransport(ctx context.Context, transport Transport) context.Context {
	return context.WithValue(ctx, contextKey, transport)
}

func CallFunction(ctx context.Context, actionName string, body any, permissionState *common.PermissionState) (any, map[string][]string, error) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("Call Function: %s", actionName))
	defer span.End()

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

	secrets := runtimectx.GetSecrets(ctx)

	tracingContext := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, tracingContext)

	meta := map[string]any{
		"headers":         requestHeaders,
		"identity":        identity,
		"secrets":         secrets,
		"tracing":         tracingContext,
		"permissionState": permissionState,
	}

	req := &FunctionsRuntimeRequest{
		ID:     ksuid.New().String(),
		Method: actionName,
		Params: body,
		Meta:   meta,
	}

	span.SetAttributes(
		attribute.String("jsonrpc.id", req.ID),
		attribute.String("jsonrpc.method", req.Method),
	)

	resp, err := transport(ctx, req)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	if resp.Error != nil {
		span.SetStatus(codes.Error, resp.Error.Message)
		span.SetAttributes(attribute.Int("error.code", resp.Error.Code))

		data := resp.Error.Data

		switch resp.Error.Code {
		case PermissionError:
			return nil, nil, common.NewPermissionError()
		case ForeignKeyConstraintError:
			return nil, nil, common.NewForeignKeyConstraintError(data["column"].(string))
		case NoResultError:
			return nil, nil, common.RuntimeError{
				Code:    common.ErrInternal,
				Message: "custom function returned no result",
			}
		case NotNullConstraintError:
			return nil, nil, common.NewNotNullError(data["column"].(string))
		case UniqueConstraintError:
			return nil, nil, common.NewUniquenessError(strings.Split(data["column"].(string), ", "))
		case RecordNotFoundError:
			return nil, nil, common.NewNotFoundError()
		default:
			// includes generic errors thrown by custom functions during execution, plus other types of DatabaseError's that aren't fk/uniqueness/null constraint errors:
			// https://www.postgresql.org/docs/current/errcodes-appendix.html
			return nil, nil, common.RuntimeError{
				Code:    common.ErrInternal,
				Message: resp.Error.Message,
			}
		}
	}

	return resp.Result, resp.Meta.Headers, nil
}