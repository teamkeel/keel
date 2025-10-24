package mcpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/mcpapi")

// NewHandler creates a new MCP protocol handler for the given API
func NewHandler(schema *proto.Schema, api *proto.Api) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "MCP")
		defer span.End()

		// Build base URL from request (needed for error responses)
		baseURL := buildBaseURL(r)

		// Only POST method is supported for MCP JSON-RPC requests
		if r.Method != http.MethodPost {
			log.WithField("method", r.Method).Warn("MCP request with invalid HTTP method")
			err := common.NewHttpMethodNotAllowedError("only HTTP POST is accepted for MCP requests")
			return newErrorResponse(ctx, nil, err, baseURL)
		}

		// Handle authentication
		identity, err := actions.HandleAuthorizationHeader(ctx, schema, r.Header)
		if err != nil {
			log.WithError(err).Warn("MCP authentication failed")
			return newErrorResponse(ctx, nil, err, baseURL)
		}
		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
			identityEmail := ""
			if email, ok := identity["email"].(string); ok {
				identityEmail = email
			}
			log.WithField("identity_email", identityEmail).Info("MCP request authenticated")
		} else {
			log.Info("MCP request without authentication")
		}

		// Parse MCP request
		req, err := parseRequest(r.Body)
		if err != nil {
			log.WithError(err).Error("Failed to parse MCP request")
			// When parsing fails, we don't have a request ID
			return newErrorResponse(ctx, nil, &ErrorObj{
				Code:    ErrorParseError,
				Message: fmt.Sprintf("error parsing JSON: %s", err.Error()),
			}, baseURL)
		}

		span.SetAttributes(
			attribute.String("request.method", req.Method),
			attribute.String("api.protocol", "MCP"),
		)

		log.WithFields(log.Fields{
			"method":        req.Method,
			"id":            req.ID,
			"authenticated": identity != nil,
			"api":           api.Name,
		}).Info("MCP request received")

		// Route to appropriate handler based on method
		var result interface{}
		var mcpErr *ErrorObj

		switch req.Method {
		case MethodInitialize:
			result, mcpErr = handleInitialize(ctx, req.Params, baseURL)
			log.WithField("protocol_version", MCPVersion).Info("MCP initialize complete")
		case MethodListTools:
			result, mcpErr = handleListTools(ctx, api, schema)
			if mcpErr == nil && result != nil {
				if listResult, ok := result.(*ListToolsResult); ok {
					log.WithFields(log.Fields{
						"tool_count": len(listResult.Tools),
						"api":        api.Name,
					}).Info("MCP tools/list complete")
				}
			}
		case MethodCallTool:
			result, mcpErr = handleCallTool(ctx, req.Params, api, schema)
			if mcpErr == nil {
				log.WithField("method", req.Method).Info("MCP tool call complete")
			}
		default:
			log.WithField("method", req.Method).Warn("Unknown MCP method")
			mcpErr = &ErrorObj{
				Code:    ErrorMethodNotFound,
				Message: fmt.Sprintf("unknown method: %s", req.Method),
			}
		}

		if mcpErr != nil {
			log.WithFields(log.Fields{
				"error_code":    mcpErr.Code,
				"error_message": mcpErr.Message,
				"method":        req.Method,
			}).Error("MCP request failed")
			return newErrorResponse(ctx, req.ID, mcpErr, baseURL)
		}

		return newSuccessResponse(ctx, req.ID, result)
	}
}

// handleInitialize handles the initialize method
func handleInitialize(ctx context.Context, params interface{}, baseURL string) (interface{}, *ErrorObj) {
	// Parse initialize params
	var initParams InitializeParams
	if params != nil {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return nil, &ErrorObj{
				Code:    ErrorInvalidParams,
				Message: "failed to parse initialize parameters",
			}
		}
		if err := json.Unmarshal(paramsBytes, &initParams); err != nil {
			return nil, &ErrorObj{
				Code:    ErrorInvalidParams,
				Message: "invalid initialize parameters",
			}
		}
	}

	// Build authentication instructions
	instructions := buildAuthenticationInstructions(baseURL)

	// Return server capabilities
	// Only advertise Tools - all Keel actions (read and write) are exposed as tools
	// per MCP spec where tools are model-controlled operations
	return &InitializeResult{
		ProtocolVersion: MCPVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "keel",
			Version: "1.0.0",
		},
		Instructions: instructions,
	}, nil
}

