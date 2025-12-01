package authapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
	keeltesting "github.com/teamkeel/keel/testing"
)

var authTestSchema = `model Post{}`

func TestTokenExchange_ValidNewIdentity(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
		Name:          "name claim",
		GivenName:     "given name claim",
		FamilyName:    "family name claim",
		MiddleName:    "middle name claim",
		NickName:      "nick name claim",
		Profile:       "profile claim",
		Picture:       "picture claim",
		Website:       "website claim",
		Gender:        "gender claim",
		ZoneInfo:      "zoneInfo claim",
		Locale:        "locale claim",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.Equal(t, "bearer", validResponse.TokenType)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.NotEmpty(t, validResponse.RefreshToken)
	require.True(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	sub, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)

	emailVerified, ok := identities[0]["email_verified"].(bool)
	require.True(t, ok)
	require.Equal(t, true, emailVerified)

	name, ok := identities[0]["name"].(string)
	require.True(t, ok)
	require.Equal(t, "name claim", name)

	givenName, ok := identities[0]["given_name"].(string)
	require.True(t, ok)
	require.Equal(t, "given name claim", givenName)

	familyName, ok := identities[0]["family_name"].(string)
	require.True(t, ok)
	require.Equal(t, "family name claim", familyName)

	middleName, ok := identities[0]["middle_name"].(string)
	require.True(t, ok)
	require.Equal(t, "middle name claim", middleName)

	nickName, ok := identities[0]["nick_name"].(string)
	require.True(t, ok)
	require.Equal(t, "nick name claim", nickName)

	profile, ok := identities[0]["profile"].(string)
	require.True(t, ok)
	require.Equal(t, "profile claim", profile)

	picture, ok := identities[0]["picture"].(string)
	require.True(t, ok)
	require.Equal(t, "picture claim", picture)

	website, ok := identities[0]["website"].(string)
	require.True(t, ok)
	require.Equal(t, "website claim", website)

	gender, ok := identities[0]["gender"].(string)
	require.True(t, ok)
	require.Equal(t, "gender claim", gender)

	zoneInfo, ok := identities[0]["zone_info"].(string)
	require.True(t, ok)
	require.Equal(t, "zoneInfo claim", zoneInfo)

	locale, ok := identities[0]["locale"].(string)
	require.True(t, ok)
	require.Equal(t, "locale claim", locale)
}

func TestTokenExchange_CustomClaims(t *testing.T) {
	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx := runtimectx.WithOAuthConfig(t.Context(), &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
		Claims: []config.IdentityClaim{
			{Key: "https://slack.com/#teamID", Field: "teamId"},
			{Key: "custom_claim", Field: "customClaim"},
			{Key: "not_exists", Field: "notExists"},
		},
	})

	ctx, database, schema := keeltesting.MakeContext(t, ctx, authTestSchema, true)
	defer database.Close()

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.Equal(t, "bearer", validResponse.TokenType)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.NotEmpty(t, validResponse.RefreshToken)
	require.True(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	sub, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	teamId, ok := identities[0]["team_id"].(string)
	require.True(t, ok)
	require.Equal(t, "342352392354", teamId)

	customClaim, ok := identities[0]["custom_claim"].(string)
	require.True(t, ok)
	require.Equal(t, "custom value", customClaim)

	notExists := identities[0]["not_exists"]
	require.Nil(t, notExists)
}

