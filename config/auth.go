package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
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

const ProviderSecretPrefix = "AUTH_PROVIDER_SECRET_"

const ReservedProviderNamePrefix = "keel_"

const (
	GoogleProvider        = "google"
	FacebookProvider      = "facebook"
	GitLabProvider        = "gitlab"
	SlackProvider         = "slack"
	OpenIdConnectProvider = "oidc"
	OAuthProvider         = "oauth"
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

type AuthConfig struct {
	Tokens      TokensConfig    `yaml:"tokens"`
	RedirectUrl *string         `yaml:"redirectUrl,omitempty"`
	Providers   []Provider      `yaml:"providers"`
	Claims      []IdentityClaim `yaml:"claims"`
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

type IdentityClaim struct {
	Key   string `yaml:"key"`
	Field string `yaml:"field"`
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
	if invalidName(name) {
		return fmt.Errorf(ConfigAuthProviderInvalidName, name)
	}
	for _, v := range c.Providers {
		if v.Name == name {
			return fmt.Errorf(ConfigAuthProviderDuplicateErrorString, name)
		}
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

// GetClientSecret generates the name of the client secret
func (p *Provider) GetClientSecretName() string {
	return fmt.Sprintf("%s%s", ProviderSecretPrefix, strings.ToUpper(p.Name))
}

// GetAuthorizeUrl retrieves the authorize URL for this provider
func (p *Provider) GetAuthorizeUrl() (*url.URL, error) {
	apiUrl, err := url.ParseRequestURI(os.Getenv("KEEL_API_URL"))
	if err != nil {
		return nil, err
	}

	return apiUrl.JoinPath("/auth/authorize/" + strings.ToLower(p.Name)), nil
}

// GetCallbackUrl retrieves the callback URL for this provider
func (p *Provider) GetCallbackUrl() (*url.URL, error) {
	apiUrl, err := url.ParseRequestURI(os.Getenv("KEEL_API_URL"))
	if err != nil {
		return nil, err
	}
	return apiUrl.JoinPath("/auth/callback/" + strings.ToLower(p.Name)), nil
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

// GetProvider retrieves the provider by its name (case insensitive)
func (c *AuthConfig) GetProvider(name string) *Provider {
	for _, p := range c.Providers {
		if strings.EqualFold(p.Name, name) {
			return &p
		}
	}
	return nil
}

// GetIssuerUrl retrieves the issuer URL for the provider
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

func (p *Provider) GetTokenUrl() (string, bool) {
	switch p.Type {
	case GoogleProvider:
		return "https://oauth2.googleapis.com/token", true
	case FacebookProvider:
		return "https://graph.facebook.com/v11.0/oauth/access_token", true
	case GitLabProvider:
		return "https://gitlab.com/oauth/token", true
	case SlackProvider:
		return "https://slack.com/api/openid.connect.token", true
	case OpenIdConnectProvider:
		return p.TokenUrl, true
	case OAuthProvider:
		return p.TokenUrl, true
	default:
		return "", false
	}
}

func (p *Provider) GetAuthorizationUrl() (string, bool) {
	switch p.Type {
	case GoogleProvider:
		return "https://accounts.google.com/o/oauth2/auth", true
	case FacebookProvider:
		return "https://www.facebook.com/v11.0/dialog/oauth", true
	case GitLabProvider:
		return "https://gitlab.com/oauth/authorize", true
	case SlackProvider:
		return "https://slack.com/openid/connect/authorize", true
	case OpenIdConnectProvider:
		return p.AuthorizationUrl, true
	case OAuthProvider:
		return p.AuthorizationUrl, true
	default:
		return "", false
	}
}

// findAuthProviderInvalidName checks for invalid provider names
func findAuthProviderInvalidName(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if invalidName(p.Name) {
			invalid = append(invalid, p)
		}
	}

	return invalid
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

// findAuthProviderReservedName checks for reserved names
func findAuthProviderReservedName(providers []Provider) []Provider {
	invalid := []Provider{}
	for _, p := range providers {
		if strings.HasPrefix(strings.ToLower(p.Name), ReservedProviderNamePrefix) {
			invalid = append(invalid, p)
		}
	}

	return invalid
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

func invalidName(name string) bool {
	return !regexp.MustCompile(`^[A-Za-z_]\w*$`).MatchString(name)
}

func invalidUrl(u string) bool {
	parsed, err := url.Parse(u)
	return err != nil || parsed.Scheme != "https"
}
