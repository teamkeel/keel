package functions

import (
	"context"
	"errors"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/proto"
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
type FunctionErrorCode int

const (
	UnknownError              FunctionErrorCode = -32001
	DatabaseError             FunctionErrorCode = -32002
	NoResultError             FunctionErrorCode = -32003
	RecordNotFoundError       FunctionErrorCode = -32004
	ForeignKeyConstraintError FunctionErrorCode = -32005
	NotNullConstraintError    FunctionErrorCode = -32006
	UniqueConstraintError     FunctionErrorCode = -32007
	PermissionError           FunctionErrorCode = -32008
)

type FunctionType string

const (
	ActionFunction     FunctionType = "action"
	JobFunction        FunctionType = "job"
	SubscriberFunction FunctionType = "subscriber"
)

type TriggerType string

const (
	ManualTrigger    TriggerType = "manual"
	ScheduledTrigger TriggerType = "scheduled"
)

type Transport func(ctx context.Context, req *FunctionsRuntimeRequest) (*FunctionsRuntimeResponse, error)

type FunctionsRuntimeRequest struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Type   FunctionType   `json:"type"`
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
	Code    FunctionErrorCode `json:"code"`
	Message string            `json:"message"`

	// Data represents any additional error metadata that the functions-runtime want to send in addition to the error code + message
	Data map[string]any `json:"data"`
}

type transportContextKey string

var contextKey transportContextKey = "transport"

func WithFunctionsTransport(ctx context.Context, transport Transport) context.Context {
	return context.WithValue(ctx, contextKey, transport)
}

// CallFunction will invoke the custom function on the runtime node server.
func CallFunction(ctx context.Context, actionName string, body any, permissionState *common.PermissionState) (any, map[string][]string, error) {
	ctx, span := tracer.Start(ctx, "Call function")
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
		Type:   ActionFunction,
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
		span.SetAttributes(attribute.Int("error.code", int(resp.Error.Code)))
		return nil, nil, toRuntimeError(resp.Error)
	}

	return resp.Result, resp.Meta.Headers, nil
}

// CallJob will invoke the job function on the runtime node server.
func CallJob(ctx context.Context, job *proto.Job, inputs map[string]any, permissionState *common.PermissionState, trigger TriggerType) error {
	ctx, span := tracer.Start(ctx, "Call job")
	defer span.End()

	transport, ok := ctx.Value(contextKey).(Transport)
	if !ok {
		return errors.New("no functions client in context")
	}

	var err error
	var identity *runtimectx.Identity
	if runtimectx.IsAuthenticated(ctx) {
		identity, err = runtimectx.GetIdentity(ctx)
		if err != nil {
			return err
		}
	}

	secrets := runtimectx.GetSecrets(ctx)

	tracingContext := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, tracingContext)

	meta := map[string]any{
		"identity":        identity,
		"secrets":         secrets,
		"tracing":         tracingContext,
		"permissionState": permissionState,
		"triggerType":     trigger,
	}

	req := &FunctionsRuntimeRequest{
		ID:     ksuid.New().String(),
		Method: strcase.ToLowerCamel(job.Name),
		Type:   JobFunction,
		Params: inputs,
		Meta:   meta,
	}

	span.SetAttributes(
		attribute.String("job.id", req.ID),
		attribute.String("job.name", job.Name),
	)

	resp, err := transport(ctx, req)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if resp.Error != nil {
		span.SetStatus(codes.Error, resp.Error.Message)
		span.SetAttributes(attribute.Int("error.code", int(resp.Error.Code)))
		return toRuntimeError(resp.Error)
	}

	return nil
}

// CallSubscriber will invoke the subscriber function on the runtime node server.
func CallSubscriber(ctx context.Context, subscriber *proto.Subscriber, event *events.Event) error {
	ctx, span := tracer.Start(ctx, "Call subscriber")
	defer span.End()

	transport, ok := ctx.Value(contextKey).(Transport)
	if !ok {
		return errors.New("no functions client in context")
	}

	secrets := runtimectx.GetSecrets(ctx)

	tracingContext := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, tracingContext)

	meta := map[string]any{
		"secrets": secrets,
		"tracing": tracingContext,
	}

	req := &FunctionsRuntimeRequest{
		ID:     ksuid.New().String(),
		Method: subscriber.Name,
		Type:   SubscriberFunction,
		Params: event,
		Meta:   meta,
	}

	span.SetAttributes(
		attribute.String("subscriber.id", req.ID),
		attribute.String("subscriber.name", subscriber.Name),
	)

	resp, err := transport(ctx, req)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if resp.Error != nil {
		span.SetStatus(codes.Error, resp.Error.Message)
		span.SetAttributes(attribute.Int("error.code", int(resp.Error.Code)))
		return toRuntimeError(resp.Error)
	}

	return nil
}

// Parse the error from the functions runtime to the appropriate go runtime error.
func toRuntimeError(errorResponse *FunctionsRuntimeError) error {
	data := errorResponse.Data

	switch errorResponse.Code {
	case PermissionError:
		return common.NewPermissionError()
	case ForeignKeyConstraintError:
		return common.NewForeignKeyConstraintError(data["column"].(string))
	case NoResultError:
		return common.RuntimeError{
			Code:    common.ErrInternal,
			Message: "custom function returned no result",
		}
	case NotNullConstraintError:
		return common.NewNotNullError(data["column"].(string))
	case UniqueConstraintError:
		return common.NewUniquenessError(strings.Split(data["column"].(string), ", "))
	case RecordNotFoundError:
		return common.NewNotFoundError()
	default:
		// includes generic errors thrown by custom functions during execution, plus other types of DatabaseError's that aren't fk/uniqueness/null constraint errors:
		// https://www.postgresql.org/docs/current/errcodes-appendix.html
		return common.RuntimeError{
			Code:    common.ErrInternal,
			Message: errorResponse.Message,
		}
	}
}