func TestTokenExchange_ValidNewIdentityNoProfileClaims(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.Equal(t, "bearer", validResponse.TokenType)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.NotEmpty(t, validResponse.RefreshToken)
	require.True(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	sub, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)

	emailVerified, ok := identities[0]["email_verified"].(bool)
	require.True(t, ok)
	require.Equal(t, true, emailVerified)

	_, ok = identities[0]["name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["given_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["family_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["middle_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["nick_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["profile"].(string)
	require.False(t, ok)

	_, ok = identities[0]["picture"].(string)
	require.False(t, ok)

	_, ok = identities[0]["website"].(string)
	require.False(t, ok)

	_, ok = identities[0]["gender"].(string)
	require.False(t, ok)

	_, ok = identities[0]["zone_info"].(string)
	require.False(t, ok)

	_, ok = identities[0]["locale"].(string)
	require.False(t, ok)
}

func TestTokenExchange_ValidNewIdentityNoEmailClaims(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.Equal(t, "bearer", validResponse.TokenType)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.NotEmpty(t, validResponse.RefreshToken)
	require.True(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	sub, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	_, ok = identities[0]["email"].(string)
	require.False(t, ok)

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)

	emailVerified, ok := identities[0]["email_verified"].(bool)
	require.True(t, ok)
	require.Equal(t, false, emailVerified)

	_, ok = identities[0]["name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["given_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["family_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["middle_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["nick_name"].(string)
	require.False(t, ok)

	_, ok = identities[0]["profile"].(string)
	require.False(t, ok)

	_, ok = identities[0]["picture"].(string)
	require.False(t, ok)

	_, ok = identities[0]["website"].(string)
	require.False(t, ok)

	_, ok = identities[0]["gender"].(string)
	require.False(t, ok)

	_, ok = identities[0]["zone_info"].(string)
	require.False(t, ok)

	_, ok = identities[0]["locale"].(string)
	require.False(t, ok)
}

func TestTokenExchangeWithJson_ValidNewIdentity(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeJsonRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.Equal(t, "bearer", validResponse.TokenType)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.NotEmpty(t, validResponse.RefreshToken)
	require.True(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	sub, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)
}

func TestTokenExchange_ValidNewIdentityAllUserInfo(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:               "keelson@keel.so",
		EmailVerified:       true,
		Name:                "Keely",
		GivenName:           "Keel",
		FamilyName:          "Keelson",
		MiddleName:          "Kool",
		NickName:            "Koolio",
		PreferredUsername:   "keel",
		Profile:             "https://github.com/teamkeel",
		Picture:             "https://avatars.githubusercontent.com/u/102726482?s=200&v=4",
		Website:             "https://keel.so",
		Gender:              "Unknown",
		ZoneInfo:            "Europe/Paris",
		Locale:              "fr-CA",
		PhoneNumber:         "+99 (999) 999-9999",
		PhoneNumberVerified: false,
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	sub, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)
}

func TestTokenExchange_ValidUpdatedIdentity(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	var inserted []map[string]any
	database.GetDB().Raw(fmt.Sprintf("INSERT INTO identity (external_id, issuer, email) VALUES ('id|285620','%s','weaveton@keel.so') RETURNING *", server.Issuer)).Scan(&inserted)
	require.Len(t, inserted, 1)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email:         "keelson@keel.so",
		EmailVerified: true,
		Name:          "name claim",
		GivenName:     "given name claim",
		FamilyName:    "family name claim",
		MiddleName:    "middle name claim",
		NickName:      "nick name claim",
		Profile:       "profile claim",
		Picture:       "picture claim",
		Website:       "website claim",
		Gender:        "gender claim",
		ZoneInfo:      "zoneInfo claim",
		Locale:        "locale claim",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.False(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	sub, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken)
	require.NoError(t, err)

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	id, ok := identities[0]["id"].(string)
	require.True(t, ok)
	require.Equal(t, id, sub)
	require.Equal(t, id, inserted[0]["id"].(string))

	email, ok := identities[0]["email"].(string)
	require.True(t, ok)
	require.Equal(t, email, "keelson@keel.so")

	externalId, ok := identities[0]["external_id"].(string)
	require.True(t, ok)
	require.Equal(t, "id|285620", externalId)

	issuer, ok := identities[0]["issuer"].(string)
	require.True(t, ok)
	require.Equal(t, issuer, server.Issuer)

	emailVerified, ok := identities[0]["email_verified"].(bool)
	require.True(t, ok)
	require.Equal(t, true, emailVerified)

	name, ok := identities[0]["name"].(string)
	require.True(t, ok)
	require.Equal(t, "name claim", name)

	givenName, ok := identities[0]["given_name"].(string)
	require.True(t, ok)
	require.Equal(t, "given name claim", givenName)

	familyName, ok := identities[0]["family_name"].(string)
	require.True(t, ok)
	require.Equal(t, "family name claim", familyName)

	middleName, ok := identities[0]["middle_name"].(string)
	require.True(t, ok)
	require.Equal(t, "middle name claim", middleName)

	nickName, ok := identities[0]["nick_name"].(string)
	require.True(t, ok)
	require.Equal(t, "nick name claim", nickName)

	profile, ok := identities[0]["profile"].(string)
	require.True(t, ok)
	require.Equal(t, "profile claim", profile)

	picture, ok := identities[0]["picture"].(string)
	require.True(t, ok)
	require.Equal(t, "picture claim", picture)

	website, ok := identities[0]["website"].(string)
	require.True(t, ok)
	require.Equal(t, "website claim", website)

	gender, ok := identities[0]["gender"].(string)
	require.True(t, ok)
	require.Equal(t, "gender claim", gender)

	zoneInfo, ok := identities[0]["zone_info"].(string)
	require.True(t, ok)
	require.Equal(t, "zoneInfo claim", zoneInfo)

	locale, ok := identities[0]["locale"].(string)
	require.True(t, ok)
	require.Equal(t, "locale claim", locale)
}

func TestTokenExchangeCreateIfNotExistsFalse_IdentityNotExists(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	createIfNotExists := false
	request := makeTokenExchangeFormRequest(ctx, idToken, &createIfNotExists)

	// Handle runtime request, expecting TokenResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "the identity does not exist", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeCreateIfNotExistsTrue_IdentityAuthorisedAndNotCreated(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	var inserted []map[string]any
	database.GetDB().Raw(fmt.Sprintf("INSERT INTO identity (external_id, issuer, email) VALUES ('id|285620','%s','keelson@keel.so') RETURNING *", server.Issuer)).Scan(&inserted)
	require.Len(t, inserted, 1)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	createIfNotExists := true
	request := makeTokenExchangeFormRequest(ctx, idToken, &createIfNotExists)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.False(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeCreateIfNotExistsTrue_IdentityCreated(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	createIfNotExists := true
	request := makeTokenExchangeFormRequest(ctx, idToken, &createIfNotExists)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.True(t, validResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_HttpGet(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)

	request.Method = http.MethodGet

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusMethodNotAllowed, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the token endpoint only accepts POST", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_ApplicationTextRequest(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	request.Header = http.Header{}
	request.Header.Add("Content-Type", "application/text")

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the request body must either be an encoded form (Content-Type: application/x-www-form-urlencoded) or JSON (Content-Type: application/json)", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpointJson_MissingGrantType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	values := map[string]string{
		"subject_token":        "mock_token",
		"subject_token_type":   "id_token",
		"requested_token_type": "access_token",
	}

	jsonValue, _ := json.Marshal(values)
	responseBody := bytes.NewBuffer(jsonValue)

	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", responseBody)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the grant_type field is required with either 'refresh_token', 'token_exchange', 'authorization_code' or 'password'", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_MissingGrantType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	form := url.Values{}
	form.Add("subject_token", "mock_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the grant_type field is required with either 'refresh_token', 'token_exchange', 'authorization_code' or 'password'", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_WrongGrantType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	form := url.Values{}
	form.Add("grant_type", "unknown")
	form.Add("subject_token", "mock_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "unsupported_grant_type", errorResponse.Error)
	require.Equal(t, "the only supported grants are 'refresh_token', 'token_exchange', 'authorization_code' or 'password'", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_NoSubjectToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the ID token must be provided in the 'subject_token' field", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpointJson_NoSubjectToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	values := map[string]string{
		"grant_type": "token_exchange",
	}

	jsonValue, _ := json.Marshal(values)
	responseBody := bytes.NewBuffer(jsonValue)

	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", responseBody)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the ID token must be provided in the 'subject_token' field", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_EmptySubjectToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the ID token must be provided in the 'subject_token' field", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_WrongSubjectTokenType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "mock_token")
	form.Add("subject_token_type", "access_token")
	form.Add("requested_token_type", "access_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the only supported subject_token_type is 'id_token'", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_WrongRequestedTokenType(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "mock_token")
	form.Add("subject_token_type", "id_token")
	form.Add("requested_token_type", "id_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the only supported requested_token_type is 'access_token'", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_InvalidCreateIfNotExists(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, "mock_token", nil)
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "mock_token")
	form.Add("subject_token_type", "id_token")
	form.Add("create_if_not_exists", "invalid")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the create_if_not_exists field is invalid and must be either 'true' or 'false'", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchangeGrant_BadIdToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)
	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "this is not a jwt token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "access denied", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestRefreshTokenGrantJson_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	refreshTokenRotation := true
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Tokens: config.TokensConfig{
			RefreshTokenRotationEnabled: &refreshTokenRotation,
		},
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeJsonRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	tokenExchangeResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// We need 1 second to pass in order to get a different access token
	time.Sleep(1000 * time.Millisecond)

	// Make a refresh token grant request
	request = makeRefreshTokenJsonRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, refreshGrantResponse.AccessToken)
	require.Equal(t, "bearer", refreshGrantResponse.TokenType)
	require.NotEmpty(t, refreshGrantResponse.ExpiresIn)
	require.NotEmpty(t, refreshGrantResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.RefreshToken, tokenExchangeResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.AccessToken, tokenExchangeResponse.AccessToken)
	require.False(t, refreshGrantResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	accessToken1Issuer, err := oauth.ExtractClaimFromJwt(tokenExchangeResponse.AccessToken, "iss")
	require.NoError(t, err)
	accessToken2Issuer, err := oauth.ExtractClaimFromJwt(refreshGrantResponse.AccessToken, "iss")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Issuer)
	require.Equal(t, accessToken1Issuer, accessToken2Issuer)

	accessToken1Sub, err := oauth.ExtractClaimFromJwt(tokenExchangeResponse.AccessToken, "sub")
	require.NoError(t, err)
	accessToken2Sub, err := oauth.ExtractClaimFromJwt(refreshGrantResponse.AccessToken, "sub")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Sub)
	require.Equal(t, accessToken1Sub, accessToken2Sub)

	// Make a refresh token grant request using the original refresh token
	request = makeRefreshTokenJsonRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenErrorResponse
	secondRefreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "possible causes may be that the refresh token has been revoked or has expired", secondRefreshGrantResponse.ErrorDescription)
}

func TestRefreshTokenGrantRotationEnabled_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	refreshTokenRotation := true
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Tokens: config.TokensConfig{
			RefreshTokenRotationEnabled: &refreshTokenRotation,
		},
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	tokenExchangeResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// We need 1 second to pass in order to get a different access token
	time.Sleep(1000 * time.Millisecond)

	// Make a refresh token grant request
	request = makeRefreshTokenFormRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, refreshGrantResponse.AccessToken)
	require.Equal(t, "bearer", refreshGrantResponse.TokenType)
	require.NotEmpty(t, refreshGrantResponse.ExpiresIn)
	require.NotEmpty(t, refreshGrantResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.RefreshToken, tokenExchangeResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.AccessToken, tokenExchangeResponse.AccessToken)
	require.False(t, refreshGrantResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	accessToken1Issuer, err := oauth.ExtractClaimFromJwt(tokenExchangeResponse.AccessToken, "iss")
	require.NoError(t, err)
	accessToken2Issuer, err := oauth.ExtractClaimFromJwt(refreshGrantResponse.AccessToken, "iss")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Issuer)
	require.Equal(t, accessToken1Issuer, accessToken2Issuer)

	accessToken1Sub, err := oauth.ExtractClaimFromJwt(tokenExchangeResponse.AccessToken, "sub")
	require.NoError(t, err)
	accessToken2Sub, err := oauth.ExtractClaimFromJwt(refreshGrantResponse.AccessToken, "sub")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Sub)
	require.Equal(t, accessToken1Sub, accessToken2Sub)

	// Make a refresh token grant request using the original refresh token
	request = makeRefreshTokenFormRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenErrorResponse
	secondRefreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "possible causes may be that the refresh token has been revoked or has expired", secondRefreshGrantResponse.ErrorDescription)
}

func TestRefreshTokenGrantRotationDisabled_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewServer()
	require.NoError(t, err)
	defer server.Close()

	// Set up auth config
	refreshTokenRotation := false
	ctx = runtimectx.WithOAuthConfig(ctx, &config.AuthConfig{
		Tokens: config.TokensConfig{
			RefreshTokenRotationEnabled: &refreshTokenRotation,
		},
		Providers: []config.Provider{
			{
				Type:      config.OpenIdConnectProvider,
				Name:      "my-oidc",
				ClientId:  "oidc-client-id",
				IssuerUrl: server.Issuer,
			},
		},
	})

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{"oidc-client-id"})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeFormRequest(ctx, idToken, nil)

	// Handle runtime request, expecting TokenResponse
	tokenExchangeResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// We need 1 second to pass in order to get a different access token
	time.Sleep(1000 * time.Millisecond)

	// Make a refresh token grant request
	request = makeRefreshTokenFormRequest(ctx, tokenExchangeResponse.RefreshToken)

	// Handle runtime request, expecting TokenResponse
	refreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, refreshGrantResponse.AccessToken)
	require.Equal(t, "bearer", refreshGrantResponse.TokenType)
	require.NotEmpty(t, refreshGrantResponse.ExpiresIn)
	require.NotEmpty(t, refreshGrantResponse.RefreshToken)
	require.Equal(t, refreshGrantResponse.RefreshToken, tokenExchangeResponse.RefreshToken)
	require.NotEqual(t, refreshGrantResponse.AccessToken, tokenExchangeResponse.AccessToken)
	require.False(t, refreshGrantResponse.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	accessToken1Issuer, err := oauth.ExtractClaimFromJwt(tokenExchangeResponse.AccessToken, "iss")
	require.NoError(t, err)
	accessToken2Issuer, err := oauth.ExtractClaimFromJwt(refreshGrantResponse.AccessToken, "iss")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Issuer)
	require.Equal(t, accessToken1Issuer, accessToken2Issuer)

	accessToken1Sub, err := oauth.ExtractClaimFromJwt(tokenExchangeResponse.AccessToken, "sub")
	require.NoError(t, err)
	accessToken2Sub, err := oauth.ExtractClaimFromJwt(refreshGrantResponse.AccessToken, "sub")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken1Sub)
	require.Equal(t, accessToken1Sub, accessToken2Sub)

	// Make a refresh token grant request using the original refresh token
	request = makeRefreshTokenFormRequest(ctx, tokenExchangeResponse.RefreshToken)

	secondRefreshGrantResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.Equal(t, tokenExchangeResponse.RefreshToken, secondRefreshGrantResponse.RefreshToken)
}

