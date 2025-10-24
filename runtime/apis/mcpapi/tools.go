package mcpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/schema/parser"
	"golang.org/x/crypto/bcrypt"
)

// listTools returns all actions as MCP tools, plus the auth tool
func listTools(ctx context.Context, api *proto.Api, schema *proto.Schema) *ListToolsResult {
	tools := []Tool{}

	// Add the authentication tool first
	tools = append(tools, Tool{
		Name:        "Auth.getToken",
		Description: "Authenticate and obtain an access token. Use this tool to get a Bearer token for authenticated requests. Returns an access token that can be used in the Authorization header for subsequent requests.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"grant_type": map[string]interface{}{
					"type":        "string",
					"description": "The OAuth grant type. Use 'password' for username/password authentication.",
					"default":     "password",
				},
				"username": map[string]interface{}{
					"type":        "string",
					"description": "The user's email address or username",
				},
				"password": map[string]interface{}{
					"type":        "string",
					"description": "The user's password",
				},
			},
			"required": []string{"username", "password"},
		},
	})

	// Add all actions as tools (both read and write)
	// Per MCP spec, tools are model-controlled operations - the AI decides when to call them
	for _, actionName := range proto.GetActionNamesForApi(schema, api) {
		protoAction := schema.FindAction(actionName)
		if protoAction == nil {
			continue
		}

		model := schema.FindModel(protoAction.ModelName)
		if model == nil {
			continue
		}

		tool := Tool{
			Name:        fmt.Sprintf("%s.%s", model.Name, protoAction.Name),
			Description: generateToolDescription(protoAction, model, schema),
			InputSchema: generateInputSchema(protoAction, model, schema),
		}

		tools = append(tools, tool)
	}

	return &ListToolsResult{
		Tools: tools,
	}
}

// callTool executes an action and returns the result
func callTool(ctx context.Context, params *CallToolParams, api *proto.Api, schema *proto.Schema) (*CallToolResult, error) {
	// Handle special Auth.getToken tool
	if params.Name == "Auth.getToken" {
		return callAuthTool(ctx, params, schema)
	}

	// Parse tool name to extract model and action
	modelName, actionName, err := parseToolName(params.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid tool name: %w", err)
	}

	// Find the action
	action := schema.FindAction(actionName)
	if action == nil {
		return nil, fmt.Errorf("action not found: %s", actionName)
	}

	// Verify model matches
	if action.ModelName != modelName {
		return nil, fmt.Errorf("action %s does not belong to model %s", actionName, modelName)
	}

	// Verify action is in this API
	if !isActionInAPI(action.Name, api, schema) {
		return nil, fmt.Errorf("action %s is not in API %s", actionName, api.Name)
	}

	// Execute the action with provided arguments
	scope := actions.NewScope(ctx, action, schema)
	inputs := params.Arguments
	if inputs == nil {
		inputs = make(map[string]interface{})
	}

	result, _, err := actions.Execute(scope, inputs)
	if err != nil {
		// Check if this is an auth/permission error
		// These should be returned as JSON-RPC errors (not tool errors)
		// so that WWW-Authenticate headers can be added
		var runtimeErr common.RuntimeError
		if errors.As(err, &runtimeErr) {
			if runtimeErr.Code == common.ErrAuthenticationFailed || runtimeErr.Code == common.ErrPermissionDenied {
				// Return as actual error to trigger JSON-RPC error with WWW-Authenticate
				return nil, err
			}
		}

		// For other errors, return as tool error result
		errorJSON, _ := json.Marshal(map[string]interface{}{
			"error":   err.Error(),
			"message": "Action execution failed",
		})

		return &CallToolResult{
			Content: []ToolContent{
				{
					Type:     "text",
					Text:     string(errorJSON),
					MimeType: "application/json",
				},
			},
			IsError: true,
		}, nil
	}

	// Serialize result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize result: %w", err)
	}

	return &CallToolResult{
		Content: []ToolContent{
			{
				Type:     "text",
				Text:     string(resultJSON),
				MimeType: "application/json",
			},
		},
		IsError: false,
	}, nil
}

// parseToolName extracts model and action names from a tool name
// Expected format: {model_name}.{action_name}
func parseToolName(toolName string) (modelName, actionName string, err error) {
	// Split on dot
	parts := []rune{}
	modelParts := []rune{}
	actionParts := []rune{}
	foundDot := false

	for _, ch := range toolName {
		if ch == '.' && !foundDot {
			foundDot = true
			modelParts = parts
			parts = []rune{}
		} else {
			parts = append(parts, ch)
		}
	}

	if !foundDot {
		return "", "", fmt.Errorf("invalid tool name format, expected {model}.{action}")
	}

	actionParts = parts
	modelName = string(modelParts)
	actionName = string(actionParts)

	if modelName == "" || actionName == "" {
		return "", "", fmt.Errorf("invalid tool name format, expected {model}.{action}")
	}

	return modelName, actionName, nil
}

// isActionInAPI checks if an action is exposed in the given API
func isActionInAPI(actionName string, api *proto.Api, schema *proto.Schema) bool {
	for _, name := range proto.GetActionNamesForApi(schema, api) {
		if name == actionName {
			return true
		}
	}
	return false
}

// callAuthTool handles the special Auth.getToken tool for authentication
func callAuthTool(ctx context.Context, params *CallToolParams, schema *proto.Schema) (*CallToolResult, error) {
	// Extract username and password from arguments
	username, ok := params.Arguments["username"].(string)
	if !ok || username == "" {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     `{"error": "username is required"}`,
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	password, ok := params.Arguments["password"].(string)
	if !ok || password == "" {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     `{"error": "password is required"}`,
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	// Look up identity by email
	identity, err := actions.FindIdentityByEmail(ctx, schema, username, oauth.KeelIssuer)
	if err != nil {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     `{"error": "database error"}`,
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	if identity == nil {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     `{"error": "invalid username or password"}`,
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	// Verify password
	passwordHash, ok := identity[parser.IdentityFieldNamePassword].(string)
	if !ok || passwordHash == "" {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     `{"error": "password authentication not configured for this identity"}`,
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     `{"error": "invalid username or password"}`,
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	// Generate access token
	identityID, ok := identity["id"].(string)
	if !ok || identityID == "" {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     `{"error": "identity ID not found"}`,
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	accessToken, expiresIn, err := oauth.GenerateAccessToken(ctx, identityID)
	if err != nil {
		return &CallToolResult{
			Content: []ToolContent{{
				Type:     "text",
				Text:     fmt.Sprintf(`{"error": "failed to generate token: %s"}`, err.Error()),
				MimeType: "application/json",
			}},
			IsError: true,
		}, nil
	}

	// Return token response
	response := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(expiresIn.Seconds()),
	}

	responseJSON, _ := json.Marshal(response)

	return &CallToolResult{
		Content: []ToolContent{{
			Type:     "text",
			Text:     string(responseJSON),
			MimeType: "application/json",
		}},
		IsError: false,
	}, nil
}