// buildBaseURL constructs the base URL from the HTTP request
func buildBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	if host == "" {
		host = "localhost:8000"
	}

	// Remove the /mcp suffix and API path to get base URL
	return fmt.Sprintf("%s://%s", scheme, host)
}

// buildAuthenticationInstructions creates detailed auth instructions for MCP clients
func buildAuthenticationInstructions(baseURL string) string {
	return fmt.Sprintf(`# Keel MCP Server

## Overview
This is a Keel backend server exposed via MCP. All Keel actions (read and write operations) are exposed as Tools that the AI model can autonomously discover and call.

## Authentication
All operations respect the same authentication and authorization rules as other Keel APIs.

OAuth 2.1 metadata is available at:
- Authorization Server: %s/.well-known/oauth-authorization-server
- Protected Resource: %s/.well-known/oauth-protected-resource

### Quick Start: Use the Auth.getToken Tool

**The easiest way to authenticate is to use the built-in Auth.getToken tool!**

This tool is always available and doesn't require authentication to call.

**Tool Name:** Auth.getToken

**Arguments:**
- username (required): Your email address
- password (required): Your password
- grant_type (optional): Authentication method (default: "password")

**Example:**
Call the Auth.getToken tool with your credentials, and it will return an access token that you can use for all subsequent requests.

**Response:**
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400
}

After getting the token, include it in the Authorization header for authenticated requests.

### Alternative: Direct HTTP Token Endpoint

You can also get a token directly via HTTP:

**Token Endpoint:** %s/auth/token

**Method:** POST
**Content-Type:** application/json

**Password Grant Example:**
{
  "grant_type": "password",
  "username": "your-email@example.com",
  "password": "your-password"
}

**Response:**
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400
}

### Using the Token

Include the access token in the Authorization header for all MCP requests:

**Header:** Authorization: Bearer <access_token>

### Available Auth Endpoints

- **%s/auth/token** - Obtain access tokens (password, authorization_code, refresh_token grants)
- **%s/auth/providers** - List available OAuth providers (Google, GitHub, etc.)
- **%s/auth/authorize** - Initiate OAuth flow
- **%s/auth/callback** - OAuth callback endpoint
- **%s/auth/revoke** - Revoke tokens

### Example: Getting a Token via CLI

curl -X POST %s/auth/token \
  -H "Content-Type: application/json" \
  -d '{"grant_type":"password","username":"user@example.com","password":"pass"}'

### MCP Client Configuration

After obtaining a token, configure your MCP client with:

{
  "command": "your-mcp-client",
  "args": ["--url", "%s"],
  "env": {
    "AUTHORIZATION": "Bearer <your-access-token>"
  }
}

Or use the Authorization header in your fetch configuration.

## Permissions

Actions may have permission rules that restrict access based on:
- Authentication status (ctx.isAuthenticated)
- Identity attributes (ctx.identity.email, ctx.identity.role, etc.)
- Data ownership rules

When permissions fail, tools will return an error response with isError: true.`,
		baseURL, baseURL, baseURL, baseURL, baseURL, baseURL, baseURL, baseURL, baseURL, baseURL)
}

// handleListResources handles the resources/list method
// handleListTools handles the tools/list method
func handleListTools(ctx context.Context, api *proto.Api, schema *proto.Schema) (interface{}, *ErrorObj) {
	result := listTools(ctx, api, schema)
	return result, nil
}

// handleCallTool handles the tools/call method
func handleCallTool(ctx context.Context, params interface{}, api *proto.Api, schema *proto.Schema) (interface{}, *ErrorObj) {
	// Parse call tool params
	var callParams CallToolParams
	if params == nil {
		return nil, &ErrorObj{
			Code:    ErrorInvalidParams,
			Message: "missing tool call parameters",
		}
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, &ErrorObj{
			Code:    ErrorInvalidParams,
			Message: "failed to parse call tool parameters",
		}
	}
	if err := json.Unmarshal(paramsBytes, &callParams); err != nil {
		return nil, &ErrorObj{
			Code:    ErrorInvalidParams,
			Message: "invalid call tool parameters",
		}
	}

	// Execute the tool
	result, err := callTool(ctx, &callParams, api, schema)
	if err != nil {
		return nil, mapErrorToMCPError(err)
	}

	return result, nil
}

