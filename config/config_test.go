package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
)

func TestValidation(t *testing.T) {
	entries, err := os.ReadDir("./fixtures")
	assert.NoError(t, err)

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("./fixtures", entry.Name()))
			assert.NoError(t, err)

			_, err = config.Load(filepath.Join("./fixtures", entry.Name()))
			if err != nil && config.ToConfigErrors(err) == nil {
				require.NoError(t, err)
			}

			configErrors := config.ToConfigErrors(err)
			expectedErrors := []string{}
			actualErrors := []string{}

			for _, line := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(line, "# ") {
					expectedErrors = append(expectedErrors, strings.TrimPrefix(line, "# "))
					continue
				}
				break
			}

			if configErrors != nil {
				for _, err := range configErrors.Errors {
					actualErrors = append(actualErrors, err.Message)
				}
			}

			unexpected, expected := lo.Difference(actualErrors, expectedErrors)

			for _, v := range unexpected {
				t.Errorf("Unexpected error: %s", v)
			}

			for _, v := range expected {
				t.Errorf("Expected error: %s", v)
			}
		})
	}
}

func TestAuthDefaults(t *testing.T) {
	t.Parallel()
	config := &config.ProjectConfig{}

	assert.Nil(t, config.Auth.Tokens.AccessTokenExpiry)
	assert.Nil(t, config.Auth.Tokens.RefreshTokenExpiry)
	assert.Nil(t, config.Auth.Tokens.RefreshTokenRotationEnabled)

	assert.Equal(t, time.Duration(24)*time.Hour, config.Auth.AccessTokenExpiry())
	assert.Equal(t, time.Duration(24)*time.Hour*90, config.Auth.RefreshTokenExpiry())
	assert.Equal(t, true, config.Auth.RefreshTokenRotationEnabled())
}

func TestGetOidcIssuer(t *testing.T) {
	t.Parallel()
	config, err := config.Load("fixtures/test_auth.yaml")
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
	t.Parallel()
	config, err := config.Load("fixtures/test_auth_same_issuers.yaml")
	assert.NoError(t, err)

	googleIssuer, err := config.Auth.GetOidcProvidersByIssuer("https://accounts.google.com/")
	assert.NoError(t, err)
	assert.Len(t, googleIssuer, 3)
}

func TestAddOidcProvider(t *testing.T) {
	t.Parallel()
	config, err := config.Load("fixtures/test_auth.yaml")
	assert.NoError(t, err)

	assert.Len(t, config.Auth.GetOidcProviders(), 1)

	err = config.Auth.AddOidcProvider("CustomAuth", "https://mycustomoidc.com", "1234")
	assert.NoError(t, err)

	assert.Len(t, config.Auth.GetOidcProviders(), 2)

	byIssuer, err := config.Auth.GetOidcProvidersByIssuer("https://mycustomoidc.com")
	assert.NoError(t, err)
	assert.Len(t, byIssuer, 1)
}

func TestAddOidcProviderInvalidName(t *testing.T) {
	t.Parallel()
	auth := &config.AuthConfig{}
	err := auth.AddOidcProvider("my client", "https://mycustomoidc.com", "1234")
	assert.ErrorContains(t, err, "auth.providers.0.name: Does not match pattern '^[a-zA-Z][a-zA-Z0-9_]+$'")
}

func TestAddOidcProviderKeelAuth(t *testing.T) {
	t.Parallel()
	auth := &config.AuthConfig{}
	err := auth.AddOidcProvider("keel_auth", "https://auth.keel.xyz/", "1234")
	assert.NoError(t, err)
}

func TestAddOidcProviderAlreadyExists(t *testing.T) {
	t.Parallel()
	auth := &config.AuthConfig{}
	err := auth.AddOidcProvider("my_client", "https://mycustomoidc.com", "1234")
	assert.NoError(t, err)
	err = auth.AddOidcProvider("my_client", "https://anothercustomoidc.com", "abcd")
	assert.ErrorContains(t, err, "auth.providers.1.name: Duplicate name my_client")
}

func TestGetCallbackUrl_Localhost(t *testing.T) {
	provider := &config.Provider{
		Name: "google",
	}
	t.Setenv("KEEL_API_URL", "http://localhost:8000")

	url, err := provider.GetCallbackUrl()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8000/auth/callback/google", url.String())
}

func TestGetCallbackUrl_Web(t *testing.T) {
	provider := &config.Provider{
		Name: "google",
	}
	t.Setenv("KEEL_API_URL", "https://myapplication.com/keel/")

	url, err := provider.GetCallbackUrl()
	assert.NoError(t, err)
	assert.Equal(t, "https://myapplication.com/keel/auth/callback/google", url.String())
}

func TestGetCallbackUrl_WithUnderscoredAndCapitals(t *testing.T) {
	provider := &config.Provider{
		Name: "GOOGLE_Client_1",
	}
	t.Setenv("KEEL_API_URL", "http://localhost:8000")

	url, err := provider.GetCallbackUrl()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8000/auth/callback/google_client_1", url.String())
}

func TestGetCallbackUrl_NoKeelApiUrl(t *testing.T) {
	t.Parallel()
	provider := &config.Provider{
		Name: "google",
	}

	url, err := provider.GetCallbackUrl()
	assert.ErrorContains(t, err, "empty url")
	assert.Nil(t, url)
}
