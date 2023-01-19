package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
)

const (
	// JSON-RPC spec compliant error codes
	JsonRpcParseErrorCode     = -32700
	JsonRpcInvalidRequestCode = -32600
	JsonRpcMethodNotFoundCode = -32601
	JsonRpcInvalidParams      = -32602
	JsonRpcInternalErrorCode  = -32603

	// Application error codes
	HttpMethodNotAllowedCode = http.StatusMethodNotAllowed
)

func NewHandler(p *proto.Schema, api *proto.Api) common.ApiHandlerFunc {
	return func(r *http.Request) common.Response {

		if r.Method != http.MethodPost {
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				Error: JsonRpcError{
					Code:    HttpMethodNotAllowedCode,
					Message: "only HTTP post accepted",
				},
			})
		}

		req, err := parseJsonRpcRequest(r.Body)
		if err != nil {
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				Error: JsonRpcError{
					Code:    JsonRpcInvalidRequestCode,
					Message: fmt.Sprintf("error parsing JSON: %s", err.Error()),
				},
			})
		}

		if !req.Valid() {
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				ID:      &req.ID,
				Error: JsonRpcError{
					Code:    JsonRpcInvalidRequestCode,
					Message: "invalid JSON-RPC 2.0 request",
				},
			})
		}

		inputs := req.Params
		actionName := req.Method

		op := proto.FindOperation(p, actionName)
		if op == nil {
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				ID:      &req.ID,
				Error: JsonRpcError{
					Code:    JsonRpcMethodNotFoundCode,
					Message: "method not found",
				},
			})
		}

		scope := actions.NewScope(r.Context(), op, p)

		response, err := actions.Execute(scope, inputs)
		if err != nil {
			// TODO: map errors here properly e.g. record not found, unique constraints etc...
			return common.NewJsonResponse(http.StatusOK, JsonRpcErrorResponse{
				JsonRpc: "2.0",
				ID:      &req.ID,
				Error: JsonRpcError{
					Code:    JsonRpcInternalErrorCode,
					Message: "error executing request",
				},
			})
		}

		return common.NewJsonResponse(http.StatusOK, JsonRpcSuccessResponse{
			JsonRpc: "2.0",
			ID:      req.ID,
			Result:  response,
		})
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