// parseRequest parses an MCP JSON-RPC request
func parseRequest(body io.ReadCloser) (*Request, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var req Request
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return nil, err
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		return nil, fmt.Errorf("invalid JSON-RPC version: %s", req.JSONRPC)
	}

	return &req, nil
}

// newSuccessResponse creates a successful MCP response
func newSuccessResponse(ctx context.Context, requestID interface{}, result interface{}) common.Response {
	response := Response{
		JSONRPC: "2.0",
		ID:      requestID,
		Result:  result,
	}

	return common.NewJsonResponse(http.StatusOK, response, nil)
}

// newErrorResponse creates an error MCP response
func newErrorResponse(ctx context.Context, requestID interface{}, err interface{}, baseURL string) common.Response {
	span := trace.SpanFromContext(ctx)

	var errorObj *ErrorObj

	switch e := err.(type) {
	case *ErrorObj:
		errorObj = e
	case error:
		errorObj = mapErrorToMCPError(e)
	default:
		errorObj = &ErrorObj{
			Code:    ErrorInternal,
			Message: "internal server error",
		}
	}

	span.SetAttributes(
		attribute.Int("error.code", errorObj.Code),
		attribute.String("error.message", errorObj.Message),
	)

	if e, ok := err.(error); ok {
		span.RecordError(e, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, e.Error())
	}

	response := Response{
		JSONRPC: "2.0",
		ID:      requestID,
		Error:   errorObj,
	}

	var metadata *common.ResponseMetadata

	// Add WWW-Authenticate header for authentication/authorization errors (MCP spec requirement)
	// Check both error types and ErrorObj with keelCode in Data
	shouldAddAuthHeader := false

	if e, ok := err.(error); ok {
		var runtimeErr common.RuntimeError
		if errors.As(e, &runtimeErr) {
			if runtimeErr.Code == common.ErrAuthenticationFailed || runtimeErr.Code == common.ErrPermissionDenied {
				shouldAddAuthHeader = true
			}
		}
	} else if errorObj != nil && errorObj.Data != nil {
		// Check if this is an ErrorObj with keelCode indicating auth/permission error
		if dataMap, ok := errorObj.Data.(map[string]interface{}); ok {
			if keelCode, ok := dataMap["keelCode"].(string); ok {
				if keelCode == common.ErrAuthenticationFailed || keelCode == common.ErrPermissionDenied {
					shouldAddAuthHeader = true
				}
			}
		}
	}

	if shouldAddAuthHeader {
		// Per MCP spec, include resource_metadata URL and scope in WWW-Authenticate header
		authHeader := fmt.Sprintf(
			`Bearer resource_metadata="%s/.well-known/oauth-protected-resource", scope="tools:execute"`,
			baseURL,
		)
		headers := http.Header{}
		headers.Set("WWW-Authenticate", authHeader)
		metadata = &common.ResponseMetadata{
			Headers: headers,
		}
	}

	return common.NewJsonResponse(http.StatusOK, response, metadata)
}

// mapErrorToMCPError converts Go errors to MCP error objects
func mapErrorToMCPError(err error) *ErrorObj {
	if err == nil {
		return nil
	}

	// Check if it's a Keel runtime error
	var runtimeErr common.RuntimeError
	if !errors.As(err, &runtimeErr) {
		// Generic error
		return &ErrorObj{
			Code:    ErrorInternal,
			Message: err.Error(),
		}
	}

	// Map Keel error codes to MCP error codes
	code := ErrorInternal
	switch runtimeErr.Code {
	case common.ErrInvalidInput:
		code = ErrorInvalidParams
	case common.ErrInputMalformed:
		code = ErrorParseError
	case common.ErrMethodNotFound:
		code = ErrorMethodNotFound
	case common.ErrAuthenticationFailed:
		code = ErrorInternal // No specific auth code in MCP
	case common.ErrPermissionDenied:
		code = ErrorInternal // No specific permission code in MCP
	}

	return &ErrorObj{
		Code:    code,
		Message: runtimeErr.Message,
		Data: map[string]interface{}{
			"keelCode": runtimeErr.Code,
		},
	}
}
