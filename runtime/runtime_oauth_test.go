package runtime_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/storage"
	"github.com/teamkeel/keel/testhelpers"
	"golang.org/x/crypto/bcrypt"
)

// responseRecorder is a simple http.ResponseWriter for tests
type responseRecorder struct {
	status  int
	headers http.Header
	body    *bytes.Buffer
}

func (r *responseRecorder) Header() http.Header {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	return r.headers
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.body == nil {
		r.body = &bytes.Buffer{}
	}
	return r.body.Write(data)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

// TestOAuthPKCEFlow tests the full OAuth 2.1 authorization code flow with PKCE
// This tests the MCP-compliant OAuth implementation end-to-end
func TestOAuthPKCEFlow(t *testing.T) {
	schema := protoSchema(t, `
		model Post {
			fields {
				title Text
			}
			actions {
				create createPost() with (title)
			}
			@permission(
				expression: ctx.isAuthenticated,
				actions: [create]
			)
		}
		api Test {
			models {
				Post
			}
		}
	`)

	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Database: "keel",
		Password: "postgres",
	}

	ctx := context.Background()
	ctx, err := testhelpers.WithTracing(ctx)
	require.NoError(t, err)

	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)

	ctx = runtimectx.WithPrivateKey(ctx, pk)
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{})

	dbName := testhelpers.DbNameForTestName("oauth_pkce_flow")
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, dbName, true)
	require.NoError(t, err)
	defer database.Close()

	ctx = db.WithDatabase(ctx, database)

	storer, err := storage.NewDbStore(ctx, database)
	require.NoError(t, err)
	ctx = runtimectx.WithStorage(ctx, storer)

	// Create test identity
	password := "testpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	identity := map[string]any{
		"id":       "test-identity-id",
		"email":    "test@example.com",
		"issuer":   "https://keel.so",
		"password": string(hashedPassword),
	}
	require.NoError(t, database.GetDB().Table("identity").Create(identity).Error)

	authHandler := runtime.NewAuthHandler(schema)

	// Step 1: Generate PKCE parameters
	codeVerifier := uniuri.NewLen(64)
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	// Step 2: Authorization request with PKCE
	authParams := url.Values{}
	authParams.Add("response_type", "code")
	authParams.Add("redirect_uri", "http://localhost:3000/callback")
	authParams.Add("code_challenge", codeChallenge)
	authParams.Add("code_challenge_method", "S256")
	authParams.Add("username", "test@example.com")
	authParams.Add("password", password)
	authParams.Add("state", "test-state-123")

	authReq := &http.Request{
		URL: &url.URL{
			Path:     "/auth/authorize/keel",
			RawQuery: authParams.Encode(),
		},
		Method: http.MethodGet,
		Header: http.Header{},
	}
	authReq = authReq.WithContext(ctx)

	authResp := authHandler(nil, authReq)

	// Should redirect with auth code
	require.Equal(t, http.StatusFound, authResp.Status)

	// Extract auth code from redirect location
	location := http.Header(authResp.Headers).Get("Location")
	require.NotEmpty(t, location)

	redirectURL, err := url.Parse(location)
	require.NoError(t, err)

	authCode := redirectURL.Query().Get("code")
	require.NotEmpty(t, authCode, "authorization code not found in redirect")

	state := redirectURL.Query().Get("state")
	assert.Equal(t, "test-state-123", state)

	// Step 3: Exchange auth code for access token with code_verifier
	tokenData := url.Values{}
	tokenData.Add("grant_type", "authorization_code")
	tokenData.Add("code", authCode)
	tokenData.Add("code_verifier", codeVerifier)

	tokenReq := &http.Request{
		URL: &url.URL{
			Path: "/auth/token",
		},
		Method: http.MethodPost,
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		Body: io.NopCloser(strings.NewReader(tokenData.Encode())),
	}
	tokenReq = tokenReq.WithContext(ctx)

	tokenResp := authHandler(nil, tokenReq)
	require.Equal(t, http.StatusOK, tokenResp.Status)

	var tokenResponse map[string]any
	require.NoError(t, json.Unmarshal(tokenResp.Body, &tokenResponse))

	accessToken, ok := tokenResponse["access_token"].(string)
	require.True(t, ok, "access_token not found in response")
	require.NotEmpty(t, accessToken)

	assert.Equal(t, "bearer", tokenResponse["token_type"])
	assert.NotZero(t, tokenResponse["expires_in"])

	// Step 4: Use access token to call MCP endpoint
	apiHandler := runtime.NewApiHandler(schema)

	mcpReq := &http.Request{
		URL: &url.URL{
			Path: "/test/mcp",
		},
		Method: http.MethodPost,
		Header: http.Header{
			"Authorization": {"Bearer " + accessToken},
		},
		Body: io.NopCloser(strings.NewReader(`{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/call",
			"params": {
				"name": "Post.createPost",
				"arguments": {
					"title": "OAuth Test Post"
				}
			}
		}`)),
	}
	mcpReq = mcpReq.WithContext(ctx)

	mcpResp := apiHandler(mcpReq)
	require.Equal(t, http.StatusOK, mcpResp.Status)

	var mcpResponse map[string]any
	require.NoError(t, json.Unmarshal(mcpResp.Body, &mcpResponse))

	// Should be successful (not an error)
	result, ok := mcpResponse["result"].(map[string]any)
	require.True(t, ok, "result not found in MCP response")
	require.False(t, result["isError"].(bool), "MCP call should succeed with valid token")

	// Verify post was created
	var count int64
	database.GetDB().Table("post").Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestOAuthPKCEValidation tests PKCE validation scenarios
