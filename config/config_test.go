package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	config, err := Load("fixtures/test_basic_config.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "TEST", config.Environment[0].Name)
	assert.Equal(t, "test", config.Environment[0].Value)
	assert.Equal(t, "API_KEY", config.Secrets[0].Name)
}

func TestFailConfigValue(t *testing.T) {
	_, err := Load("fixtures/test_failing_config.yaml")
	assert.Error(t, err)

	assert.Equal(t, "could not unmarshal config file: yaml: unmarshal errors:\n  line 4: cannot unmarshal !!seq into string", err.Error())
}

func TestDuplicates(t *testing.T) {
	config, err := Load("fixtures/test_duplicates.yaml")
	assert.Error(t, err)

	assert.Equal(t, "TEST", config.Environment[0].Name)
	assert.Equal(t, "environment variable TEST has a duplicate set\n", err.Error())
}

func TestSnakecaseValidateFormat(t *testing.T) {
	_, err := Load("fixtures/test_snakecase_config.yaml")
	assert.Error(t, err)

	assert.NotContains(t, err.Error(), "THIS_IS_ALLOWED")

	assert.Contains(t, err.Error(), "this_is_not_Allowed1 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "THIS_IS_NOT_ALLOWEd2 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "thisIsNotAllowed3 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "This_IS_nOT_AlloWED4 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "This_IS_nOT_AlloWED4 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "Not_Allowed_Secret_Name must be written in upper snakecase\n")
}

func TestReservedNameValidateFormat(t *testing.T) {
	_, err := Load("fixtures/test_reserved_name_config.yaml")
	assert.Error(t, err)

	assert.NotContains(t, err.Error(), "THIS_IS_ALLOWED")

	assert.Contains(t, err.Error(), "environment variable AWS_NOT_ALLOWED1 cannot start with AWS_ as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable KEEL_NOT_ALLOWED2 cannot start with KEEL_ as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable OTEL_NOT_ALLOWED3 cannot start with OTEL_ as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable OPENCOLLECTOR_CONFIG_NOT_ALLOWED4 cannot start with OPENCOLLECTOR_CONFIG as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable _NOT_ALLOWED5 cannot start with _ as it is reserved\n")
}

func TestDefaultApiOmitted(t *testing.T) {
	config, err := Load("fixtures/test_empty_config.yaml")
	assert.NoError(t, err)

	assert.True(t, config.DefaultApi())
}
func TestDefaultApiTrue(t *testing.T) {
	config, err := Load("fixtures/test_default_api_true.yaml")
	assert.NoError(t, err)

	assert.True(t, config.DefaultApi())
}
func TestDefaultApiFalse(t *testing.T) {
	config, err := Load("fixtures/test_default_api_false.yaml")
	assert.NoError(t, err)

	assert.False(t, config.DefaultApi())
}

func TestAuthTokens(t *testing.T) {
	config, err := Load("fixtures/test_auth.yaml")
	assert.NoError(t, err)

	assert.Equal(t, 3600, *config.Auth.Tokens.AccessTokenExpiry)
	assert.Equal(t, 604800, *config.Auth.Tokens.RefreshTokenExpiry)
	assert.Equal(t, false, *config.Auth.Tokens.RefreshTokenRotationEnabled)

	assert.Equal(t, time.Duration(3600)*time.Second, config.Auth.AccessTokenExpiry())
	assert.Equal(t, time.Duration(604800)*time.Second, config.Auth.RefreshTokenExpiry())
	assert.Equal(t, false, config.Auth.RefreshTokenRotationEnabled())
}

func TestAuthInvalidRedirectUrl(t *testing.T) {
	_, err := Load("fixtures/test_auth_invalid_redirect_url.yaml")

	assert.Contains(t, err.Error(), "auth redirectUrl 'not a url' is not a valid url\n")
}

func TestAuthDefaults(t *testing.T) {
	config, err := Load("fixtures/test_empty_config.yaml")
	assert.NoError(t, err)

	assert.Nil(t, config.Auth.Tokens.AccessTokenExpiry)
	assert.Nil(t, config.Auth.Tokens.RefreshTokenExpiry)
	assert.Nil(t, config.Auth.Tokens.RefreshTokenRotationEnabled)

	assert.Equal(t, time.Duration(24)*time.Hour, config.Auth.AccessTokenExpiry())
	assert.Equal(t, time.Duration(24)*time.Hour*90, config.Auth.RefreshTokenExpiry())
	assert.Equal(t, true, config.Auth.RefreshTokenRotationEnabled())
}

func TestAuthNegativeTokenLifespan(t *testing.T) {
	_, err := Load("fixtures/test_auth_negative_token_lifespan.yaml")

	assert.Contains(t, err.Error(), "access token lifespan cannot be negative or zero for field: accessTokenExpiry\n")
	assert.Contains(t, err.Error(), "refresh token lifespan cannot be negative or zero for field: refreshTokenExpiry\n")
}

func TestAuthProviders(t *testing.T) {
	config, err := Load("fixtures/test_auth.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "google", config.Auth.Providers[0].Type)
	assert.Equal(t, "_Google_Client", config.Auth.Providers[0].Name)
	assert.Equal(t, "foo_1", config.Auth.Providers[0].ClientId)

	assert.Equal(t, "google", config.Auth.Providers[1].Type)
	assert.Equal(t, "google_2", config.Auth.Providers[1].Name)
	assert.Equal(t, "foo_2", config.Auth.Providers[1].ClientId)

	assert.Equal(t, "oidc", config.Auth.Providers[2].Type)
	assert.Equal(t, "Baidu", config.Auth.Providers[2].Name)
	assert.Equal(t, "https://dev-skhlutl45lbqkvhv.us.auth0.com", config.Auth.Providers[2].IssuerUrl)
	assert.Equal(t, "kasj28fnq09ak", config.Auth.Providers[2].ClientId)
}

func TestInvalidProviderName(t *testing.T) {
	_, err := Load("fixtures/test_auth_invalid_names.yaml")

	assert.Contains(t, err.Error(), "auth provider name '12 34' must only include alphanumeric characters and underscores, and cannot start with a number\n")
	assert.Contains(t, err.Error(), "auth provider name 'Google Client' must only include alphanumeric characters and underscores, and cannot start with a number\n")
	assert.Contains(t, err.Error(), "auth provider name 'google/' must only include alphanumeric characters and underscores, and cannot start with a number\n")
	assert.Contains(t, err.Error(), "auth provider name 'google\\client' must only include alphanumeric characters and underscores, and cannot start with a number\n")
	assert.Contains(t, err.Error(), "auth provider name '1google' must only include alphanumeric characters and underscores, and cannot start with a number\n")
	assert.Contains(t, err.Error(), "auth provider name 'Google-Client' must only include alphanumeric characters and underscores, and cannot start with a number\n")
}

func TestMissingProviderName(t *testing.T) {
	_, err := Load("fixtures/test_auth_missing_names.yaml")

	assert.Contains(t, err.Error(), "auth provider at index 0 is missing field: name\n")
	assert.Contains(t, err.Error(), "auth provider at index 1 is missing field: name\n")
	assert.Contains(t, err.Error(), "auth provider at index 2 is missing field: name\n")
}

func TestDuplicateProviderName(t *testing.T) {
	_, err := Load("fixtures/test_auth_duplicate_names.yaml")

	assert.Equal(t, "auth provider name 'my_google' has been defined more than once, but must be unique\n", err.Error())
}

func TestInvalidProviderTypes(t *testing.T) {
	_, err := Load("fixtures/test_auth_invalid_types.yaml")

	assert.Contains(t, err.Error(), "auth provider 'google_1' has invalid type 'google_1' which must be one of: google, facebook, gitlab, slack, oidc\n")
	assert.Contains(t, err.Error(), "auth provider 'google_2' has invalid type 'Google' which must be one of: google, facebook, gitlab, slack, oidc\n")
	assert.Contains(t, err.Error(), "auth provider 'Github' has invalid type 'oauth' which must be one of: google, facebook, gitlab, slack, oidc\n")
	assert.Contains(t, err.Error(), "auth provider 'Baidu' has invalid type 'whoops' which must be one of: google, facebook, gitlab, slack, oidc\n")
}

func TestMissingClientId(t *testing.T) {
	_, err := Load("fixtures/test_auth_missing_client_ids.yaml")

	assert.Contains(t, err.Error(), "auth provider 'google_1' is missing field: clientId\n")
	assert.Contains(t, err.Error(), "auth provider 'Baidu' is missing field: clientId\n")
	assert.Contains(t, err.Error(), "auth provider 'Github' is missing field: clientId\n")
}

func TestMissingOrInvalidIssuerUrl(t *testing.T) {
	_, err := Load("fixtures/test_auth_invalid_issuer.yaml")

	assert.Contains(t, err.Error(), "auth provider 'not-https' has missing or invalid https url for field: issuerUrl\n")
	assert.Contains(t, err.Error(), "auth provider 'missing-issuer' has missing or invalid https url for field: issuerUrl\n")
	assert.Contains(t, err.Error(), "auth provider 'no-schema' has missing or invalid https url for field: issuerUrl\n")
	assert.Contains(t, err.Error(), "auth provider 'invalid-url' has missing or invalid https url for field: issuerUrl\n")
}

func TestMissingOrInvalidTokenEndpoint(t *testing.T) {
	_, err := Load("fixtures/test_auth_invalid_token_url.yaml")

	assert.Contains(t, err.Error(), "auth provider 'not-https' has missing or invalid https url for field: tokenUrl\n")
	assert.Contains(t, err.Error(), "auth provider 'missing-schema' has missing or invalid https url for field: tokenUrl\n")
	assert.Contains(t, err.Error(), "auth provider 'missing-endpoint' has missing or invalid https url for field: tokenUrl\n")
}

func TestGetOidcIssuer(t *testing.T) {
	config, err := Load("fixtures/test_auth.yaml")
	assert.NoError(t, err)

	googleIssuer, err := config.Auth.GetOidcProvidersByIssuer("https://accounts.google.com/")
	assert.NoError(t, err)
	assert.Len(t, googleIssuer, 2)

	auth0Issuer, err := config.Auth.GetOidcProvidersByIssuer("https://dev-skhlutl45lbqkvhv.us.auth0.com")
	assert.NoError(t, err)
	assert.Len(t, auth0Issuer, 1)

	nopeIssuer, err := config.Auth.GetOidcProvidersByIssuer("https://nope.com")
	assert.NoError(t, err)
	assert.Len(t, nopeIssuer, 0)
}

func TestGetOidcSameIssuers(t *testing.T) {
	config, err := Load("fixtures/test_auth_same_issuers.yaml")
	assert.NoError(t, err)

	googleIssuer, err := config.Auth.GetOidcProvidersByIssuer("https://accounts.google.com/")
	assert.NoError(t, err)
	assert.Len(t, googleIssuer, 3)
}

func TestAddOidcProvider(t *testing.T) {
	config, err := Load("fixtures/test_auth.yaml")
	assert.NoError(t, err)

	assert.Len(t, config.Auth.GetOidcProviders(), 1)

	err = config.Auth.AddOidcProvider("CustomAuth", "https://mycustomoidc.com", "1234")
	assert.NoError(t, err)

	assert.Len(t, config.Auth.GetOidcProviders(), 2)

	byIssuer, err := config.Auth.GetOidcProvidersByIssuer("https://mycustomoidc.com")
	assert.NoError(t, err)
	assert.Len(t, byIssuer, 1)
}

func TestAddOidcProviderReservedPrefix(t *testing.T) {
	_, err := Load("fixtures/test_auth_reserved_prefix.yaml")
	assert.ErrorContains(t, err, "cannot use reserved 'keel_' prefix in auth provider name: keel_client")
	assert.ErrorContains(t, err, "cannot use reserved 'keel_' prefix in auth provider name: KEEL_CLIENT")
}

func TestAddOidcProviderInvalidName(t *testing.T) {
	auth := &AuthConfig{}
	err := auth.AddOidcProvider("my client", "https://mycustomoidc.com", "1234")
	assert.ErrorContains(t, err, "auth provider name 'my client' must only include alphanumeric characters and underscores, and cannot start with a number")
}

func TestAddOidcProviderAlreadyExists(t *testing.T) {
	auth := &AuthConfig{}
	err := auth.AddOidcProvider("my_client", "https://mycustomoidc.com", "1234")
	assert.NoError(t, err)
	err = auth.AddOidcProvider("my_client", "https://anothercustomoidc.com", "abcd")
	assert.ErrorContains(t, err, "auth provider name 'my_client' has been defined more than once, but must be unique")
}

func TestGetCallbackUrl_Localhost(t *testing.T) {
	provider := &Provider{
		Name: "google",
	}
	t.Setenv("KEEL_API_URL", "http://localhost:8000")

	url, err := provider.GetCallbackUrl()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8000/auth/callback/google", url.String())
}

func TestGetCallbackUrl_Web(t *testing.T) {
	provider := &Provider{
		Name: "google",
	}
	t.Setenv("KEEL_API_URL", "https://myapplication.com/keel/")

	url, err := provider.GetCallbackUrl()
	assert.NoError(t, err)
	assert.Equal(t, "https://myapplication.com/keel/auth/callback/google", url.String())
}

func TestGetCallbackUrl_WithUnderscoredAndCapitals(t *testing.T) {
	provider := &Provider{
		Name: "GOOGLE_Client_1",
	}
	t.Setenv("KEEL_API_URL", "http://localhost:8000")

	url, err := provider.GetCallbackUrl()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8000/auth/callback/google_client_1", url.String())
}

func TestGetCallbackUrl_NoKeelApiUrl(t *testing.T) {
	provider := &Provider{
		Name: "google",
	}

	url, err := provider.GetCallbackUrl()
	assert.ErrorContains(t, err, "empty url")
	assert.Nil(t, url)
}

func TestAuthClaims(t *testing.T) {
	config, err := Load("fixtures/test_auth_identity_claims.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "https://slack.com/#team_ID", config.Auth.Claims[0].Key)
	assert.Equal(t, "teamId", config.Auth.Claims[0].Field)
	assert.Equal(t, true, config.Auth.Claims[0].Unique)

	assert.Equal(t, "something-else", config.Auth.Claims[1].Key)
	assert.Equal(t, "somethingElse", config.Auth.Claims[1].Field)
	assert.Equal(t, false, config.Auth.Claims[1].Unique)
}
