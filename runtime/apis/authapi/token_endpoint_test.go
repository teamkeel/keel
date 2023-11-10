package authapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/runtime/oauth/oauthtest"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
)

var authTestSchema = `model Post{}`

func TestTokenExchange_ValidNewIdentity(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.Equal(t, "bearer", validResponse.TokenType)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	sub, iss, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken, "")
	require.NoError(t, err)
	require.Equal(t, "https://keel.so", iss)

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
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

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
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	sub, iss, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken, "")
	require.NoError(t, err)
	require.Equal(t, "https://keel.so", iss)

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

	// TODO: test all the user info
}

func TestTokenExchange_ValidUpdatedIdentity(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// Set up auth config
	ctx = runtimectx.WithAuthConfig(ctx, runtimectx.AuthConfig{
		AllowAnyIssuers: true,
	})

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	var inserted []map[string]any
	database.GetDB().Raw(fmt.Sprintf("INSERT INTO identity (external_id, issuer, email) VALUES ('id|285620','%s','weaveton@keel.so') RETURNING *", server.Issuer)).Scan(&inserted)
	require.Len(t, inserted, 1)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	// Handle runtime request, expecting TokenResponse
	validResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	require.NotEmpty(t, validResponse.AccessToken)
	require.NotEmpty(t, validResponse.ExpiresIn)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))

	sub, iss, err := oauth.ValidateAccessToken(ctx, validResponse.AccessToken, "")
	require.NoError(t, err)
	require.Equal(t, "https://keel.so", iss)

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
}

func TestTokenEndpoint_HttpGet(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	request.Method = http.MethodGet

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusMethodNotAllowed, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the token endpoint only accepts POST", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_ApplicationJsonRequest(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	request.Header = http.Header{}
	request.Header.Add("Content-Type", "application/json")

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the request must be an encoded form with Content-Type application/x-www-form-urlencoded", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_MissingGrantType(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	form := url.Values{}
	form.Add("subject_token", idToken)
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the grant-type field is required with either 'refresh_token' or 'token_exchange'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenEndpoint_WrongGrantType(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("subject_token", idToken)
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "unsupported_grant_type", errorResponse.Error)
	require.Equal(t, "the only supported grants are 'refresh_token' and 'token_exchange'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchange_NoSubjectToken(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the ID token must be provided in the subject_token field", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchange_EmptySubjectToken(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the ID token in the subject_token field cannot be an empty string", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchange_WrongSubjectTokenType(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", idToken)
	form.Add("subject_token_type", "access_token")
	form.Add("requested_token_type", "access_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the only supported subject_token_type is 'id_token'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchange_WrongRequestedTokenType(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", idToken)
	form.Add("subject_token_type", "id_token")
	form.Add("requested_token_type", "id_token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, httpResponse.StatusCode)
	require.Equal(t, "invalid_request", errorResponse.Error)
	require.Equal(t, "the only supported requested_token_type is 'access_token'", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func TestTokenExchange_BadIdToken(t *testing.T) {
	ctx, database, schema := newContext(t, authTestSchema, true)
	defer database.Close()

	// OIDC test server
	server, err := oauthtest.NewOIDCServer()
	require.NoError(t, err)

	server.SetUser("id|285620", &oauth.UserClaims{
		Email: "keelson@keel.so",
		Name:  "Keelson",
	})

	// Get ID token from server
	idToken, err := server.FetchIdToken("id|285620", []string{})
	require.NoError(t, err)

	// Make a token exchange grant request
	request := makeTokenExchangeRequest(ctx, idToken)

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", "this is not a jwt token")
	request.URL.RawQuery = form.Encode()

	// Handle runtime request, expecting TokenErrorResponse
	errorResponse, httpResponse, err := handleRuntimeRequest[authapi.TokenErrorResponse](schema, request)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, httpResponse.StatusCode)
	require.Equal(t, "invalid_client", errorResponse.Error)
	require.Equal(t, "possible causes may be that the id token is invalid, has expired, or has insufficient claims", errorResponse.ErrorDescription)
	require.True(t, authapi.HasContentType(httpResponse.Header, "application/json"))
}

func newContext(t *testing.T, keelSchema string, resetDatabase bool) (context.Context, db.Database, *proto.Schema) {
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema)
	require.NoError(t, err)

	ctx := context.Background()

	// Add private key to context
	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, pk)

	ctx, err = testhelpers.WithTracing(ctx)
	require.NoError(t, err)

	// Add database to context
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, "runtime_test", resetDatabase)
	require.NoError(t, err)
	ctx = db.WithDatabase(ctx, database)

	return ctx, database, schema
}

func handleRuntimeRequest[T any](schema *proto.Schema, req *http.Request) (T, *http.Response, error) {
	var response T
	handler := runtime.NewHttpHandler(schema)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	httpResponse := w.Result()
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

func makeTokenExchangeRequest(ctx context.Context, token string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "http://mykeelapp.keel.so/auth/token", nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	form := url.Values{}
	form.Add("grant_type", "token_exchange")
	form.Add("subject_token", token)
	form.Add("subject_token_type", "id_token")
	form.Add("requested_token_type", "access_token")
	request.URL.RawQuery = form.Encode()

	request = request.WithContext(ctx)

	return request
}
