package auth_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/patrickmn/go-cache"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util/mocks"
)

var privateKey *rsa.PrivateKey

func init() {
	auth.HttpClient = &mocks.MockClient{}
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	privateKey = pk
}

func newContext() context.Context {
	ctx := context.Background()
	NewCache(ctx)

	pk, _ := testhelpers.GetEmbeddedPrivateKey()
	ctx = runtimectx.WithPrivateKey(ctx, pk)

	return ctx
}

func NewCache(ctx context.Context) {
	auth.RequestCache = cache.New(5*time.Minute, 10*time.Minute)
	auth.JwkCache = jwk.NewCache(ctx)
}

func TestOIDCConfig(t *testing.T) {

	ctx := newContext()

	issuerUrl := "https://example.com/"

	requests := 0
	requestedUrls := []string{}

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		requestedUrls = append(requestedUrls, req.URL.String())
		requests = requests + 1
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	_, err := auth.GetOpenIDConnectConfig(ctx, issuerUrl)
	require.NoError(t, err)

	require.Equal(t, 1, requests)
	require.Contains(t, requestedUrls, issuerUrl+".well-known/openid-configuration")
}

func TestMultipleOIDCConfig(t *testing.T) {

	ctx := newContext()

	requests := 0

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		requests = requests + 1
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: "https://example.com/",
			},
			{
				Iss: "https://google.com/",
			}},
	})

	config, err := auth.GetOpenIDConnectConfig(ctx, "https://example.com/")
	require.NoError(t, err)

	require.Equal(t, 1, requests)
	require.Equal(t, "https://example.com/jwks", config.JWKSURL)

	_, err = auth.GetOpenIDConnectConfig(ctx, "https://google.com/")
	require.NoError(t, err)

	// Request again to check cache
	config, err = auth.GetOpenIDConnectConfig(ctx, "https://google.com/")
	require.NoError(t, err)

	require.Equal(t, 2, requests)
	require.Equal(t, "https://google.com/jwks", config.JWKSURL)
}

func TestOIDCConfigNoCache(t *testing.T) {

	ctx := newContext()
	issuerUrl := "https://example.com/"
	requests := 0

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		requests = requests + 1
		res, _ := OIDCMockResponse(req)

		res.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		res.Header.Set("Pragma", "no-cache")
		res.Header.Set("Expires", "0")

		return res, nil
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	_, err := auth.GetOpenIDConnectConfig(ctx, issuerUrl)
	require.NoError(t, err)

	_, err = auth.GetOpenIDConnectConfig(ctx, issuerUrl)
	require.NoError(t, err)

	require.Equal(t, 2, requests)
}

func TestUserInfo(t *testing.T) {

	ctx := newContext()
	issuerUrl := "https://example.com/"

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	requests := 0
	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		requests = requests + 1
		return OIDCMockResponse(req)
	}

	token, err := makeJWT(issuerUrl, "1234567890", []string{})
	require.NoError(t, err)

	_, err = auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)

	user, err := auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)
	require.Equal(t, "1234567890", user.Subject)
	require.Equal(t, "johndoe@example.com", user.Email)
	require.Equal(t, true, user.EmailVerified)
	require.Equal(t, "John", user.GivenName)

	// No cache headers so no caching
	require.Equal(t, 3, requests)

	token, err = makeJWT(issuerUrl, "123", []string{})
	require.NoError(t, err)

	_, err = auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)

	user, err = auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)
	require.Equal(t, "123", user.Subject)

	// Make another call for this new user
	require.Equal(t, 5, requests)
}

