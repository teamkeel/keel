package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/teamkeel/keel/casing"
	"gopkg.in/yaml.v3"
)

const Empty = ""

// ProjectConfig is the configuration for a keel project
type ProjectConfig struct {
	Environment   []Input    `yaml:"environment"`
	UseDefaultApi *bool      `yaml:"useDefaultApi,omitempty"`
	Secrets       []Input    `yaml:"secrets"`
	Auth          AuthConfig `yaml:"auth"`
	DisableAuth   bool       `yaml:"disableKeelAuth"`
}

func (p *ProjectConfig) GetEnvVars() map[string]string {
	nameToValueMap := map[string]string{}

	for _, input := range p.Environment {
		nameToValueMap[input.Name] = input.Value
	}

	return nameToValueMap
}

// AllEnvironmentVariables returns a slice of all of the unique environment variable key names
// defined across all environments
func (c *ProjectConfig) AllEnvironmentVariables() []string {
	var environmentVariables []string

	for _, envVar := range c.Environment {
		environmentVariables = append(environmentVariables, envVar.Name)
	}

	return environmentVariables
}

func (c *ProjectConfig) AllSecrets() []string {
	var secrets []string

	for _, secret := range c.Secrets {
		secrets = append(secrets, secret.Name)
	}

	return secrets
}

// DefaultApi provides the value of useDefaultApi from the config or a default value of true
// if no value is specified in the config
func (c *ProjectConfig) DefaultApi() bool {
	if c.UseDefaultApi == nil {
		return true
	} else {
		return *c.UseDefaultApi
	}
}

func (c *ProjectConfig) UsesAuthHook(hook HookFunction) bool {
	return slices.Contains(c.Auth.Hooks, hook)
}

// EnvironmentConfig is the configuration for a keel environment default, staging, production
type EnvironmentConfig struct {
	Default     []Input `yaml:"default"`
	Development []Input `yaml:"development"`
	Staging     []Input `yaml:"staging"`
	Production  []Input `yaml:"production"`
	Test        []Input `yaml:"test"`
}

// Input is the configuration for a keel environment variable or secret
type Input struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value,omitempty"`
}

type ConfigError struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}

const (
	ConfigDuplicateErrorString                       = "environment variable %s has a duplicate set"
	ConfigIncorrectNamingErrorString                 = "%s must be written in upper snakecase"
	ConfigReservedNameErrorString                    = "environment variable %s cannot start with %s as it is reserved"
	ConfigAuthTokenExpiryMustBePositive              = "%s token lifespan cannot be negative or zero for field: %s"
	ConfigAuthProviderInvalidName                    = "auth provider name '%s' must only include alphanumeric characters and underscores, and cannot start with a number"
	ConfigAuthProviderReservedPrefex                 = "cannot use reserved 'keel_' prefix in auth provider name: %s"
	ConfigAuthProviderMissingFieldAtIndexErrorString = "auth provider at index %v is missing field: %s"
	ConfigAuthProviderMissingFieldErrorString        = "auth provider '%s' is missing field: %s"
	ConfigAuthProviderInvalidTypeErrorString         = "auth provider '%s' has invalid type '%s' which must be one of: %s"
	ConfigAuthProviderDuplicateErrorString           = "auth provider name '%s' has been defined more than once, but must be unique"
	ConfigAuthProviderInvalidHttpUrlErrorString      = "auth provider '%s' has missing or invalid https url for field: %s"
	ConfigAuthInvalidRedirectUrlErrorString          = "auth redirectUrl '%s' is not a valid url"
	ConfigAuthInvalidHook                            = "%s is not a recognised hook"
)

type ConfigErrors struct {
	Errors []*ConfigError `json:"errors"`
}

func (c ConfigError) Error() string {
	return c.Message
}

func (c ConfigErrors) Error() string {
	str := ""

	for _, err := range c.Errors {
		str += fmt.Sprintf("%s\n", err.Message)
	}

	return str
}

