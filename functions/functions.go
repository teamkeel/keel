package functions

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// FunctionErrorCode represents custom error codes returned from custom function runtime
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
	BadRequestError           FunctionErrorCode = -32009
	InternalError             FunctionErrorCode = -32010
)

type FunctionType string

const (
	ActionFunction     FunctionType = "action"
	JobFunction        FunctionType = "job"
	SubscriberFunction FunctionType = "subscriber"
	FlowFunction       FunctionType = "flow"
	RouteFunction      FunctionType = "route"
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
	Status  int                 `json:"status"`
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

func CallPredefinedHook(ctx context.Context, hook config.FunctionHook) error {
	cfg, err := runtimectx.GetOAuthConfig(ctx)
	if err != nil {
		return err
	}

	if slices.Contains(cfg.EnabledHooks(), hook) {
		permissionState := common.NewPermissionState()
		permissionState.Grant()

		_, _, err = CallFunction(
			ctx,
			string(hook),
			nil,
			permissionState,
		)
	}

	return err
}

// CallFunction will invoke the custom function on the runtime node server.
func CallFunction(ctx context.Context, actionName string, body any, permissionState *common.PermissionState) (any, *FunctionsRuntimeMeta, error) {
	span := trace.SpanFromContext(ctx)

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

	var identity auth.Identity
	if auth.IsAuthenticated(ctx) {
		identity, err = auth.GetIdentity(ctx)
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

	return resp.Result, resp.Meta, nil
}

type RouteRequest struct {
	Body   string            `json:"body"`
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Params map[string]string `json:"params"`
	Query  string            `json:"query"`
}

type RouteResponse struct {
	Body       string            `json:"body"`
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}

func CallRoute(ctx context.Context, handler string, body *RouteRequest) (*RouteResponse, *FunctionsRuntimeMeta, error) {
	span := trace.SpanFromContext(ctx)

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

	var identity auth.Identity
	if auth.IsAuthenticated(ctx) {
		identity, err = auth.GetIdentity(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	secrets := runtimectx.GetSecrets(ctx)

	tracingContext := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, tracingContext)

	meta := map[string]any{
		"headers":  requestHeaders,
		"identity": identity,
		"secrets":  secrets,
		"tracing":  tracingContext,
	}

	req := &FunctionsRuntimeRequest{
		ID:     ksuid.New().String(),
		Method: handler,
		Type:   RouteFunction,
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

	b, err := json.Marshal(resp.Result)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	var routeResponse RouteResponse
	err = json.Unmarshal(b, &routeResponse)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	return &routeResponse, resp.Meta, nil
}

// CallJob will invoke the job function on the runtime node server.
func CallJob(ctx context.Context, job *proto.Job, inputs map[string]any, permissionState *common.PermissionState, trigger TriggerType) error {
	span := trace.SpanFromContext(ctx)

	transport, ok := ctx.Value(contextKey).(Transport)
	if !ok {
		return errors.New("no functions client in context")
	}

	var err error
	var identity auth.Identity
	if auth.IsAuthenticated(ctx) {
		identity, err = auth.GetIdentity(ctx)
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
		Method: strcase.ToLowerCamel(job.GetName()),
		Type:   JobFunction,
		Params: inputs,
		Meta:   meta,
	}

	span.SetAttributes(
		attribute.String("job.id", req.ID),
		attribute.String("job.name", job.GetName()),
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
	span := trace.SpanFromContext(ctx)

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
		Method: subscriber.GetName(),
		Type:   SubscriberFunction,
		Params: event,
		Meta:   meta,
	}

	span.SetAttributes(
		attribute.String("subscriber.id", req.ID),
		attribute.String("subscriber.name", subscriber.GetName()),
		attribute.String("event.name", event.EventName),
		attribute.String("event.target_id", event.Target.Id),
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

// CallFlow will invoke the flow function on the runtime node server.
func CallFlow(ctx context.Context, flow *proto.Flow, runId string, inputs map[string]any, data map[string]any, action string) (any, *FunctionsRuntimeMeta, error) {
	span := trace.SpanFromContext(ctx)

	transport, ok := ctx.Value(contextKey).(Transport)
	if !ok {
		return nil, nil, errors.New("no functions client in context")
	}

	secrets := runtimectx.GetSecrets(ctx)

	tracingContext := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, tracingContext)

	var identity auth.Identity
	var err error
	if auth.IsAuthenticated(ctx) {
		identity, err = auth.GetIdentity(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	meta := map[string]any{
		"runId":    runId,
		"secrets":  secrets,
		"tracing":  tracingContext,
		"inputs":   inputs,
		"data":     data,
		"identity": identity,
	}

	if action != "" {
		meta["action"] = action
		span.SetAttributes(attribute.String("action", action))
	}

	req := &FunctionsRuntimeRequest{
		ID:     ksuid.New().String(),
		Method: strcase.ToLowerCamel(flow.GetName()),
		Type:   FlowFunction,
		Meta:   meta,
	}

	span.SetAttributes(
		attribute.String("request.id", req.ID),
		attribute.String("run.id", runId),
		attribute.String("flow.name", flow.GetName()),
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

	return resp.Result, resp.Meta, nil
}

// Parse the error from the functions runtime to the appropriate go runtime error.
func toRuntimeError(errorResponse *FunctionsRuntimeError) error {
	data := errorResponse.Data

	switch errorResponse.Code {
	case InternalError:
		return common.RuntimeError{
			Code:    common.ErrInternal,
			Message: "error executing request",
		}
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
		return common.NewNotFoundError(errorResponse.Message)
	case BadRequestError:
		return common.NewValidationError(errorResponse.Message)
	case DatabaseError:
		// other types of DatabaseError's that aren't fk/uniqueness/null constraint errors:
		// https://www.postgresql.org/docs/current/errcodes-appendix.html
		return common.RuntimeError{
			Code:    common.ErrInternal,
			Message: errorResponse.Message,
		}
	default:
		// All other errors that originate from user code
		return common.RuntimeError{
			Code:    common.ErrUnknown,
			Message: errorResponse.Message,
		}
	}
}
