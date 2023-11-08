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
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/teamkeel/keel/runtime/oauth"
)

type OidcServer struct {
	Issuer          string
	Config          map[string]any
	IdTokenLifespan time.Duration

	server     *httptest.Server
	PrivateKey *rsa.PrivateKey
	users      map[string]*oauth.UserClaims
}

func (o *OidcServer) SetUser(sub string, claims *oauth.UserClaims) {
	o.users[sub] = claims
}

func (o *OidcServer) FetchIdToken(sub string, aud []string) (string, error) {
	user, ok := o.users[sub]
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

func NewOIDCServer() (*OidcServer, error) {
	oidcServer := &OidcServer{}

	// Every server has its own private key
	err := oidcServer.RenewPrivateKey()
	if err != nil {
		return nil, err
	}

	// Start test HTTP server
	oidcServer.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resBody := ""

		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			bytes, err := json.Marshal(oidcServer.Config)
			if err != nil {
				_, _ = w.Write([]byte("cannot marshell oidc config"))
			}

			resBody = string(bytes)
		case "/jwks":
			res, err := generateJWKSResponse(*oidcServer.PrivateKey)
			if err != nil {
				_, _ = w.Write([]byte(fmt.Sprintf("couldn't make jwks response: %s", err)))
			}
			resBody = res
		}

		res := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(resBody))),
			Header:     make(http.Header),
		}

		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
		}

		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", res.Header.Get("Content-Length"))
		_, err = io.Copy(w, res.Body)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
		}

		res.Body.Close()
	}))

	oidcServer.Issuer = oidcServer.server.URL
	oidcServer.users = map[string]*oauth.UserClaims{}
	oidcServer.IdTokenLifespan = 5 * time.Minute
	oidcServer.Config = map[string]any{
		"issuer":                                oidcServer.Issuer,
		"authorization_endpoint":                fmt.Sprintf("%s/oauth2/authorize", oidcServer.Issuer),
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
