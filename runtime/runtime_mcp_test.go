package runtime_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/storage"
	"github.com/teamkeel/keel/testhelpers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRuntimeMCP(t *testing.T) {
	for _, tCase := range mcpTestCases {
		t.Run(tCase.name, func(t *testing.T) {
			schema := protoSchema(t, tCase.keelSchema)

			dbConnInfo := &db.ConnectionInfo{
				Host:     "localhost",
				Port:     "8001",
				Username: "postgres",
				Database: "keel",
				Password: "postgres",
			}

			handler := runtime.NewApiHandler(schema)

			request := &http.Request{
				URL: &url.URL{
					Path: "/test/mcp",
				},
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(tCase.Body)),
				Header: tCase.Headers,
			}

			ctx := request.Context()

			ctx, err := testhelpers.WithTracing(ctx)
			require.NoError(t, err)

			pk, err := testhelpers.GetEmbeddedPrivateKey()
			require.NoError(t, err)

			ctx = runtimectx.WithPrivateKey(ctx, pk)

			// Add OAuth config for token generation
			ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{})

			dbName := testhelpers.DbNameForTestName(tCase.name)
			database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, dbName, true)
			require.NoError(t, err)
			defer database.Close()

			ctx = db.WithDatabase(ctx, database)

			storer, err := storage.NewDbStore(ctx, database)
			require.NoError(t, err)
			ctx = runtimectx.WithStorage(ctx, storer)

			request = request.WithContext(ctx)

			if tCase.databaseSetup != nil {
				tCase.databaseSetup(t, database.GetDB())
			}

			response := handler(request)
			body := string(response.Body)
			var res map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &res))

			if tCase.assertDatabase != nil {
				tCase.assertDatabase(t, database.GetDB(), res)
			}

			if tCase.assertError != nil {
				tCase.assertError(t, res)
			}

			if tCase.assertResponse != nil {
				tCase.assertResponse(t, res)
			}
		})
	}
}

type mcpTestCase struct {
	name           string
	keelSchema     string
	Body           string
	Headers        map[string][]string
	databaseSetup  func(t *testing.T, database *gorm.DB)
	assertDatabase func(t *testing.T, database *gorm.DB, response map[string]any)
	assertError    func(t *testing.T, data map[string]any)
	assertResponse func(t *testing.T, data map[string]any)
}

