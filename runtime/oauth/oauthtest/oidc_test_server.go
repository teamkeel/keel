package oauthtest

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/dchest/uniuri"
	"github.com/golang-jwt/jwt/v4"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/oauth"
)

type IodcTokenResponse struct {
	authapi.TokenResponse
	IdToken string `json:"id_token"`
}

type OAuthClient struct {
	ClientId     string
	ClientSecret string
	RedirectUrl  string
}

type OidcServer struct {
	Issuer          string
	Config          map[string]any
	IdTokenLifespan time.Duration
	server          *httptest.Server
	PrivateKey      *rsa.PrivateKey
	Users           map[string]*oauth.UserClaims
	clients         []*OAuthClient
}

func (o *OidcServer) SetUser(sub string, claims *oauth.UserClaims) {
	o.Users[sub] = claims
}

func (o *OidcServer) FetchIdToken(sub string, aud []string) (string, error) {
	user, ok := o.Users[sub]
	if !ok {
		return "", errors.New("user sub not found")
	}

	now := time.Now().UTC()
	claims := oauth.IdTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    o.Issuer,
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(now.Add(o.IdTokenLifespan)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserClaims: oauth.UserClaims{
			Email: user.Email,
		},
	}

	if len(aud) > 0 {
		claims.Audience = aud
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(o.PrivateKey)
}

func (o *OidcServer) RenewPrivateKey() error {
	var err error
	o.PrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	return err
}

func (o *OidcServer) Close() {
	o.server.Close()
}

func (o *OidcServer) WithOAuthClient(client *OAuthClient) {
	o.clients = append(o.clients, client)
}

func NewServer() (*OidcServer, error) {
	oidcServer := &OidcServer{}

	// Every server has its own private key
	err := oidcServer.RenewPrivateKey()
	if err != nil {
		return nil, err
	}

	// Start test HTTP server
	oidcServer.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
		}

		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			b, err := json.Marshal(oidcServer.Config)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("cannot marshal oidc config"))
			}

			res.Body = io.NopCloser(bytes.NewReader(b))
			res.Body.Close()
		case "/jwks":
			jwks, err := generateJWKSResponse(*oidcServer.PrivateKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(fmt.Sprintf("couldn't make jwks response: %s", err)))
			}

			res.Body = io.NopCloser(bytes.NewReader([]byte(jwks)))
			res.Body.Close()
		case "/oauth2/authorize":
			values := url.Values{}

			// If an authorization request fails validation due to a missing,
			// invalid, or mismatching redirection URI, the authorization server
			// SHOULD inform the resource owner of the error and MUST NOT
			// automatically redirect the user-agent to the invalid redirection URI.
			// https://datatracker.ietf.org/doc/html/rfc6749#section-3.1.2.4
			if !r.URL.Query().Has("redirect_uri") {
				res.StatusCode = http.StatusNotFound
				response := &authapi.ErrorResponse{
					Error:            authapi.TokenErrInvalidRequest,
					ErrorDescription: "redirect uri missing",
				}

				b, _ := json.Marshal(response)
				res.Body = io.NopCloser(bytes.NewReader(b))
				res.Body.Close()
				w.WriteHeader(http.StatusBadRequest)
				break
			}

			redirectUrl, err := url.Parse(r.URL.Query().Get("redirect_uri"))
			if err != nil {
				res.StatusCode = http.StatusNotFound
				response := &authapi.ErrorResponse{
					Error:            authapi.TokenErrInvalidRequest,
					ErrorDescription: "redirect uri invalid",
				}

				b, _ := json.Marshal(response)
				res.Body = io.NopCloser(bytes.NewReader(b))
				res.Body.Close()
				w.WriteHeader(http.StatusBadRequest)
				break
			}

			if r.URL.Query().Get("response_type") != "code" {
				values.Add("error", "unsupported_response_type")
				values.Add("error_description", "only 'code' response_type supported")
				redirectUrl.RawQuery = values.Encode()
				http.Redirect(w, r, redirectUrl.String(), http.StatusFound)
				break
			}

			clientId := r.URL.Query().Get("client_id")
			var client *OAuthClient
			for _, v := range oidcServer.clients {
				if v.ClientId == clientId {
					client = v
				}
			}

			// Client not found on the server
			if client == nil {
				values.Add("error", "invalid_request")
				values.Add("error_description", "client id not registered on server")
				redirectUrl.RawQuery = values.Encode()
				http.Redirect(w, r, redirectUrl.String(), http.StatusFound)
				break
			}

			// The redirect URL does not match the server's registered client
			// https://datatracker.ietf.org/doc/html/rfc6749#section-3.1.2.4
			if client.RedirectUrl != redirectUrl.String() {
				res.StatusCode = http.StatusNotFound
				response := &authapi.ErrorResponse{
					Error:            authapi.TokenErrInvalidClient,
					ErrorDescription: "redirect uri does not match",
				}

				b, _ := json.Marshal(response)
				res.Body = io.NopCloser(bytes.NewReader(b))
				res.Body.Close()
				w.WriteHeader(http.StatusBadRequest)
				break
			}

			values.Add("iss", oidcServer.Issuer)
			values.Add("code", uniuri.NewLen(10))
			redirectUrl.RawQuery = values.Encode()

			// If the end-user denies the login request or if the request fails for reasons other than an
			// invalid client_id or redirect_uri, the OIDC server will pass any errors onto the redirect_uri.
			http.Redirect(w, r, redirectUrl.String(), http.StatusFound)

		case "/oauth2/token":
			idToken, _ := oidcServer.FetchIdToken("id|285620", []string{oidcServer.clients[0].ClientId})

			clientId := r.FormValue("client_id")
			var client *OAuthClient
			for _, v := range oidcServer.clients {
				if v.ClientId == clientId {
					client = v
				}
			}

			// Client not found on the server
			if client == nil {
				res.StatusCode = http.StatusNotFound
				response := &authapi.ErrorResponse{
					Error:            authapi.TokenErrInvalidClient,
					ErrorDescription: "client id not registered on server",
				}

				b, _ := json.Marshal(response)
				res.Body = io.NopCloser(bytes.NewReader(b))
				res.Body.Close()
				w.WriteHeader(http.StatusBadRequest)
				break
			}

			// Client secret incorrect
			if client.ClientSecret != r.FormValue("client_secret") {
				res.StatusCode = http.StatusNotFound
				response := &authapi.ErrorResponse{
					Error:            authapi.TokenErrInvalidRequest,
					ErrorDescription: "client credentials are incorrect",
				}

				b, _ := json.Marshal(response)
				res.Body = io.NopCloser(bytes.NewReader(b))
				res.Body.Close()
				w.WriteHeader(http.StatusBadRequest)
				break
			}

			tokenResponse := &IodcTokenResponse{
				IdToken: idToken,
				TokenResponse: authapi.TokenResponse{
					AccessToken:  "opaque-access-token",
					TokenType:    "bearer",
					ExpiresIn:    int(360),
					RefreshToken: "opaque-refresh-token",
				},
			}

			b, _ := json.Marshal(tokenResponse)

			w.Header().Add("Content-Type", "application/json")
			res.Body = io.NopCloser(bytes.NewReader(b))
			res.Body.Close()
		default:
			res.StatusCode = http.StatusNotFound
			res.Body = io.NopCloser(bytes.NewReader([]byte("not found")))
			res.Body.Close()
		}

		if res.Body != nil {
			_, err = io.Copy(w, res.Body)
			if err != nil {
				_, _ = w.Write([]byte(err.Error()))
			}
		}
	}))

	oidcServer.Issuer = oidcServer.server.URL
	oidcServer.Users = map[string]*oauth.UserClaims{}
	oidcServer.IdTokenLifespan = 5 * time.Minute
	oidcServer.Config = map[string]any{
		"issuer":                                oidcServer.Issuer,
		"authorization_endpoint":                fmt.Sprintf("%s/oauth2/authorize", oidcServer.Issuer),
		"token_endpoint":                        fmt.Sprintf("%s/oauth2/token", oidcServer.Issuer),
		"jwks_uri":                              fmt.Sprintf("%s/jwks", oidcServer.Issuer),
		"userinfo_endpoint":                     fmt.Sprintf("%s/userinfo", oidcServer.Issuer),
		"revocation_endpoint":                   fmt.Sprintf("%s/oauth2/revoke", oidcServer.Issuer),
		"response_types_supported":              []string{"code", "token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
		"claims_supported":                      []string{"sub", "iss", "aud", "exp", "iat", "name", "email"},
		"code_challenge_methods_supported":      []string{"plain", "S256"},
		"introspection_endpoint":                fmt.Sprintf("%s/oauth2/introspect", oidcServer.Issuer),
	}

	return oidcServer, nil
}

func generateJWKSResponse(privateKey rsa.PrivateKey) (string, error) {
	jwksResponse := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"alg": "RS256",
				"use": "sig",
				"kid": "mykey",
				"n":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}), // Exponent 65537
			},
		},
	}

	jwksBytes, err := json.Marshal(jwksResponse)
	if err != nil {
		return "", err
	}

	return string(jwksBytes), nil
}
