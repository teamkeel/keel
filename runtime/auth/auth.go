package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/patrickmn/go-cache"
	"github.com/pquerna/cachecontrol"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OpenidConfig struct {
	Issuer   string `json:"issuer"`
	AuthURL  string `json:"authorization_endpoint"`
	TokenURL string `json:"token_endpoint"`

	JWKSURL     string   `json:"jwks_uri"`
	UserInfoURL string   `json:"userinfo_endpoint"`
	Algorithms  []string `json:"id_token_signing_alg_values_supported"`
}

type UserInfo struct {
	Subject       string `json:"sub"`
	Profile       string `json:"profile"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`

	// OIDC Standard claims (non-exhaustive)
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Name       string `json:"name"`
	Picture    string `json:"picture"`
	Gender     string `json:"gender"`
	Zoneinfo   string `json:"zoneinfo"`
	Locale     string `json:"locale"`
	UpdatedAt  string `json:"updated_at"`

	Claims []byte
}

var tracer = otel.Tracer("github.com/teamkeel/keel/auth")

var (
	HttpClient   HTTPClient
	RequestCache *cache.Cache
	JwkCache     *jwk.Cache
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(string) (*http.Response, error)
}

func init() {
	HttpClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	RequestCache = cache.New(5*time.Minute, 10*time.Minute)
	JwkCache = jwk.NewCache(context.Background())
}

func GetOpenIDConnectConfig(ctx context.Context, issuer string) (*OpenidConfig, error) {
	ctx, span := tracer.Start(ctx, "Fetching OpenID configuration")
	defer span.End()

	trimmed := strings.TrimSuffix(issuer, "/")
	configUrl := trimmed + "/.well-known/openid-configuration"

	span.SetAttributes(
		attribute.String("issuer", issuer),
		attribute.String("url", configUrl),
	)

	req, err := http.NewRequest("GET", configUrl, nil)
	if err != nil {
		return nil, err
	}
	body, _, err := cachedRequest(ctx, req.URL.String(), req)
	if err != nil {
		return nil, err
	}

	config := &OpenidConfig{}
	err = json.Unmarshal(body, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal: %s", err)
	}

	if issuer != config.Issuer {
		return nil, fmt.Errorf("oidc issuer did not match the issuer returned by provider, expected %q got %q", config.Issuer, issuer)
	}

	return config, nil
}

func GetUserInfo(ctx context.Context, issuer string, token string) (*UserInfo, error) {
	ctx, span := tracer.Start(ctx, "Fetch OpenID user info")
	defer span.End()

	sub, err := ExtractClaimFromToken(token, "sub")
	if err != nil {
		return nil, err
	}

	oidc, err := GetOpenIDConnectConfig(ctx, issuer)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", oidc.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	body, _, err := cachedRequest(ctx, fmt.Sprintf("%s-%s", req.URL.String(), sub), req)
	if err != nil {
		return nil, fmt.Errorf("Fetch failed: %s", err)
	}

	userInfo := &UserInfo{}
	err = json.Unmarshal(body, userInfo)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal: %s", err)
	}

	return userInfo, nil

}

func ExtractClaimFromToken(token string, claim string) (string, error) {
	// Parse the JWT without verifying the signature
	t, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("error parsing JWT: %s", err)
	}

	// Extract the claim
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("claims not found")

	}

	value, ok := claims[claim].(string)
	if !ok {
		return "", fmt.Errorf("%s claim not found or not a string", claim)
	}

	return value, nil
}