func TestOAuthPKCEValidation(t *testing.T) {
	schema := protoSchema(t, `
		model Post {
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
	`)

	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Database: "keel",
		Password: "postgres",
	}

	ctx := context.Background()
	ctx, err := testhelpers.WithTracing(ctx)
	require.NoError(t, err)

	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)

	ctx = runtimectx.WithPrivateKey(ctx, pk)
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{})

	dbName := testhelpers.DbNameForTestName("oauth_pkce_validation")
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, dbName, true)
	require.NoError(t, err)
	defer database.Close()

	ctx = db.WithDatabase(ctx, database)

	storer, err := storage.NewDbStore(ctx, database)
	require.NoError(t, err)
	ctx = runtimectx.WithStorage(ctx, storer)

	// Create test identity
	password := "testpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	identity := map[string]any{
		"id":       "test-identity-id-2",
		"email":    "pkce@example.com",
		"issuer":   "https://keel.so",
		"password": string(hashedPassword),
	}
	require.NoError(t, database.GetDB().Table("identity").Create(identity).Error)

	authHandler := runtime.NewAuthHandler(schema)

	tests := []struct {
		name            string
		codeVerifier    string
		expectSuccess   bool
		errorContains   string
	}{
		{
			name:          "valid_code_verifier",
			codeVerifier:  "", // Will be set to correct verifier
			expectSuccess: true,
		},
		{
			name:          "invalid_code_verifier",
			codeVerifier:  "wrong-verifier-that-does-not-match-the-challenge",
			expectSuccess: false,
			errorContains: "invalid code_verifier",
		},
		{
			name:          "missing_code_verifier",
			codeVerifier:  "",
			expectSuccess: false,
			errorContains: "code_verifier is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate PKCE
			correctVerifier := uniuri.NewLen(64)
			h := sha256.New()
			h.Write([]byte(correctVerifier))
			codeChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

			// Get auth code
			authParams := url.Values{}
			authParams.Add("response_type", "code")
			authParams.Add("redirect_uri", "http://localhost:3000/callback")
			authParams.Add("code_challenge", codeChallenge)
			authParams.Add("code_challenge_method", "S256")
			authParams.Add("username", "pkce@example.com")
			authParams.Add("password", password)

			authReq := &http.Request{
				URL: &url.URL{
					Path:     "/auth/authorize/keel",
					RawQuery: authParams.Encode(),
				},
				Method: http.MethodGet,
				Header: http.Header{},
			}
			authReq = authReq.WithContext(ctx)

			authResp := authHandler(nil, authReq)
			require.Equal(t, http.StatusFound, authResp.Status)

			location := http.Header(authResp.Headers).Get("Location")
			redirectURL, err := url.Parse(location)
			require.NoError(t, err)

			authCode := redirectURL.Query().Get("code")
			require.NotEmpty(t, authCode)

			// Use the verifier from test case, or correct one for success case
			verifier := tt.codeVerifier
			if tt.expectSuccess {
				verifier = correctVerifier
			}

			// Token request
			tokenData := url.Values{}
			tokenData.Add("grant_type", "authorization_code")
			tokenData.Add("code", authCode)
			if verifier != "" {
				tokenData.Add("code_verifier", verifier)
			}

			tokenReq := &http.Request{
				URL: &url.URL{
					Path: "/auth/token",
				},
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": {"application/x-www-form-urlencoded"},
				},
				Body: io.NopCloser(strings.NewReader(tokenData.Encode())),
			}
			tokenReq = tokenReq.WithContext(ctx)

			tokenResp := authHandler(nil, tokenReq)

			if tt.expectSuccess {
				require.Equal(t, http.StatusOK, tokenResp.Status)

				var tokenResponse map[string]any
				require.NoError(t, json.Unmarshal(tokenResp.Body, &tokenResponse))

				assert.NotEmpty(t, tokenResponse["access_token"])
			} else {
				// Should get error response
				var errorResponse map[string]any
				require.NoError(t, json.Unmarshal(tokenResp.Body, &errorResponse))

				errorMsg, ok := errorResponse["error_description"].(string)
				require.True(t, ok, "error_description not found")
				assert.Contains(t, errorMsg, tt.errorContains)
			}
		})
	}
}

