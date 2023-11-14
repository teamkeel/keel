package config

import (
	"fmt"
	"net/url"

	"github.com/samber/lo"
)

const (
	GoogleProvider        = "google"
	OpenIdConnectProvider = "oidc"
	OAuthProvider         = "oauth"
)

var (
	SupportedProviderTypes = []string{
		GoogleProvider,
		OpenIdConnectProvider,
		OAuthProvider,
	}
)

type AuthConfig struct {
	Tokens    *TokensConfig `yaml:"tokens"`
	Providers []Provider    `yaml:"providers"`
}

type TokensConfig struct {
	AccessTokenExpiry  int `yaml:"accessTokenExpiry"`
	RefreshTokenExpiry int `yaml:"refreshTokenExpiry"`
}

type Provider struct {
	Type             string `yaml:"type"`
	Name             string `yaml:"name"`
	ClientId         string `yaml:"clientId"`
	IssuerUrl        string `yaml:"issuerUrl"`
	TokenUrl         string `yaml:"tokenUrl"`
	AuthorizationUrl string `yaml:"authorizationUrl"`
}

func (c *AuthConfig) GetOidcProviders() []Provider {
	oidcProviders := []Provider{}
	for _, p := range c.Providers {
		if p.Type == OpenIdConnectProvider {
			oidcProviders = append(oidcProviders, p)
		}
	}
	return oidcProviders
}

// GetProvidersOidcIssuer gets all providers by issuer url.
// It's possible that multiple providers from the same issuer as configured.
func (c *AuthConfig) GetProvidersOidcIssuer(issuer string) ([]Provider, error) {
	providers := []Provider{}

	for _, p := range c.Providers {
		if p.Type == OAuthProvider {
			continue
		}

		issuerUrl, err := p.GetIssuer()
		if err != nil {
			return nil, err
		}
		if issuerUrl == issuer {
			providers = append(providers, p)
		}
	}

	return providers, nil
}

func (c *Provider) GetIssuer() (string, error) {
	switch c.Type {
	case GoogleProvider:
		return "https://accounts.google.com/", nil
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
		u, err := url.Parse(p.IssuerUrl)
		if err != nil || u.Scheme != "https" {
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
		u, err := url.Parse(p.TokenUrl)
		if err != nil || u.Scheme != "https" {
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
		u, err := url.Parse(p.AuthorizationUrl)
		if err != nil || u.Scheme != "https" {
			invalid = append(invalid, p)
			continue
		}
	}

	return invalid
}