func TestRefreshTokenGrant_NoRefreshToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a refresh token grant request
	request := makeRefreshTokenFormRequest(ctx, "")
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token in the 'refresh_token' field is required", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestRefreshTokenGrant_EmptyRefreshToken(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a refresh token grant request
	request := makeRefreshTokenFormRequest(ctx, "")

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the refresh token in the 'refresh_token' field is required", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestAuthorizationCodeGrant_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)

	code, err := oauth.NewAuthCode(ctx, identity["id"].(string))
	require.NoError(t, err)

	// Make a auth code grant request
	request := makeAuthorizationCodeFormRequest(ctx, code)

	// Handle runtime request, expecting TokenResponse
	response, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, response.AccessToken)
	require.Equal(t, "bearer", response.TokenType)
	require.NotEmpty(t, response.ExpiresIn)
	require.NotEmpty(t, response.RefreshToken)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	accessTokenIssuer, err := oauth.ExtractClaimFromJwt(response.AccessToken, "iss")
	require.NoError(t, err)
	require.Equal(t, accessTokenIssuer, "https://keel.so")

	accessTokenSub, err := oauth.ExtractClaimFromJwt(response.AccessToken, "sub")
	require.NoError(t, err)
	require.Equal(t, accessTokenSub, identity["id"].(string))
}

