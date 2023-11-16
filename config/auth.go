package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/samber/lo"
)

const (
	// 24 hours is the default access token expiry period
	DefaultAccessTokenExpiry time.Duration = time.Hour * 24
	// 3 months is the default refresh token expiry period
	DefaultRefreshTokenExpiry time.Duration = time.Hour * 24 * 90
)

const (
	GoogleProvider        = "google"
	FacebookProvider      = "facebook"
	GitLabProvider        = "gitlab"
	OpenIdConnectProvider = "oidc"
	OAuthProvider         = "oauth"
)

var (
	SupportedProviderTypes = []string{
		GoogleProvider,
		FacebookProvider,
		GitLabProvider,
		OpenIdConnectProvider,
		OAuthProvider,
	}
)

type AuthConfig struct {
	Tokens    TokensConfig `yaml:"tokens"`
	Providers []Provider   `yaml:"providers"`
}

type TokensConfig struct {
	AccessTokenExpiry           *int  `yaml:"accessTokenExpiry,omitempty"`
	RefreshTokenExpiry          *int  `yaml:"refreshTokenExpiry,omitempty"`
	RefreshTokenRotationEnabled *bool `yaml:"refreshTokenRotationEnabled,omitempty"`
}

type Provider struct {
	Type             string `yaml:"type"`
	Name             string `yaml:"name"`
	ClientId         string `yaml:"clientId"`
	IssuerUrl        string `yaml:"issuerUrl"`
	TokenUrl         string `yaml:"tokenUrl"`
	AuthorizationUrl string `yaml:"authorizationUrl"`
}

// AccessTokenExpiry retrieves the configured or default access token expiry
func (c *AuthConfig) AccessTokenExpiry() time.Duration {
	if c.Tokens.AccessTokenExpiry != nil {
		return time.Duration(*c.Tokens.AccessTokenExpiry) * time.Second
	} else {
		return DefaultAccessTokenExpiry
	}
}

// RefreshTokenExpiry retrieves the configured or default refresh token expiry
func (c *AuthConfig) RefreshTokenExpiry() time.Duration {
	if c.Tokens.RefreshTokenExpiry != nil {
		return time.Duration(*c.Tokens.RefreshTokenExpiry) * time.Second
	} else {
		return DefaultRefreshTokenExpiry
	}
}

// RefreshTokenRotationEnabled retrieves the configured or default refresh token rotation
func (c *AuthConfig) RefreshTokenRotationEnabled() bool {
	if c.Tokens.RefreshTokenRotationEnabled != nil {
		return *c.Tokens.RefreshTokenRotationEnabled
	} else {
		return true
	}
}

// AddOidcProvider adds an OpenID Connect provider to the list of supported authentication providers
func (c *AuthConfig) AddOidcProvider(name string, issuerUrl string, clientId string) error {
	if name == "" {
		return errors.New("provider name cannot be empty")
	}
	if invalidUrl(issuerUrl) {
		return fmt.Errorf("invalid issuerUrl: %s", issuerUrl)
	}
	if clientId == "" {
		return errors.New("provider clientId cannot be empty")
	}

	provider := Provider{
		Type:      OpenIdConnectProvider,
		Name:      name,
		IssuerUrl: issuerUrl,
		ClientId:  clientId,
	}

	c.Providers = append(c.Providers, provider)
	return nil
}

// GetOidcProviders returns all OpenID Connect compatible authentication providers
func (c *AuthConfig) GetOidcProviders() []Provider {
	oidcProviders := []Provider{}
	for _, p := range c.Providers {
		if p.Type == OpenIdConnectProvider {
			oidcProviders = append(oidcProviders, p)
		}
	}
	return oidcProviders
}

