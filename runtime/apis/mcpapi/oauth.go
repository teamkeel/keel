package mcpapi

import (
	"net/http"

	"github.com/teamkeel/keel/runtime/common"
)

// ProtectedResourceMetadataHandler implements RFC 9728 - OAuth 2.0 Protected Resource Metadata
// This endpoint advertises where MCP clients can find the authorization server
func ProtectedResourceMetadataHandler() common.HandlerFunc {
	return func(r *http.Request) common.Response {
		baseURL := buildBaseURL(r)

		metadata := map[string]interface{}{
			"resource": baseURL,
			"authorization_servers": []string{
				baseURL,
			},
			"scopes_supported": []string{
				"tools:read",
				"tools:execute",
			},
		}

		return common.NewJsonResponse(http.StatusOK, metadata, nil)
	}
}

// AuthorizationServerMetadataHandler implements RFC 8414 - OAuth 2.0 Authorization Server Metadata
// This endpoint advertises the OAuth endpoints and capabilities
func AuthorizationServerMetadataHandler() common.HandlerFunc {
	return func(r *http.Request) common.Response {
		baseURL := buildBaseURL(r)

		metadata := map[string]interface{}{
			"issuer":                                baseURL,
			"authorization_endpoint":                baseURL + "/auth/authorize/keel",
			"token_endpoint":                        baseURL + "/auth/token",
			"revocation_endpoint":                   baseURL + "/auth/mcp/revoke",
			"registration_endpoint":                 baseURL + "/auth/mcp/register",
			"token_endpoint_auth_methods_supported": []string{"none"}, // Public client (PKCE)
			"code_challenge_methods_supported":      []string{"S256"},
			"grant_types_supported": []string{
				"authorization_code",
				"refresh_token",
			},
			"response_types_supported": []string{"code"},
			"scopes_supported": []string{
				"tools:read",
				"tools:execute",
			},
			"resource_documentation": baseURL + "/docs",
		}

		return common.NewJsonResponse(http.StatusOK, metadata, nil)
	}
}