func TestUserInfoCache(t *testing.T) {

	ctx := newContext()
	issuerUrl := "https://example.com/"

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	requests := 0
	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		requests = requests + 1

		res, _ := OIDCMockResponse(req)

		expirationTime := time.Now().Add(10 * time.Minute)
		res.Header.Set("Cache-Control", "public, max-age=600")
		res.Header.Set("Expires", expirationTime.UTC().Format(http.TimeFormat))

		return res, nil
	}

	token, err := makeJWT(issuerUrl, "1234567890", []string{})
	require.NoError(t, err)

	_, err = auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)

	user, err := auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)
	require.Equal(t, "1234567890", user.Subject)
	require.Equal(t, "johndoe@example.com", user.Email)
	require.Equal(t, true, user.EmailVerified)
	require.Equal(t, "John", user.GivenName)

	// One request for the config and one for user info
	require.Equal(t, 2, requests)

	token, err = makeJWT(issuerUrl, "123", []string{})
	require.NoError(t, err)

	_, err = auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)

	user, err = auth.GetUserInfo(ctx, issuerUrl, token)
	require.NoError(t, err)
	require.Equal(t, "123", user.Subject)

	require.Equal(t, 3, requests)
}

func TestGetJWKS(t *testing.T) {

	ctx := newContext()
	issuerUrl := "https://example.com/"

	requests := 0

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		requests = requests + 1
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	_, err := auth.GetJWKS(ctx, issuerUrl)
	require.NoError(t, err)

	// One OIDC lookup and one JWKS fetch
	require.Equal(t, 2, requests)
}

func TestGetJWKSNoCache(t *testing.T) {

	ctx := newContext()
	issuerUrl := "https://example.com/"

	requests := 0

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		requests = requests + 1
		res, _ := OIDCMockResponse(req)

		res.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		res.Header.Set("Pragma", "no-cache")
		res.Header.Set("Expires", "0")

		return res, nil
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	_, err := auth.GetJWKS(ctx, issuerUrl)
	require.NoError(t, err)

	_, err = auth.GetJWKS(ctx, issuerUrl)
	require.NoError(t, err)

	// One OIDC lookup and two JWKS fetch
	require.Equal(t, 3, requests)
}

func TestOIDCTokenValidation(t *testing.T) {

	ctx := newContext()

	issuerUrl := "https://example.com/"

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: false,
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	token, err := makeJWT(issuerUrl, "user_1", []string{})
	require.NoError(t, err)

	litter.Dump(token)

	sub, iss, err := actions.ValidateBearerToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, "user_1", sub)
	require.Equal(t, issuerUrl, iss)

}

func TestOIDCTokenValidationIncorrectAudience(t *testing.T) {

	ctx := newContext()

	issuerUrl := "https://example.com/"

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		return OIDCMockResponse(req)
	}

	aud := "no match"

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss:      issuerUrl,
				Audience: &aud,
			},
		},
	})

	token, err := makeJWT(issuerUrl, "user_1", []string{})
	require.NoError(t, err)

	_, _, err = actions.ValidateBearerToken(ctx, token)
	require.Error(t, err)

}

func TestOIDCTokenValidationCorrectAudience(t *testing.T) {

	ctx := newContext()

	issuerUrl := "https://example.com/"

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		return OIDCMockResponse(req)
	}

	aud := "staff"

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss:      issuerUrl,
				Audience: &aud,
			},
		},
	})

	token, err := makeJWT(issuerUrl, "user_1", []string{"staff"})
	require.NoError(t, err)

	sub, iss, err := actions.ValidateBearerToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, "user_1", sub)
	require.Equal(t, issuerUrl, iss)

}

func TestOIDCTokenValidationInvalidIssuer(t *testing.T) {

	ctx := newContext()

	issuerUrl := "https://exampl"

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		body := "Not Found"
		r := io.NopCloser(bytes.NewReader([]byte(body)))
		return &http.Response{
			StatusCode:    http.StatusNotFound,
			ContentLength: int64(len(body)),
			Body:          r,
			Header:        make(http.Header),
		}, nil
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		Issuers: []runtimectx.ExternalIssuer{
			{
				Iss: issuerUrl,
			},
		},
	})

	token, err := makeJWT(issuerUrl, "user_1", []string{})
	require.NoError(t, err)

	_, _, err = actions.ValidateBearerToken(ctx, token)
	require.Error(t, err)

}