func TestAuthorizationCodeGrantJson_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	identity, err := actions.CreateIdentity(ctx, schema, "test@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)

	code, err := oauth.NewAuthCode(ctx, identity["id"].(string))
	require.NoError(t, err)

	// Make a auth code grant request
	request := makeAuthorizationCodeJsonRequest(ctx, code)

	// Handle runtime request, expecting TokenResponse
	response, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, response.AccessToken)
	require.Equal(t, "bearer", response.TokenType)
	require.NotEmpty(t, response.ExpiresIn)
	require.NotEmpty(t, response.RefreshToken)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	accessTokenIssuer, err := oauth.ExtractClaimFromJwt(response.AccessToken, "iss")
	require.NoError(t, err)
	require.Equal(t, accessTokenIssuer, "https://keel.so")

	accessTokenSub, err := oauth.ExtractClaimFromJwt(response.AccessToken, "sub")
	require.NoError(t, err)
	require.Equal(t, accessTokenSub, identity["id"].(string))
}

func TestAuthorizationCodeGrant_InvalidCode(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a auth code grant request
	request := makeAuthorizationCodeFormRequest(ctx, "whoops")

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "possible causes may be that the auth code has been consumed or has expired", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestAuthorizationCodeGrantForm_NoCode(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a auth code grant request
	request := makeAuthorizationCodeFormRequest(ctx, "")
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the authorization code in the 'code' field is required", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestAuthorizationCodeGrantJson_NoCode(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a auth code grant request
	request := makeAuthorizationCodeFormRequest(ctx, "")
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting ErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the authorization code in the 'code' field is required", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestPasswordGrantForm_IdentityCreated(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", nil)

	// Handle runtime request, expecting TokenResponse
	response, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, response.AccessToken)
	require.Equal(t, "bearer", response.TokenType)
	require.NotEmpty(t, response.ExpiresIn)
	require.NotEmpty(t, response.RefreshToken)
	require.True(t, response.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	accessTokenIssuer, err := oauth.ExtractClaimFromJwt(response.AccessToken, "iss")
	require.NoError(t, err)
	require.Equal(t, accessTokenIssuer, "https://keel.so")

	accessTokenSub, err := oauth.ExtractClaimFromJwt(response.AccessToken, "sub")
	require.NoError(t, err)
	require.Equal(t, accessTokenSub, identities[0]["id"])
}

