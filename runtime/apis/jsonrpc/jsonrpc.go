package jsonrpc

import (
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

	// Application error codes
	HttpMethodNotAllowedCode = http.StatusMethodNotAllowed
)

func NewHandler(p *proto.Schema, api *proto.Api) common.ApiHandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "JsonRpc")
		defer span.End()

		if r.Method != http.MethodPost {
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				Error: JsonRpcError{
					Code:    HttpMethodNotAllowedCode,
					Message: "only HTTP post accepted",
				},
			}, nil)
		}

		identity, err := actions.HandleAuthorizationHeader(ctx, p, r.Header)
		if err != nil {
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				Error: JsonRpcError{
					Code:    RuntimeErrorCodeToJsonRpcErrorCode(common.ErrAuthenticationFailed),
					Message: "not authenticated",
				},
			}, nil)
		}
		if identity != nil {
			ctx = runtimectx.WithIdentity(ctx, identity)
		}

		req, err := parseJsonRpcRequest(r.Body)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				Error: JsonRpcError{
					Code:    JsonRpcInvalidRequestCode,
					Message: fmt.Sprintf("error parsing JSON: %s", err.Error()),
				},
			}, nil)
		}

		if !req.Valid() {
			span.SetStatus(codes.Error, "invalid JSON-RPC request")
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				ID:      &req.ID,
				Error: JsonRpcError{
					Code:    JsonRpcInvalidRequestCode,
					Message: "invalid JSON-RPC 2.0 request",
				},
			}, nil)
		}

		inputs := req.Params
		actionName := req.Method

		span.SetAttributes(
			attribute.String("request.id", req.ID),
			attribute.String("api.protocol", "RPC"),
		)

		op := proto.FindOperation(p, actionName)
		if op == nil {
			span.SetStatus(codes.Error, "action not found")
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				ID:      &req.ID,
				Error: JsonRpcError{
					Code:    JsonRpcMethodNotFoundCode,
					Message: "method not found",
				},
			}, nil)
		}

		scope := actions.NewScope(ctx, op, p)

		response, headers, err := actions.Execute(scope, inputs)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())

			code := JsonRpcInternalErrorCode
			message := "error executing request"

			var runtimeError common.RuntimeError
			if errors.As(err, &runtimeError) {
				span.SetAttributes(
					attribute.String("error.code", runtimeError.Code),
					attribute.String("error.message", runtimeError.Message),
				)
				code = RuntimeErrorCodeToJsonRpcErrorCode(runtimeError.Code)
				message = runtimeError.Message
			}

			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				ID:      &req.ID,
				Error: JsonRpcError{
					Code:    code,
					Message: message,
				},
			}, nil)
		}

		return common.NewJsonResponse(http.StatusOK, JsonRpcSuccessResponse{
			JsonRpc: "2.0",
			ID:      req.ID,
			Result:  response,
		}, headers)
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

func parseJsonRpcRequest(b io.ReadCloser) (req *JsonRpcRequest, err error) {
	body, err := io.ReadAll(b)
	if err != nil {
		return nil, err
	}

	req = &JsonRpcRequest{}
	err = json.Unmarshal(body, req)
	return req, err
}

func RuntimeErrorCodeToJsonRpcErrorCode(code string) int {
	switch code {
	case common.ErrAuthenticationFailed:
		return JsonRpcUnauthorized
	case common.ErrPermissionDenied:
		return JsonRpcForbidden
	case common.ErrInvalidInput, common.ErrRecordNotFound:
		return JsonRpcInvalidParams
	default:
		return JsonRpcInternalErrorCode
	}
}