// GetOidcProvidersByIssuer gets all OpenID Connect providers by issuer url.
// It's possible that multiple providers from the same issuer are configured.
func (c *AuthConfig) GetOidcProvidersByIssuer(issuer string) ([]Provider, error) {
	providers := []Provider{}

	for _, p := range c.Providers {
		if p.Type == OAuthProvider {
			continue
		}

		issuerUrl, err := p.GetIssuer()
		if err != nil {
			return nil, err
		}
		if strings.TrimSuffix(issuerUrl, "/") == strings.TrimSuffix(issuer, "/") {
			providers = append(providers, p)
		}
	}

	return providers, nil
}

// GetIssuer retrieves the issuer URL for the provider
func (c *Provider) GetIssuer() (string, error) {
	switch c.Type {
	case GoogleProvider:
		return "https://accounts.google.com", nil
	case FacebookProvider:
		return "https://www.facebook.com", nil
	case GitLabProvider:
		return "https://gitlab.com", nil
	case OpenIdConnectProvider:
		return c.IssuerUrl, nil
	default:
		return "", fmt.Errorf("the provider type '%s' should not have an issuer url configured", c.Type)
	}
}

func (c *AuthConfig) GetOAuthProviders() []Provider {
	oidcProviders := []Provider{}
	for _, p := range c.Providers {
		if p.Type == OAuthProvider {
			oidcProviders = append(oidcProviders, p)
		}
	}
	return oidcProviders
}

func (c *Provider) GetTokenUrl() (string, error) {
	switch c.Type {
	case GoogleProvider:
		return "https://accounts.google.com/o/oauth2/token", nil
	case OAuthProvider:
		return c.TokenUrl, nil
	default:
		return "", fmt.Errorf("the provider type '%s' should not have a token url configured", c.Type)
	}
}

func (c *Provider) GetAuthorizationUrl() (string, error) {
	switch c.Type {
	case GoogleProvider:
		return "https://accounts.google.com/o/oauth2/auth", nil
	case OAuthProvider:
		return c.AuthorizationUrl, nil
	default:
		return "", fmt.Errorf("the provider type '%s' should not have a token url configured", c.Type)
	}
}

// findAuthProviderMissingName checks for missing provider names
func findAuthProviderMissingName(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if p.Name == "" {
			invalid = append(invalid, p)
		}
	}

	return invalid
}

// findAuthProviderDuplicateName checks for duplicate auth provider names
func findAuthProviderDuplicateName(providers []Provider) []Provider {
	keys := make(map[string]bool)

	duplicates := []Provider{}
	for _, p := range providers {
		if _, value := keys[p.Name]; !value {
			keys[p.Name] = true
		} else {
			duplicates = append(duplicates, p)
		}
	}

	return duplicates
}

// findAuthProviderInvalidType checks for invalid provider types
func findAuthProviderInvalidType(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if !lo.Contains(SupportedProviderTypes, p.Type) {
			invalid = append(invalid, p)
		}
	}

	return invalid
}

// findAuthProviderMissingClientId checks for missing client IDs
func findAuthProviderMissingClientId(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if p.ClientId == "" {
			invalid = append(invalid, p)
		}
	}

	return invalid
}

// findAuthProviderMissingIssuerUrl checks for missing or invalid issuer URLs
func findAuthProviderMissingOrInvalidIssuerUrl(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if invalidUrl(p.IssuerUrl) {
			invalid = append(invalid, p)
			continue
		}
	}

	return invalid
}

// findAuthProviderMissingOrInvalidTokenUrl checks for missing or invalid token URLs
func findAuthProviderMissingOrInvalidTokenUrl(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if invalidUrl(p.TokenUrl) {
			invalid = append(invalid, p)
			continue
		}
	}

	return invalid
}

// findAuthProviderMissingOrInvalidAuthorizationUrl checks for missing or invalid authorization URLs
func findAuthProviderMissingOrInvalidAuthorizationUrl(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if invalidUrl(p.AuthorizationUrl) {
			invalid = append(invalid, p)
			continue
		}
	}
	return invalid
}

func invalidUrl(u string) bool {
	parsed, err := url.Parse(u)
	return err != nil || parsed.Scheme != "https"
}
