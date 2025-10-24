package mcpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/teamkeel/keel/runtime/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ClientRegistrationRequest represents an OAuth 2.0 Dynamic Client Registration request
// per RFC 7591: https://datatracker.ietf.org/doc/html/rfc7591
type ClientRegistrationRequest struct {
	RedirectURIs             []string `json:"redirect_uris,omitempty"`
	TokenEndpointAuthMethod  string   `json:"token_endpoint_auth_method,omitempty"`
	GrantTypes               []string `json:"grant_types,omitempty"`
	ResponseTypes            []string `json:"response_types,omitempty"`
	ClientName               string   `json:"client_name,omitempty"`
	ClientURI                string   `json:"client_uri,omitempty"`
	LogoURI                  string   `json:"logo_uri,omitempty"`
	Scope                    string   `json:"scope,omitempty"`
	Contacts                 []string `json:"contacts,omitempty"`
	TosURI                   string   `json:"tos_uri,omitempty"`
	PolicyURI                string   `json:"policy_uri,omitempty"`
	JwksURI                  string   `json:"jwks_uri,omitempty"`
	SoftwareID               string   `json:"software_id,omitempty"`
	SoftwareVersion          string   `json:"software_version,omitempty"`
	CodeChallengeMethod      string   `json:"code_challenge_method,omitempty"`
}

// ClientRegistrationResponse represents the OAuth 2.0 Dynamic Client Registration response
type ClientRegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at,omitempty"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	ClientSecretExpiresAt   int64    `json:"client_secret_expires_at,omitempty"`
	RedirectURIs            []string `json:"redirect_uris,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	Contacts                []string `json:"contacts,omitempty"`
	TosURI                  string   `json:"tos_uri,omitempty"`
	PolicyURI               string   `json:"policy_uri,omitempty"`
	JwksURI                 string   `json:"jwks_uri,omitempty"`
	SoftwareID              string   `json:"software_id,omitempty"`
	SoftwareVersion         string   `json:"software_version,omitempty"`
}

// ClientRegistrationHandler implements RFC 7591 - OAuth 2.0 Dynamic Client Registration
// This allows MCP clients to register themselves dynamically
func ClientRegistrationHandler() common.HandlerFunc {
	return func(r *http.Request) common.Response {
		_, span := tracer.Start(r.Context(), "Client Registration Endpoint")
		defer span.End()

		// Only POST is supported
		if r.Method != http.MethodPost {
			span.SetStatus(codes.Error, "invalid_request")
			span.SetAttributes(attribute.String("error", "only POST method is supported"))
			return common.NewJsonResponse(http.StatusMethodNotAllowed, map[string]string{
				"error":             "invalid_request",
				"error_description": "only POST method is supported",
			}, nil)
		}

		// Parse the registration request
		var req ClientRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, "invalid_request")
			span.SetAttributes(attribute.String("error", "failed to parse registration request"))
			return common.NewJsonResponse(http.StatusBadRequest, map[string]string{
				"error":             "invalid_request",
				"error_description": "failed to parse registration request",
			}, nil)
		}

		// Generate a client ID
		clientID := "keel_" + uniuri.NewLen(32)

		// Set defaults if not provided
		if req.TokenEndpointAuthMethod == "" {
			req.TokenEndpointAuthMethod = "none" // Public client (PKCE)
		}
		if len(req.GrantTypes) == 0 {
			req.GrantTypes = []string{"authorization_code", "refresh_token"}
		}
		if len(req.ResponseTypes) == 0 {
			req.ResponseTypes = []string{"code"}
		}
		if req.Scope == "" {
			req.Scope = "tools:read tools:execute"
		}

		// Create the response
		// Note: For public clients using PKCE, we don't issue a client_secret
		response := ClientRegistrationResponse{
			ClientID:                clientID,
			ClientIDIssuedAt:        0, // Not tracking issuance time
			RedirectURIs:            req.RedirectURIs,
			TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
			GrantTypes:              req.GrantTypes,
			ResponseTypes:           req.ResponseTypes,
			ClientName:              req.ClientName,
			ClientURI:               req.ClientURI,
			LogoURI:                 req.LogoURI,
			Scope:                   req.Scope,
			Contacts:                req.Contacts,
			TosURI:                  req.TosURI,
			PolicyURI:               req.PolicyURI,
			JwksURI:                 req.JwksURI,
			SoftwareID:              req.SoftwareID,
			SoftwareVersion:         req.SoftwareVersion,
		}

		// Note: In a production system, you would:
		// 1. Validate the redirect_uris
		// 2. Store the client registration in a database
		// 3. Enforce the registered redirect_uris during authorization
		// 4. Track client usage and implement rate limiting
		//
		// For Keel's local development / MCP use case, we accept any client
		// and rely on PKCE for security rather than client authentication

		return common.NewJsonResponse(http.StatusCreated, response, nil)
	}
}