var mcpTestCases = []mcpTestCase{
	{
		name: "mcp_initialize",
		keelSchema: `
			model Thing {
				fields {
					name Text
				}
				actions {
					list listThings()
				}
				@permission(
					expression: true,
					actions: [list]
				)
			}
			api Test {
				models {
					Thing
				}
			}
		`,
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "initialize",
			"params": {
				"protocolVersion": "2024-11-05",
				"capabilities": {},
				"clientInfo": {
					"name": "test-client",
					"version": "1.0.0"
				}
			}
		}`,
		assertResponse: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			assert.NotNil(t, data["result"])

			result := data["result"].(map[string]any)
			assert.Equal(t, "2025-03-26", result["protocolVersion"])
			assert.NotNil(t, result["capabilities"])
			assert.NotNil(t, result["serverInfo"])

			serverInfo := result["serverInfo"].(map[string]any)
			assert.Equal(t, "keel", serverInfo["name"])

			// Verify auth instructions are included
			instructions := result["instructions"].(string)
			assert.Contains(t, instructions, "Authentication")
			assert.Contains(t, instructions, "/auth/token")
			assert.Contains(t, instructions, "Authorization: Bearer")
			assert.Contains(t, instructions, "grant_type")
		},
	},
	{
		name: "mcp_list_tools_includes_all_actions",
		keelSchema: `
			model Post {
				fields {
					title Text
					content Text
				}
				actions {
					get getPost(id)
					list listPosts()
					create createPost() with (title, content)
					update updatePost(id) with (title, content)
					delete deletePost(id)
				}
				@permission(
					expression: true,
					actions: [get, list, create, update, delete]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/list"
		}`,
		assertResponse: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			result := data["result"].(map[string]any)
			tools := result["tools"].([]interface{})

			// Should have 6 tools (Auth.getToken + all 5 Post actions)
			assert.Len(t, tools, 6)

			toolNames := []string{}
			for _, tl := range tools {
				tool := tl.(map[string]any)
				toolNames = append(toolNames, tool["name"].(string))
				assert.NotEmpty(t, tool["description"])
				assert.NotNil(t, tool["inputSchema"])

				// Verify input schema has proper structure
				inputSchema := tool["inputSchema"].(map[string]any)
				assert.Equal(t, "object", inputSchema["type"])
			}

			assert.Contains(t, toolNames, "Auth.getToken")
			assert.Contains(t, toolNames, "Post.getPost")
			assert.Contains(t, toolNames, "Post.listPosts")
			assert.Contains(t, toolNames, "Post.createPost")
			assert.Contains(t, toolNames, "Post.updatePost")
			assert.Contains(t, toolNames, "Post.deletePost")
		},
	},
	{
		name: "mcp_call_tool_executes_create_action",
		keelSchema: `
			model Post {
				fields {
					title Text
				}
				actions {
					create createPost() with (title)
				}
				@permission(
					expression: true,
					actions: [create]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/call",
			"params": {
				"name": "Post.createPost",
				"arguments": {
					"title": "New Post"
				}
			}
		}`,
		assertResponse: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			result := data["result"].(map[string]any)
			content := result["content"].([]interface{})

			require.Len(t, content, 1)
			contentItem := content[0].(map[string]any)
			assert.Equal(t, "text", contentItem["type"])
			assert.Equal(t, "application/json", contentItem["mimeType"])

			// Parse the result to verify post was created
			var post map[string]any
			require.NoError(t, json.Unmarshal([]byte(contentItem["text"].(string)), &post))
			assert.Equal(t, "New Post", post["title"])
			assert.NotEmpty(t, post["id"])
		},
		assertDatabase: func(t *testing.T, database *gorm.DB, response map[string]any) {
			var count int64
			database.Table("post").Count(&count)
			assert.Equal(t, int64(1), count)
		},
	},
	{
		name: "mcp_auth_failure_invalid_token",
		keelSchema: `
			model Post {
				fields {
					title Text
				}
				actions {
					list listPosts()
				}
				@permission(
					expression: ctx.isAuthenticated,
					actions: [list]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Headers: map[string][]string{
			"Authorization": {"Bearer invalid.token"},
		},
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/list"
		}`,
		assertError: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			assert.NotNil(t, data["error"])

			errorObj := data["error"].(map[string]any)
			assert.NotEqual(t, 0, errorObj["code"])
			assert.NotEmpty(t, errorObj["message"])
		},
	},
	{
		name: "mcp_invalid_method",
		keelSchema: `
			model Post {
				actions {
					list listPosts()
				}
				@permission(
					expression: true,
					actions: [list]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "invalid/method"
		}`,
		assertError: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			assert.NotNil(t, data["error"])

			errorObj := data["error"].(map[string]any)
			assert.Equal(t, float64(-32601), errorObj["code"]) // Method not found
		},
	},
	{
		name: "mcp_invalid_json",
		keelSchema: `
			model Post {
				actions {
					list listPosts()
				}
				@permission(
					expression: true,
					actions: [list]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Body: `{invalid json}`,
		assertError: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			assert.NotNil(t, data["error"])

			errorObj := data["error"].(map[string]any)
			assert.Equal(t, float64(-32700), errorObj["code"]) // Parse error
		},
	},
	{
		name: "mcp_list_tools_includes_auth_tool",
		keelSchema: `
			model Post {
				fields {
					title Text
				}
				actions {
					create createPost() with (title)
				}
				@permission(
					expression: true,
					actions: [create]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/list"
		}`,
		assertResponse: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			result := data["result"].(map[string]any)
			tools := result["tools"].([]interface{})

			// Should include Auth.getToken tool
			toolNames := []string{}
			for _, tl := range tools {
				tool := tl.(map[string]any)
				toolNames = append(toolNames, tool["name"].(string))
			}

			assert.Contains(t, toolNames, "Auth.getToken")

			// Find the auth tool and verify its schema
			var authTool map[string]any
			for _, tl := range tools {
				tool := tl.(map[string]any)
				if tool["name"].(string) == "Auth.getToken" {
					authTool = tool
					break
				}
			}

			assert.NotNil(t, authTool)
			assert.NotEmpty(t, authTool["description"])
			assert.NotNil(t, authTool["inputSchema"])

			inputSchema := authTool["inputSchema"].(map[string]any)
			properties := inputSchema["properties"].(map[string]any)
			assert.NotNil(t, properties["username"])
			assert.NotNil(t, properties["password"])
		},
	},
	{
		name: "mcp_call_auth_tool_success",
		keelSchema: `
			model Post {
				actions {
					create createPost()
				}
				@permission(
					expression: true,
					actions: [create]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		databaseSetup: func(t *testing.T, database *gorm.DB) {
			// Create an identity with password
			password := "testpassword"
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			require.NoError(t, err)

			identity := map[string]any{
				"id":       "test-identity-id",
				"email":    "test@example.com",
				"issuer":   "https://keel.so",
				"password": string(hashedPassword),
			}
			require.NoError(t, database.Table("identity").Create(identity).Error)
		},
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/call",
			"params": {
				"name": "Auth.getToken",
				"arguments": {
					"username": "test@example.com",
					"password": "testpassword"
				}
			}
		}`,
		assertResponse: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])

			// Check for error response
			if data["error"] != nil {
				t.Fatalf("MCP error: %v", data["error"])
			}

			result := data["result"].(map[string]any)

			// Check if there was an error first
			if result["isError"] != nil && result["isError"].(bool) {
				content := result["content"].([]interface{})
				if len(content) > 0 {
					t.Fatalf("Auth tool returned error: %v", content[0].(map[string]any)["text"])
				}
			}

			// Check isError field exists before asserting
			require.NotNil(t, result["isError"], "isError field missing from result")
			assert.False(t, result["isError"].(bool))

			content := result["content"].([]interface{})
			require.Len(t, content, 1)

			contentItem := content[0].(map[string]any)
			assert.Equal(t, "text", contentItem["type"])
			assert.Equal(t, "application/json", contentItem["mimeType"])

			// Parse the token response
			var tokenResponse map[string]any
			require.NoError(t, json.Unmarshal([]byte(contentItem["text"].(string)), &tokenResponse))

			assert.NotEmpty(t, tokenResponse["access_token"])
			assert.Equal(t, "Bearer", tokenResponse["token_type"])
			assert.NotZero(t, tokenResponse["expires_in"])
		},
	},
	{
		name: "mcp_call_auth_tool_invalid_credentials",
		keelSchema: `
			model Post {
				actions {
					create createPost()
				}
				@permission(
					expression: true,
					actions: [create]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/call",
			"params": {
				"name": "Auth.getToken",
				"arguments": {
					"username": "nonexistent@example.com",
					"password": "wrongpassword"
				}
			}
		}`,
		assertResponse: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])
			result := data["result"].(map[string]any)
			assert.True(t, result["isError"].(bool))

			content := result["content"].([]interface{})
			require.Len(t, content, 1)

			contentItem := content[0].(map[string]any)
			text := contentItem["text"].(string)
			assert.Contains(t, text, "error")
		},
	},
	{
		name: "mcp_call_tool_permission_denied",
		keelSchema: `
			model Post {
				fields {
					title Text
				}
				actions {
					create createPost() with (title)
				}
				@permission(
					expression: false,
					actions: [create]
				)
			}
			api Test {
				models {
					Post
				}
			}
		`,
		Body: `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/call",
			"params": {
				"name": "Post.createPost",
				"arguments": {
					"title": "New Post"
				}
			}
		}`,
		assertError: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "2.0", data["jsonrpc"])

			// Permission errors are now returned as JSON-RPC errors (not tool errors)
			// This allows WWW-Authenticate headers to be added
			errorObj := data["error"].(map[string]any)
			assert.NotNil(t, errorObj)
			assert.NotEmpty(t, errorObj["message"])

			// Should have keelCode indicating permission denied
			if dataMap, ok := errorObj["data"].(map[string]any); ok {
				assert.Equal(t, "ERR_PERMISSION_DENIED", dataMap["keelCode"])
			}
		},
	},
}
