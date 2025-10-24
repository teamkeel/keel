package authapi

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

// KeelAuthorizeHandler handles the authorization request for the Keel native provider
// This supports PKCE for MCP OAuth 2.1 compliance
func KeelAuthorizeHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "Keel Authorize Endpoint")
		defer span.End()

		// Parse query parameters
		query := r.URL.Query()

		responseType := query.Get("response_type")
		if responseType != "code" {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "only response_type=code is supported", nil)
		}

		redirectURI := query.Get("redirect_uri")
		if redirectURI == "" {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "redirect_uri is required", nil)
		}

		// Validate redirect URI
		parsedRedirectURI, err := url.Parse(redirectURI)
		if err != nil {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "invalid redirect_uri", err)
		}

		// PKCE parameters (required for MCP)
		codeChallenge := query.Get("code_challenge")
		codeChallengeMethod := query.Get("code_challenge_method")

		// MCP requires PKCE with S256
		if codeChallenge == "" || codeChallengeMethod != "S256" {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "PKCE with S256 is required (code_challenge and code_challenge_method=S256)", nil)
		}

		// Resource parameter (RFC 8707 - required for MCP)
		resource := query.Get("resource")

		// If GET request, show login page
		if r.Method == http.MethodGet {
			// Check for credentials in query params (for backwards compatibility / programmatic access)
			username := query.Get("username")
			password := query.Get("password")

			if username == "" || password == "" {
				// Show HTML login form
				return renderLoginPage(r.URL.String(), "")
			}
			// Fall through to authentication if credentials provided
		}

		// Extract credentials from either query params (GET) or form data (POST)
		var username, password string
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "failed to parse form data", err)
			}
			username = r.FormValue("username")
			password = r.FormValue("password")
		} else {
			username = query.Get("username")
			password = query.Get("password")
		}

		if username == "" || password == "" {
			return jsonErrResponse(ctx, http.StatusBadRequest, AuthorizationErrInvalidRequest, "username and password are required for Keel provider", nil)
		}

		// Find identity
		identity, err := actions.FindIdentityByEmail(ctx, schema, username, oauth.KeelIssuer)
		if err != nil {
			if r.Method == http.MethodPost {
				return renderLoginPage(r.URL.String(), "An error occurred during authentication. Please try again.")
			}
			return common.InternalServerErrorResponse(ctx, err)
		}

		if identity == nil {
			// If this was a POST (form submission), show login page with error
			if r.Method == http.MethodPost {
				return renderLoginPage(r.URL.String(), "Invalid email or password")
			}
			return jsonErrResponse(ctx, http.StatusUnauthorized, AuthorizationErrAccessDenied, "invalid credentials", nil)
		}

		// Verify password
		passwordHash, ok := identity[parser.IdentityFieldNamePassword].(string)
		if !ok || passwordHash == "" {
			if r.Method == http.MethodPost {
				return renderLoginPage(r.URL.String(), "Password authentication not configured")
			}
			return jsonErrResponse(ctx, http.StatusUnauthorized, AuthorizationErrAccessDenied, "password authentication not configured", nil)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
			// If this was a POST (form submission), show login page with error
			if r.Method == http.MethodPost {
				return renderLoginPage(r.URL.String(), "Invalid email or password")
			}
			return jsonErrResponse(ctx, http.StatusUnauthorized, AuthorizationErrAccessDenied, "invalid credentials", nil)
		}

		identityID, ok := identity[parser.FieldNameId].(string)
		if !ok {
			return common.InternalServerErrorResponse(ctx, errors.New("identity ID not found"))
		}

		// Generate auth code with PKCE parameters
		authCode, err := oauth.NewAuthCodeWithPKCE(ctx, identityID, codeChallenge, codeChallengeMethod, resource)
		if err != nil {
			// Log the actual error for debugging
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetAttributes(
				attribute.String("error", err.Error()),
				attribute.String("identity_id", identityID),
				attribute.String("code_challenge", codeChallenge),
				attribute.String("code_challenge_method", codeChallengeMethod),
				attribute.String("resource", resource),
			)
			if r.Method == http.MethodPost {
				// Show more detailed error in development
				return renderLoginPage(r.URL.String(), "Failed to generate authorization code: "+err.Error())
			}
			return common.InternalServerErrorResponse(ctx, err)
		}

		// Redirect back with auth code
		values := url.Values{}
		values.Add("code", authCode)

		// Preserve state if provided
		if state := query.Get("state"); state != "" {
			values.Add("state", state)
		}

		parsedRedirectURI.RawQuery = values.Encode()

		return common.NewRedirectResponse(parsedRedirectURI)
	}
}

// renderLoginPage returns an HTML login form for the Keel OAuth authorization flow
func renderLoginPage(authURL string, errorMessage string) common.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Keel Authorization</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 400px;
            width: 100%;
            padding: 40px;
        }
        h1 {
            color: #333;
            font-size: 24px;
            margin-bottom: 8px;
            text-align: center;
        }
        .subtitle {
            color: #666;
            font-size: 14px;
            text-align: center;
            margin-bottom: 32px;
        }
        .error {
            background: #fee;
            border: 1px solid #fcc;
            color: #c33;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            color: #333;
            font-size: 14px;
            font-weight: 500;
            margin-bottom: 8px;
        }
        input[type="email"],
        input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.2s;
        }
        input[type="email"]:focus,
        input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            width: 100%;
            padding: 12px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.1s;
        }
        button:hover {
            transform: translateY(-1px);
        }
        button:active {
            transform: translateY(0);
        }
        .footer {
            margin-top: 24px;
            text-align: center;
            color: #999;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Sign in to Keel</h1>
        <p class="subtitle">Enter your credentials to authorize access</p>
        ` + func() string {
		if errorMessage != "" {
			return `<div class="error">` + errorMessage + `</div>`
		}
		return ""
	}() + `
        <form method="POST" action="` + authURL + `">
            <div class="form-group">
                <label for="username">Email</label>
                <input type="email" id="username" name="username" required autofocus>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            <button type="submit">Sign In</button>
        </form>
        <div class="footer">
            Powered by Keel
        </div>
    </div>
</body>
</html>`

	return common.Response{
		Status: http.StatusOK,
		Body:   []byte(html),
		Headers: map[string][]string{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	}
}