// TestOAuthMetadataEndpoints tests the OAuth metadata discovery endpoints
func TestOAuthMetadataEndpoints(t *testing.T) {
	schema := protoSchema(t, `
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
	`)

	// Use the full HTTP handler to access root-level .well-known endpoints
	handler := runtime.NewHttpHandler(schema)

	ctx := context.Background()
	ctx, err := testhelpers.WithTracing(ctx)
	require.NoError(t, err)

	t.Run("protected_resource_metadata", func(t *testing.T) {
		req := &http.Request{
			URL: &url.URL{
				Path: "/.well-known/oauth-protected-resource",
			},
			Method: http.MethodGet,
			Header: http.Header{},
			Host:   "localhost:8000",
		}
		req = req.WithContext(ctx)

		recorder := &responseRecorder{}
		handler.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.status)

		var metadata map[string]any
		require.NoError(t, json.Unmarshal(recorder.body.Bytes(), &metadata))

		assert.NotEmpty(t, metadata["resource"])

		// authorization_servers is an array, need to check it exists first
		if authServers, ok := metadata["authorization_servers"].([]interface{}); ok {
			assert.Contains(t, authServers, "http://localhost:8000")
		} else {
			t.Fatalf("authorization_servers not found or wrong type: %v", metadata["authorization_servers"])
		}

		// scopes_supported
		if scopes, ok := metadata["scopes_supported"].([]interface{}); ok {
			assert.NotEmpty(t, scopes)
		}
	})

	t.Run("authorization_server_metadata", func(t *testing.T) {
		req := &http.Request{
			URL: &url.URL{
				Path: "/.well-known/oauth-authorization-server",
			},
			Method: http.MethodGet,
			Header: http.Header{},
			Host:   "localhost:8000",
		}
		req = req.WithContext(ctx)

		recorder := &responseRecorder{}
		handler.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.status)

		var metadata map[string]any
		require.NoError(t, json.Unmarshal(recorder.body.Bytes(), &metadata))

		assert.Equal(t, "http://localhost:8000", metadata["issuer"])
		assert.Equal(t, "http://localhost:8000/auth/token", metadata["token_endpoint"])
		assert.Equal(t, "http://localhost:8000/auth/mcp/register", metadata["registration_endpoint"])
		assert.Contains(t, metadata["grant_types_supported"], "authorization_code")
		assert.Contains(t, metadata["code_challenge_methods_supported"], "S256")
	})

	t.Run("client_registration", func(t *testing.T) {
		httpHandler := runtime.NewHttpHandler(schema)

		registrationReq := map[string]interface{}{
			"client_name":    "Test MCP Client",
			"redirect_uris":  []string{"http://localhost:3000/callback"},
			"grant_types":    []string{"authorization_code", "refresh_token"},
			"response_types": []string{"code"},
			"scope":          "tools:read tools:execute",
		}

		body, err := json.Marshal(registrationReq)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8000/auth/mcp/register", strings.NewReader(string(body)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)

		recorder := &responseRecorder{}
		httpHandler.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusCreated, recorder.status)

		var registrationResp map[string]any
		require.NoError(t, json.Unmarshal(recorder.body.Bytes(), &registrationResp))

		// Verify client_id was generated
		assert.NotEmpty(t, registrationResp["client_id"])
		clientID, ok := registrationResp["client_id"].(string)
		require.True(t, ok)
		assert.True(t, strings.HasPrefix(clientID, "keel_"))

		// Verify other fields echo back
		assert.Equal(t, "Test MCP Client", registrationResp["client_name"])
		assert.Equal(t, "none", registrationResp["token_endpoint_auth_method"])
	})
}