func TestPasswordGrantCreateIfNotExistsTrue_IdentityCreated(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	createIfNotExists := true
	request := makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", &createIfNotExists)

	// Handle runtime request, expecting TokenResponse
	response, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, response.AccessToken)
	require.Equal(t, "bearer", response.TokenType)
	require.NotEmpty(t, response.ExpiresIn)
	require.NotEmpty(t, response.RefreshToken)
	require.True(t, response.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	accessTokenIssuer, err := oauth.ExtractClaimFromJwt(response.AccessToken, "iss")
	require.NoError(t, err)
	require.Equal(t, accessTokenIssuer, "https://keel.so")

	accessTokenSub, err := oauth.ExtractClaimFromJwt(response.AccessToken, "sub")
	require.NoError(t, err)
	require.Equal(t, accessTokenSub, identities[0]["id"])
}

func TestPasswordGrantCreateIfNotExistsFalse_IdentityNotCreated(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	createIfNotExists := false
	request := makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", &createIfNotExists)

	// Handle runtime request, expecting TokenResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "the identity does not exist or the credentials are incorrect", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestPasswordGrantCreateIfNotExistsFalse_Authorised(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	request := makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", nil)

	// Handle runtime request, expecting TokenResponse
	_, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// Make a password grant request
	createIfNotExists := false
	request = makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", &createIfNotExists)

	// Handle runtime request, expecting TokenResponse
	response, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, response.AccessToken)
	require.Equal(t, "bearer", response.TokenType)
	require.NotEmpty(t, response.ExpiresIn)
	require.NotEmpty(t, response.RefreshToken)
	require.False(t, response.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	accessTokenIssuer, err := oauth.ExtractClaimFromJwt(response.AccessToken, "iss")
	require.NoError(t, err)
	require.Equal(t, accessTokenIssuer, "https://keel.so")

	accessTokenSub, err := oauth.ExtractClaimFromJwt(response.AccessToken, "sub")
	require.NoError(t, err)
	require.Equal(t, accessTokenSub, identities[0]["id"])
}