func Load(dir string) (*ProjectConfig, error) {
	// If an absolute path to a file is provided then use it, otherwise append the default
	// file name
	if !strings.HasSuffix(dir, ".yaml") {
		dir = filepath.Join(dir, "keelconfig.yaml")
	}
	loadConfig, err := os.ReadFile(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ProjectConfig{}, nil
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	return parseAndValidate(loadConfig)
}

func LoadFromBytes(data []byte) (*ProjectConfig, error) {
	return parseAndValidate(data)
}

func parseAndValidate(data []byte) (*ProjectConfig, error) {
	var config ProjectConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, &ConfigErrors{
			Errors: []*ConfigError{
				{
					Type:    "parsing",
					Message: fmt.Sprintf("could not unmarshal config file: %s", err.Error()),
				},
			},
		}
	}

	configErrors := Validate(&config)
	if configErrors != nil {
		return &config, configErrors
	}

	return &config, nil
}

var reservedEnvVarRegex = regexp.MustCompile(`^AWS_|^_|^OTEL_|^OPENCOLLECTOR_CONFIG|^KEEL_`)

func Validate(config *ProjectConfig) *ConfigErrors {
	var errors []*ConfigError

	duplicates, results := checkForDuplicates(config)
	if duplicates {
		for duplicatedEnvVarName := range results {
			errors = append(errors, &ConfigError{
				Type:    "duplicate",
				Message: fmt.Sprintf(ConfigDuplicateErrorString, duplicatedEnvVarName),
			})
		}
	}

	hasIncorrectNames, incorrectNames := validateFormat(config, "snakecase")
	if hasIncorrectNames {
		for incorrectName := range incorrectNames {
			errors = append(errors, &ConfigError{
				Type:    "nonSnakecase",
				Message: fmt.Sprintf(ConfigIncorrectNamingErrorString, incorrectName),
			})
		}
	}

	hasIncorrectNames, incorrectNames = validateFormat(config, "reserved")
	if hasIncorrectNames {
		for incorrectName := range incorrectNames {
			startsWith := reservedEnvVarRegex.FindString(incorrectName)
			errors = append(errors, &ConfigError{
				Type:    "reserved",
				Message: fmt.Sprintf(ConfigReservedNameErrorString, incorrectName, startsWith),
			})
		}
	}

	if config.Auth.AccessTokenExpiry() <= 0 {
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthTokenExpiryMustBePositive, "access", "accessTokenExpiry"),
		})
	}

	if config.Auth.RefreshTokenExpiry() <= 0 {
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthTokenExpiryMustBePositive, "refresh", "refreshTokenExpiry"),
		})
	}

	invalidProviderNames := findAuthProviderInvalidName(config.Auth.Providers)
	for _, p := range invalidProviderNames {
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthProviderInvalidName, p.Name),
		})
	}

	missingProviderNames := findAuthProviderMissingName(config.Auth.Providers)
	for i := range missingProviderNames {
		errors = append(errors, &ConfigError{
			Type:    "missing",
			Message: fmt.Sprintf(ConfigAuthProviderMissingFieldAtIndexErrorString, i, "name"),
		})
	}

	invalidProviderTypes := findAuthProviderInvalidType(config.Auth.Providers)
	for _, p := range invalidProviderTypes {
		if p.Name == "" {
			continue
		}
		errors = append(errors, &ConfigError{
			Type:    "missing",
			Message: fmt.Sprintf(ConfigAuthProviderInvalidTypeErrorString, p.Name, p.Type, strings.Join(SupportedProviderTypes, ", ")),
		})
	}

	duplicateProviders := findAuthProviderDuplicateName(config.Auth.Providers)
	for _, p := range duplicateProviders {
		if p.Name == "" {
			continue
		}
		errors = append(errors, &ConfigError{
			Type:    "duplicate",
			Message: fmt.Sprintf(ConfigAuthProviderDuplicateErrorString, p.Name),
		})
	}

	reservedNames := findAuthProviderReservedName(config.Auth.Providers)
	for _, p := range reservedNames {
		if p.Name == "" {
			continue
		}
		errors = append(errors, &ConfigError{
			Type:    "reserved",
			Message: fmt.Sprintf(ConfigAuthProviderReservedPrefex, p.Name),
		})
	}

	missingClientIds := findAuthProviderMissingClientId(config.Auth.Providers)
	for _, p := range missingClientIds {
		if p.Name == "" {
			continue
		}
		errors = append(errors, &ConfigError{
			Type:    "missing",
			Message: fmt.Sprintf(ConfigAuthProviderMissingFieldErrorString, p.Name, "clientId"),
		})
	}

	missingOrInvalidIssuerUrls := findAuthProviderMissingOrInvalidIssuerUrl(config.Auth.GetOidcProviders())
	for _, p := range missingOrInvalidIssuerUrls {
		if p.Name == "" {
			continue
		}
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthProviderInvalidHttpUrlErrorString, p.Name, "issuerUrl"),
		})
	}

	missingOrInvalidTokenUrls := findAuthProviderMissingOrInvalidTokenUrl(config.Auth.GetOAuthProviders())
	for _, p := range missingOrInvalidTokenUrls {
		if p.Name == "" {
			continue
		}
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthProviderInvalidHttpUrlErrorString, p.Name, "tokenUrl"),
		})
	}

	missingOrInvalidAuthUrls := findAuthProviderMissingOrInvalidAuthorizationUrl(config.Auth.GetOAuthProviders())
	for _, p := range missingOrInvalidAuthUrls {
		if p.Name == "" {
			continue
		}
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthProviderInvalidHttpUrlErrorString, p.Name, "authorizationUrl"),
		})
	}

	if config.Auth.RedirectUrl != nil {
		_, err := url.ParseRequestURI(*config.Auth.RedirectUrl)
		if err != nil {
			errors = append(errors, &ConfigError{
				Type:    "invalid",
				Message: fmt.Sprintf(ConfigAuthInvalidRedirectUrlErrorString, *config.Auth.RedirectUrl),
			})
		}
	}

	if config.Auth.Hooks != nil {
		for _, v := range config.Auth.Hooks {
			if !slices.Contains(SupportedAuthHooks, v) {
				errors = append(errors, &ConfigError{
					Type:    "invalid",
					Message: fmt.Sprintf(ConfigAuthInvalidHook, v),
				})
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return &ConfigErrors{
		Errors: errors,
	}
}

// checkForDuplicates checks for duplicate environment variables in a keel project
func checkForDuplicates(config *ProjectConfig) (bool, map[string][]string) {
	results := make(map[string][]string, 2)
	envDuplicates, staging := findDuplicates(config.Environment)

	if len(staging) > 0 {
		for _, key := range staging {
			results[key] = append(results[key], "staging")
		}
	}

	secretDuplicates, secrets := findDuplicates(config.Secrets)
	if len(secrets) > 0 {
		for _, key := range secrets {
			results[key] = append(results[key], "secrets")
		}
	}

	if envDuplicates || secretDuplicates {
		return true, results
	}

	return false, results
}

// findDuplicates checks for duplicate environment variables or secrets for a given environment
func findDuplicates(environment []Input) (bool, []string) {
	keys := make(map[string]bool)

	duplicates := []string{}
	for _, envVar := range environment {
		if _, value := keys[envVar.Name]; !value {
			keys[envVar.Name] = true
		} else {
			duplicates = append(duplicates, envVar.Name)
		}
	}

	return len(duplicates) > 0, duplicates
}

// validateFormat checks if any secret name or environment variables are written in the wrong format.
// must be screaming snakecase or, if an environment variable, a non-reserved name.
func validateFormat(config *ProjectConfig, formatType string) (bool, map[string]bool) {
	envs := config.Environment
	secrets := config.Secrets

	incorrectNamesMap := make(map[string]bool)

	if formatType == "snakecase" {
		for _, secret := range secrets {
			if ok := incorrectNamesMap[secret.Name]; ok {
				continue
			}

			ssName := strings.ToUpper(secret.Name)

			if secret.Name != ssName {
				incorrectNamesMap[secret.Name] = true
			}
			continue
		}
	}

	for _, envVar := range envs {
		if ok := incorrectNamesMap[envVar.Name]; ok {
			continue
		}

		switch formatType {
		case "snakecase":
			ssName := casing.ToScreamingSnake(envVar.Name)

			if envVar.Name != ssName {
				incorrectNamesMap[envVar.Name] = true
			}
			continue
		case "reserved":
			found := reservedEnvVarRegex.FindString(envVar.Name)

			if found != "" {
				incorrectNamesMap[envVar.Name] = true
			}
			continue
		default:
			break
		}
	}

	return len(incorrectNamesMap) > 0, incorrectNamesMap
}

func (c *ProjectConfig) ValidateSecrets(localSecrets map[string]string) (bool, []string) {
	var missing []string

	for _, secret := range c.Secrets {
		if _, ok := localSecrets[secret.Name]; !ok {
			missing = append(missing, secret.Name)
		}
	}

	return len(missing) > 0, missing
}
