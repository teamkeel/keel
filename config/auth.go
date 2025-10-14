package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// 24 hours is the default access token expiry period.
	DefaultAccessTokenExpiry time.Duration = time.Hour * 24
	// 3 months is the default refresh token expiry period.
	DefaultRefreshTokenExpiry time.Duration = time.Hour * 24 * 90
)

const ProviderSecretPrefix = "AUTH_PROVIDER_SECRET_"

const ReservedProviderNamePrefix = "keel_"

const (
	GoogleProvider        = "google"
	FacebookProvider      = "facebook"
	GitLabProvider        = "gitlab"
	SlackProvider         = "slack"
	OpenIdConnectProvider = "oidc"
)

var (
	SupportedProviderTypes = []string{
		GoogleProvider,
		FacebookProvider,
		GitLabProvider,
		SlackProvider,
		OpenIdConnectProvider,
	}
)

type FunctionHook string

const (
	HookAfterAuthentication  FunctionHook = "afterAuthentication"
	HookAfterIdentityCreated FunctionHook = "afterIdentityCreated"
)

type AuthConfig struct {
	Tokens      TokensConfig    `yaml:"tokens"`
	RedirectUrl *string         `yaml:"redirectUrl,omitempty"`
	Providers   []Provider      `yaml:"providers"`
	Claims      []IdentityClaim `yaml:"claims"`
	Hooks       []FunctionHook  `yaml:"hooks"`
}

type TokensConfig struct {
	AccessTokenExpiry           *int  `yaml:"accessTokenExpiry,omitempty"`
	RefreshTokenExpiry          *int  `yaml:"refreshTokenExpiry,omitempty"`
	RefreshTokenRotationEnabled *bool `yaml:"refreshTokenRotationEnabled,omitempty"`
}

type Provider struct {
	Type      string   `yaml:"type"`
	Name      string   `yaml:"name"`
	ClientId  string   `yaml:"clientId"`
	IssuerUrl string   `yaml:"issuerUrl,omitempty"`
	Scopes    []string `yaml:"scopes,omitempty"`
}

type IdentityClaim struct {
	Key    string `yaml:"key"`
	Field  string `yaml:"field"`
	Unique bool   `yaml:"unique"`
}

// AccessTokenExpiry retrieves the configured or default access token expiry.
func (c *AuthConfig) AccessTokenExpiry() time.Duration {
	if c.Tokens.AccessTokenExpiry != nil {
		return time.Duration(*c.Tokens.AccessTokenExpiry) * time.Second
	} else {
		return DefaultAccessTokenExpiry
	}
}

func (c *AuthConfig) EnabledHooks() []FunctionHook {
	return c.Hooks
}

// RefreshTokenExpiry retrieves the configured or default refresh token expiry.
func (c *AuthConfig) RefreshTokenExpiry() time.Duration {
	if c.Tokens.RefreshTokenExpiry != nil {
		return time.Duration(*c.Tokens.RefreshTokenExpiry) * time.Second
	} else {
		return DefaultRefreshTokenExpiry
	}
}

// RefreshTokenRotationEnabled retrieves the configured or default refresh token rotation.
func (c *AuthConfig) RefreshTokenRotationEnabled() bool {
	if c.Tokens.RefreshTokenRotationEnabled != nil {
		return *c.Tokens.RefreshTokenRotationEnabled
	} else {
		return true
	}
}

// AddOidcProvider adds an OpenID Connect provider to the list of supported authentication providers.
func (c *AuthConfig) AddOidcProvider(name string, issuerUrl string, clientId string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if issuerUrl == "" {
		return fmt.Errorf("issuerUrl is required")
	}
	if clientId == "" {
		return fmt.Errorf("clientId is required")
	}

	provider := Provider{
		Type:      OpenIdConnectProvider,
		Name:      name,
		IssuerUrl: issuerUrl,
		ClientId:  clientId,
	}

	c.Providers = append(c.Providers, provider)

	b, err := yaml.Marshal(&ProjectConfig{Auth: *c})
	if err != nil {
		return err
	}

	_, err = LoadFromBytes(b, "")
	if err != nil && ToConfigErrors(err) == nil {
		return err
	}

	if err == nil {
		return nil
	}

	newProviderIndex := len(c.Providers) - 1
	for _, err := range ToConfigErrors(err).Errors {
		if strings.HasPrefix(err.Message, fmt.Sprintf("auth.providers.%d.name", newProviderIndex)) {
			// This function allows the adding of internal auth providers which can start with 'keel_'
			if !strings.Contains(err.Message, "Cannot start with 'keel_'") {
				return err
			}
		}
		if strings.HasPrefix(err.Message, fmt.Sprintf("auth.providers.%d.issuerUrl", newProviderIndex)) {
			return err
		}
		if strings.HasPrefix(err.Message, fmt.Sprintf("auth.providers.%d.clientId", newProviderIndex)) {
			return err
		}
	}

	return nil
}

// GetOidcProviders returns all OpenID Connect compatible authentication providers.
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
		issuerUrl, hasIssuer := p.GetIssuerUrl()
		if !hasIssuer {
			return nil, fmt.Errorf("issuer url has not been configured: %s", issuer)
		}

		if strings.TrimSuffix(issuerUrl, "/") == strings.TrimSuffix(issuer, "/") {
			providers = append(providers, p)
		}
	}

	return providers, nil
}

// GetClientSecret generates the name of the client secret.
func (p *Provider) GetClientSecretName() string {
	return fmt.Sprintf("%s%s", ProviderSecretPrefix, strings.ToUpper(p.Name))
}

// GetAuthorizeUrl retrieves the authorize URL for this provider.
func (p *Provider) GetAuthorizeUrl() (*url.URL, error) {
	apiUrl, err := url.ParseRequestURI(os.Getenv("KEEL_API_URL"))
	if err != nil {
		return nil, err
	}

	return apiUrl.JoinPath("/auth/authorize/" + strings.ToLower(p.Name)), nil
}

// GetCallbackUrl retrieves the callback URL for this provider.
func (p *Provider) GetCallbackUrl() (*url.URL, error) {
	apiUrl, err := url.ParseRequestURI(os.Getenv("KEEL_API_URL"))
	if err != nil {
		return nil, err
	}
	return apiUrl.JoinPath("/auth/callback/" + strings.ToLower(p.Name)), nil
}

// GetProvider retrieves the provider by its name (case insensitive).
func (c *AuthConfig) GetProvider(name string) *Provider {
	for _, p := range c.Providers {
		if strings.EqualFold(p.Name, name) {
			return &p
		}
	}
	return nil
}

// GetIssuerUrl retrieves the issuer URL for the provider.
func (p *Provider) GetIssuerUrl() (string, bool) {
	switch p.Type {
	case GoogleProvider:
		return "https://accounts.google.com", true
	case FacebookProvider:
		return "https://www.facebook.com", true
	case GitLabProvider:
		return "https://gitlab.com", true
	case SlackProvider:
		return "https://slack.com", true
	case OpenIdConnectProvider:
		return p.IssuerUrl, true
	default:
		return "", false
	}
}
