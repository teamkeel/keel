package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/jsonrpc")

const (
	// JSON-RPC spec compliant error codes
	JsonRpcParseErrorCode     = -32700
	JsonRpcInvalidRequestCode = -32600
	JsonRpcMethodNotFoundCode = -32601
	JsonRpcInvalidParams      = -32602
	JsonRpcInternalErrorCode  = -32603
	JsonRpcUnauthorized       = -32001 // Not part of the official spec
	JsonRpcForbidden          = -32003 // Not part of the official spec
)

func NewHandler(p *proto.Schema, api *proto.Api) common.ApiHandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "JsonRpc")
		defer span.End()

		if r.Method != http.MethodPost {
			err := common.NewHttpMethodNotAllowedError("only HTTP post is accepted")
			return NewErrorResponse(ctx, nil, err)
		}

		identity, err := actions.HandleAuthorizationHeader(ctx, p, r.Header)
		if err != nil {
			return NewErrorResponse(ctx, nil, err)
		}
		if identity != nil {
			ctx = runtimectx.WithIdentity(ctx, identity)
		}

		req, err := parseJsonRpcRequest(r.Body)
		if err != nil {
			err = common.NewInputMalformedError(fmt.Sprintf("error parsing JSON: %s", err.Error()))
			return NewErrorResponse(ctx, &req.ID, err)
		}

		if !req.Valid() {
			err = common.NewInputMalformedError("invalid JSON-RPC request")
			return NewErrorResponse(ctx, &req.ID, err)
		}

		inputs := req.Params
		actionName := req.Method

		span.SetAttributes(
			attribute.String("request.id", req.ID),
			attribute.String("api.protocol", "RPC"),
		)

		action := proto.FindAction(p, actionName)
		if action == nil {
			err = common.NewMethodNotFoundError()
			return NewErrorResponse(ctx, &req.ID, err)
		}

		scope := actions.NewScope(ctx, action, p)

		response, headers, err := actions.Execute(scope, inputs)
		if err != nil {
			return NewErrorResponse(ctx, &req.ID, err)
		}

		return NewSuccessResponse(ctx, req.ID, response, headers)
	}
}

type JsonRpcRequest struct {
	JsonRpc string         `json:"jsonrpc"`
	ID      string         `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

func (r JsonRpcRequest) Valid() bool {
	return r.Method != "" && r.ID != "" && r.JsonRpc == "2.0"
}

type JsonRpcSuccessResponse struct {
	JsonRpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  any    `json:"result"`
}

type JsonRpcErrorResponse struct {
	JsonRpc string       `json:"jsonrpc"`
	ID      *string      `json:"id"`
	Error   JsonRpcError `json:"error"`
}

type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

func NewSuccessResponse(ctx context.Context, requestId string, response any, headers map[string][]string) common.Response {
	return common.NewJsonResponse(http.StatusOK, JsonRpcSuccessResponse{
		JsonRpc: "2.0",
		ID:      requestId,
		Result:  response,
	}, headers)
}

func NewErrorResponse(ctx context.Context, requestId *string, err error) common.Response {
	span := trace.SpanFromContext(ctx)

	var response JsonRpcError
	var runtimeError common.RuntimeError

	switch {
	case errors.As(err, &runtimeError):
		response = JsonRpcError{
			Code:    runtimeErrorCodeToJsonRpcErrorCode(runtimeError.Code),
			Message: runtimeError.Message,
		}

		span.SetAttributes(
			attribute.String("error.code", runtimeError.Code),
			attribute.String("error.message", runtimeError.Message),
		)
	default:
		response = JsonRpcError{
			Code:    JsonRpcInternalErrorCode,
			Message: "error executing request",
		}
	}

	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())

	return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
		JsonRpc: "2.0",
		ID:      requestId,
		Error:   response,
	}, nil)
}

func parseJsonRpcRequest(b io.ReadCloser) (req *JsonRpcRequest, err error) {
	body, err := io.ReadAll(b)
	if err != nil {
		return nil, err
	}

	req = &JsonRpcRequest{}
	err = json.Unmarshal(body, req)
	return req, err
}

func runtimeErrorCodeToJsonRpcErrorCode(code string) int {
	switch code {
	case common.ErrAuthenticationFailed:
		return JsonRpcUnauthorized
	case common.ErrPermissionDenied:
		return JsonRpcForbidden
	case common.ErrInvalidInput, common.ErrRecordNotFound:
		return JsonRpcInvalidParams
	case common.ErrMethodNotFound:
		return JsonRpcMethodNotFoundCode
	case common.ErrInputMalformed:
		return JsonRpcInvalidRequestCode
	default:
		return JsonRpcInternalErrorCode
	}
}