func TestAllowNoIssuer(t *testing.T) {

	ctx := newContext()

	issuerUrl := "https://example.com/"

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{})

	token, err := makeJWT(issuerUrl, "user_1", []string{})
	require.NoError(t, err)

	_, _, err = actions.ValidateBearerToken(ctx, token)
	require.Error(t, err)
}

func TestAllowAllIssuers(t *testing.T) {

	ctx := newContext()

	issuerUrl := "https://example.com/"

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	token, err := makeJWT(issuerUrl, "user_1", []string{})
	require.NoError(t, err)

	_, _, err = actions.ValidateBearerToken(ctx, token)
	require.NoError(t, err)
}

func TestEnVarFallback(t *testing.T) {

	ctx := newContext()

	mocks.DoFunc = func(req *http.Request) (*http.Response, error) {
		return OIDCMockResponse(req)
	}

	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	os.Setenv("KEEL_EXTERNAL_ISSUERS", "https://example.com/,https://google.com/")

	ctx = runtimectx.WithIssuersFromEnv(ctx)

	config, err := runtimectx.GetAuthConfig(ctx)
	require.NoError(t, err)

	require.Equal(t, 2, len(config.Issuers))
}

// Test issuer validates token
// Test issuer + audience validates token

// SAMPLE DATA
var sampleOidcConfig = `{
	"issuer": "https://%s/",
	"authorization_endpoint": "https://%s/oauth2/authorize",
	"jwks_uri": "https://%s/jwks",
	"userinfo_endpoint": "https://%s/userinfo",
	"revocation_endpoint": "https://%s/oauth2/revoke",
	"response_types_supported": ["code", "token"],
	"subject_types_supported": ["public"],
	"id_token_signing_alg_values_supported": ["RS256"],
	"scopes_supported": ["openid", "profile", "email"],
	"token_endpoint_auth_methods_supported": ["client_secret_basic"],
	"claims_supported": ["sub", "iss", "aud", "exp", "iat", "name", "email"],
	"code_challenge_methods_supported": ["plain", "S256"],
	"introspection_endpoint": "https://%s/oauth2/introspect"
  }`

var sampleUserInfo = `{
	"sub": "%s",
	"name": "John Doe",
	"given_name": "John",
	"family_name": "Doe",
	"email": "johndoe@example.com",
	"email_verified": true,
	"picture": "https://example.com/johndoe.jpg",
	"birthdate": "1980-01-15",
	"gender": "male",
	"locale": "en_US",
	"zoneinfo": "America/New_York"
  }`

func OIDCMockResponse(req *http.Request) (*http.Response, error) {

	resBody := "unknown"

	authHeader := req.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	switch req.URL.Path {
	case "/.well-known/openid-configuration":
		resBody = fmt.Sprintf(sampleOidcConfig, req.URL.Host, req.URL.Host, req.URL.Host, req.URL.Host, req.URL.Host, req.URL.Host)

	case "/userinfo":
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) { return &privateKey.PublicKey, nil })
		if err != nil {
			return nil, fmt.Errorf("Failed to parse token: %s", err)
		}
		claims := token.Claims.(jwt.MapClaims)
		sub, _ := claims["sub"].(string)

		resBody = fmt.Sprintf(sampleUserInfo, sub)

	case "/jwks":
		res, err := generateJWKSResponse(*privateKey)
		if err != nil {
			return nil, fmt.Errorf("Couldn't make jwks response: %s", err)
		}
		resBody = res
	}

	r := io.NopCloser(bytes.NewReader([]byte(resBody)))
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       r,
		Header:     make(http.Header),
	}

	return res, nil

}

func makeJWT(issuer string, sub string, audience []string) (string, error) {

	now := time.Now().UTC()

	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   sub,
		ExpiresAt: jwt.NewNumericDate(now.Add(actions.ResetTokenExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	if len(audience) > 0 {
		claims.Audience = audience
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)

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