// TestMCPWWWAuthenticateHeader tests that MCP returns WWW-Authenticate headers on auth failures
func TestMCPWWWAuthenticateHeader(t *testing.T) {
	schema := protoSchema(t, `
		model Post {
			fields {
				title Text
			}
			actions {
				create createPost() with (title)
			}
			@permission(
				expression: ctx.isAuthenticated,
				actions: [create]
			)
		}
		api Test {
			models {
				Post
			}
		}
	`)

	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Database: "keel",
		Password: "postgres",
	}

	ctx := context.Background()
	ctx, err := testhelpers.WithTracing(ctx)
	require.NoError(t, err)

	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)

	ctx = runtimectx.WithPrivateKey(ctx, pk)
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{})

	dbName := testhelpers.DbNameForTestName("mcp_www_authenticate")
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, dbName, true)
	require.NoError(t, err)
	defer database.Close()

	ctx = db.WithDatabase(ctx, database)

	storer, err := storage.NewDbStore(ctx, database)
	require.NoError(t, err)
	ctx = runtimectx.WithStorage(ctx, storer)

	apiHandler := runtime.NewApiHandler(schema)

	// Try to call MCP without authentication
	req := &http.Request{
		URL: &url.URL{
			Path: "/test/mcp",
		},
		Method: http.MethodPost,
		Header: http.Header{},
		Body: io.NopCloser(strings.NewReader(`{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/call",
			"params": {
				"name": "Post.createPost",
				"arguments": {
					"title": "Test"
				}
			}
		}`)),
	}
	req = req.WithContext(ctx)

	resp := apiHandler(req)

	// MCP returns 200 with JSON-RPC error
	require.Equal(t, http.StatusOK, resp.Status)

	// Should have WWW-Authenticate header
	wwwAuth := http.Header(resp.Headers).Get("WWW-Authenticate")
	assert.NotEmpty(t, wwwAuth, "WWW-Authenticate header should be present")
	assert.Contains(t, wwwAuth, "Bearer")
	assert.Contains(t, wwwAuth, "resource_metadata")
	assert.Contains(t, wwwAuth, ".well-known/oauth-protected-resource")

	var mcpResponse map[string]any
	require.NoError(t, json.Unmarshal(resp.Body, &mcpResponse))

	// Should have JSON-RPC error (not a tool error)
	errorObj, ok := mcpResponse["error"].(map[string]any)
	require.True(t, ok, "should have error field in JSON-RPC response")
	assert.NotEmpty(t, errorObj["code"])
	assert.NotEmpty(t, errorObj["message"])

	// Should have keelCode in data
	if data, ok := errorObj["data"].(map[string]any); ok {
		assert.Equal(t, "ERR_PERMISSION_DENIED", data["keelCode"])
	}
}
