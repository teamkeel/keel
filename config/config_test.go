package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	config, err := Load("fixtures/test_basic_config.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "TEST", config.Environment.Default[0].Name)
	assert.Equal(t, "test", config.Environment.Default[0].Value)
	assert.Equal(t, "API_KEY", config.Secrets[0].Name)
}

func TestFailConfigValue(t *testing.T) {
	_, err := Load("fixtures/test_failing_config.yaml")
	assert.Error(t, err)

	assert.Equal(t, "could not unmarshal config file: yaml: unmarshal errors:\n  line 5: cannot unmarshal !!seq into string", err.Error())
}

func TestDuplicates(t *testing.T) {
	config, err := Load("fixtures/test_duplicates.yaml")
	assert.Error(t, err)

	assert.Equal(t, "TEST", config.Environment.Default[0].Name)
	assert.Equal(t, "environment variable TEST has a duplicate set in environment: [staging]\n", err.Error())
}

func TestRequiredFail(t *testing.T) {
	config, err := Load("fixtures/test_required_fail.yaml")
	assert.Error(t, err)

	assert.Equal(t, "TEST", config.Environment.Staging[0].Name)
	assert.Equal(t, "environment variable TEST is required but not defined in the following environments: [production]\n", err.Error())
}

func TestRequired(t *testing.T) {
	config, err := Load("fixtures/test_required.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "TEST", config.Environment.Staging[0].Name)
	assert.Equal(t, "TEST", config.Environment.Production[0].Name)
}

func TestRequiredValuesKeys(t *testing.T) {
	config, err := Load("fixtures/test_required.yaml")
	assert.NoError(t, err)

	notOk, keys := requiredValuesKeys(config)
	assert.False(t, notOk)

	var nil []string
	assert.Equal(t, nil, keys["staging"])

	config, err = Load("fixtures/test_required_fail.yaml")
	assert.Error(t, err)

	notOk, keys = requiredValuesKeys(config)
	assert.True(t, notOk)
	assert.Equal(t, []string{"production"}, keys["TEST"])
}

func TestEmptyConfig(t *testing.T) {
	config, err := Load("fixtures/test_required_fail_empty.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "TEST", config.Environment.Staging[0].Name)
	assert.Equal(t, "test2_duplicate", config.Environment.Staging[0].Value)
}

func TestAllEnvironmentVariables(t *testing.T) {
	config, err := Load("fixtures/test_required.yaml")
	assert.NoError(t, err)

	allEnvVars := config.AllEnvironmentVariables()
	assert.Equal(t, []string{"TEST", "TEST_API_KEY"}, allEnvVars)
}

func TestGetEnvVars(t *testing.T) {
	config, err := Load("fixtures/test_merge_envs_config.yaml")
	assert.NoError(t, err)

	envVars := config.GetEnvVars("test")
	assert.Equal(t, map[string]string{"DOG_NAME": "Peggy", "PERSON_NAME": "Louisa"}, envVars)
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

func TestAuthTokens(t *testing.T) {
	config, err := Load("fixtures/test_auth.yaml")
	assert.NoError(t, err)

	assert.Equal(t, 3600, config.Auth.Tokens.AccessTokenExpiry)
	assert.Equal(t, 604800, config.Auth.Tokens.RefreshTokenExpiry)
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
	assert.Equal(t, "google-1", config.Auth.Providers[0].Name)
	assert.Equal(t, "foo_1", config.Auth.Providers[0].ClientId)

	assert.Equal(t, "google", config.Auth.Providers[1].Type)
	assert.Equal(t, "google_2", config.Auth.Providers[1].Name)
	assert.Equal(t, "foo_2", config.Auth.Providers[1].ClientId)

	assert.Equal(t, "oidc", config.Auth.Providers[2].Type)
	assert.Equal(t, "Baidu", config.Auth.Providers[2].Name)
	assert.Equal(t, "https://dev-skhlutl45lbqkvhv.us.auth0.com", config.Auth.Providers[2].IssuerUrl)
	assert.Equal(t, "kasj28fnq09ak", config.Auth.Providers[2].ClientId)

	assert.Equal(t, "oauth", config.Auth.Providers[3].Type)
	assert.Equal(t, "Github", config.Auth.Providers[3].Name)
	assert.Equal(t, "hfjuw983h1hfsdf", config.Auth.Providers[3].ClientId)
	assert.Equal(t, "https://github.com/auth", config.Auth.Providers[3].AuthorizationUrl)
	assert.Equal(t, "https://github.com/token", config.Auth.Providers[3].TokenUrl)
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

	assert.Contains(t, err.Error(), "auth provider 'google_1' has invalid type 'google_1' which must be one of: google, oidc, oauth\n")
	assert.Contains(t, err.Error(), "auth provider 'google_2' has invalid type 'Google' which must be one of: google, oidc, oauth\n")
	assert.Contains(t, err.Error(), "auth provider 'Baidu' has invalid type 'whoops' which must be one of: google, oidc, oauth\n")
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

func TestHasIssuer(t *testing.T) {
	config, err := Load("fixtures/test_auth.yaml")
	assert.NoError(t, err)

	hasGoogleIssuer, err := config.Auth.HasOidcIssuer("https://accounts.google.com/")
	assert.NoError(t, err)
	assert.True(t, hasGoogleIssuer)

	hasCustomIssuer, err := config.Auth.HasOidcIssuer("https://dev-skhlutl45lbqkvhv.us.auth0.com")
	assert.NoError(t, err)
	assert.True(t, hasCustomIssuer)

	hasUnknownIssuer, err := config.Auth.HasOidcIssuer("https://nope.com")
	assert.NoError(t, err)
	assert.False(t, hasUnknownIssuer)
}