func GetJWKS(ctx context.Context, issuer string) (jwk.Set, error) {
	ctx, span := tracer.Start(ctx, "External issuer")
	defer span.End()

	span.SetAttributes(attribute.String("issuer", issuer))

	authConfig, err := runtimectx.GetAuthConfig(ctx)
	if err != nil {
		return nil, err
	}

	match := false
	for _, iss := range authConfig.Issuers {
		if iss.Iss == issuer {
			match = true
		}
	}

	if !match {
		return nil, errors.New("unknown issuer")
	}

	odic, err := GetOpenIDConnectConfig(ctx, issuer)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.String("JWK cache", "miss"))

	if JwkCache.IsRegistered(odic.JWKSURL) {
		cachedJwk, err := JwkCache.Get(ctx, odic.JWKSURL)
		if err == nil {
			span.SetAttributes(attribute.String("JWK cache", "hit"))
			return cachedJwk, nil
		}
		// Value in cache has expired, fetch it
		jwk, err := FetchJWKS(ctx, odic.JWKSURL)
		if err != nil {
			return nil, err
		}
		return *jwk, nil
	}

	jwk, err := RegisterAndFetchJWKS(ctx, issuer, odic.JWKSURL)
	if err != nil {
		return nil, err
	}

	return *jwk, err
}

func RegisterAndFetchJWKS(ctx context.Context, issuer string, url string) (*jwk.Set, error) {
	ctx, span := tracer.Start(ctx, "Fetch JWKs")
	defer span.End()

	err := JwkCache.Register(url, jwk.WithHTTPClient(HttpClient))
	if err != nil {
		return nil, err
	}

	// Check the url is valid
	jwk, err := JwkCache.Refresh(ctx, url)
	if err != nil {
		return nil, err
	}

	return &jwk, nil
}

func FetchJWKS(ctx context.Context, url string) (*jwk.Set, error) {
	ctx, span := tracer.Start(ctx, "Fetch JWKs")
	defer span.End()

	jwk, err := JwkCache.Refresh(ctx, url)
	if err != nil {
		return nil, err
	}

	return &jwk, nil
}

func cachedRequest(ctx context.Context, key string, req *http.Request) (body []byte, cacheHit bool, err error) {

	span := trace.SpanFromContext(ctx)

	if cached, found := RequestCache.Get(key); found {
		span.SetAttributes(attribute.String("cache", "hit"))
		cachedBody := cached.([]byte)
		return cachedBody, true, nil
	}

	span.SetAttributes(attribute.String("cache", "miss"))

	resp, err := HttpClient.Do(req)
	if err != nil {
		return []byte{}, cacheHit, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, cacheHit, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("failed to fetch url: %s  Status: %d  ", req.URL.String(), resp.StatusCode)
	}

	// Cache the response based on the cache control information
	reasons, expires, err := cachecontrol.CachableResponse(req, resp, cachecontrol.Options{})
	if err == nil {
		shouldCache := len(reasons) == 0

		if shouldCache {
			duration := time.Until(expires)
			RequestCache.Set(key, body, duration)
		}
	}

	return body, cacheHit, nil
}

func PublicKeyForIssuer(ctx context.Context, issuerUri string, tokenKid string) (*rsa.PublicKey, error) {
	jwks, err := GetJWKS(ctx, issuerUri)

	if err != nil {
		return nil, err
	}

	publicKey, err := ExtractJWKSPublicKey(ctx, jwks, tokenKid)

	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func ExtractJWKSPublicKey(ctx context.Context, jwks jwk.Set, tokenKid string) (*rsa.PublicKey, error) {
	allKeys := jwks.Keys(ctx)
	found := false
	var publicKey rsa.PublicKey

	span := trace.SpanFromContext(ctx)

	if jwks.Len() > 1 && tokenKid == "" {
		span.AddEvent("Multiple jwks but no kid in token, using first jwk")
	}

	for allKeys.Next(ctx) {
		curr := allKeys.Pair()

		switch v := curr.Value.(type) {
		case jwk.RSAPublicKey:
			kid := v.KeyID()

			if tokenKid == "" || tokenKid == kid {
				found = true
				err := v.Raw(&publicKey)
				if err != nil {
					found = false

				}

				if found {
					break
				}
			}
		}
	}

	if !found {
		return nil, errors.New("no RSA public key found")
	}

	return &publicKey, nil
}