func TestPasswordGrantJson_Valid(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordJsonRequest(ctx, "user@example.com", "myP@ssword1234!")

	// Handle runtime request, expecting TokenResponse
	response, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, response.AccessToken)
	require.Equal(t, "bearer", response.TokenType)
	require.NotEmpty(t, response.ExpiresIn)
	require.NotEmpty(t, response.RefreshToken)
	require.True(t, response.Created)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))

	var identities []map[string]any
	database.GetDB().Raw("SELECT * FROM identity").Scan(&identities)
	require.Len(t, identities, 1)

	accessTokenIssuer, err := oauth.ExtractClaimFromJwt(response.AccessToken, "iss")
	require.NoError(t, err)
	require.Equal(t, accessTokenIssuer, "https://keel.so")

	accessTokenSub, err := oauth.ExtractClaimFromJwt(response.AccessToken, "sub")
	require.NoError(t, err)
	require.Equal(t, accessTokenSub, identities[0]["id"])
}

func TestPasswordGrant_IncorrectCredentials(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", nil)

	// Handle runtime request, expecting TokenResponse
	_, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// Make another password grant request
	request = makePasswordFormRequest(ctx, "user@example.com", "whoops!", nil)

	// Handle runtime request, expecting TokenResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "the identity does not exist or the credentials are incorrect", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestPasswordGrant_CorrectCredentials(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", nil)

	// Handle runtime request, expecting TokenResponse
	_, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	// Make another password grant request
	request = makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", nil)

	// Handle runtime request, expecting TokenResponse
	_, httpResponse, err = handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
}

func TestPasswordGrant_InvalidEmail(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordFormRequest(ctx, "user", "myP@ssword1234!", nil)

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "invalid email address", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestPasswordGrant_MissingEmail(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordFormRequest(ctx, "user@example.com", "myP@ssword1234!", nil)
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("password", "myP@ssword1234!")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the identity's email in the 'username' field is required", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestPasswordGrant_MissingPassword(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordFormRequest(ctx, "user", "myP@ssword1234!", nil)
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", "user@example.com")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the identity's password in the 'password' field is required", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func TestPasswordGrant_InvalidCreateIfNotExists(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), authTestSchema, true)
	defer database.Close()

	// Make a password grant request
	request := makePasswordFormRequest(ctx, "user", "myP@ssword1234!", nil)
	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", "user@example.com")
	form.Add("create_if_not_exists", "invalid")
	request.URL.RawQuery = form.Encode()

	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.ErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the create_if_not_exists field is invalid and must be either 'true' or 'false'", errorResponse.ErrorDescription)
	require.True(t, common.HasContentType(httpResponse.Header, "application/json"))
}

func handleRuntimeRequest[T any](schema *proto.Schema, req *http.Request) (T, *http.Response, error) {
	var response T
	handler := runtime.NewHttpHandler(schema)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	httpResponse := w.Result()

	if httpResponse.StatusCode == http.StatusInternalServerError {
		return response, nil, errors.New("internal server response from oidc server")
	}

	data, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return response, nil, err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return response, nil, err
	}

	return response, httpResponse, nil
}

func makeTokenExchangeFormRequest(ctx context.Context, token string, createIfNotExists *bool) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", token)
	form.Add("subject_token_type", "id_token")
	form.Add("requested_token_type", "access_token")

	if createIfNotExists != nil {
		form.Add("create_if_not_exists", strconv.FormatBool(*createIfNotExists))
	}

	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}

func makeTokenExchangeJsonRequest(ctx context.Context, token string) *http.Request {
	values := map[string]string{
		"grant_type":           "token_exchange",
		"subject_token":        token,
		"subject_token_type":   "id_token",
		"requested_token_type": "access_token",
	}

	jsonValue, _ := json.Marshal(values)
	responseBody := bytes.NewBuffer(jsonValue)

	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", responseBody)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	return request
}

func makeRefreshTokenFormRequest(ctx context.Context, token string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", token)
	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}

func makeRefreshTokenJsonRequest(ctx context.Context, token string) *http.Request {
	values := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": token,
	}

	jsonValue, _ := json.Marshal(values)
	responseBody := bytes.NewBuffer(jsonValue)

	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", responseBody)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	return request
}

func makeAuthorizationCodeFormRequest(ctx context.Context, code string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}

func makeAuthorizationCodeJsonRequest(ctx context.Context, code string) *http.Request {
	values := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}

	jsonValue, _ := json.Marshal(values)
	responseBody := bytes.NewBuffer(jsonValue)

	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", responseBody)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	return request
}

func makePasswordFormRequest(ctx context.Context, username string, password string, createIfNotExists *bool) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", username)
	form.Add("password", password)

	if createIfNotExists != nil {
		form.Add("create_if_not_exists", strconv.FormatBool(*createIfNotExists))
	}

	request.URL.RawQuery = form.Encode()
	request = request.WithContext(ctx)

	return request
}

func makePasswordJsonRequest(ctx context.Context, username string, password string) *http.Request {
	values := map[string]string{
		"grant_type": "password",
		"username":   username,
		"password":   password,
	}

	jsonValue, _ := json.Marshal(values)
	responseBody := bytes.NewBuffer(jsonValue)

	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", responseBody)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	return request
}
