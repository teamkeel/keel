package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/teamkeel/keel/casing"
	"gopkg.in/yaml.v3"
)

// Config is the configuration for the Keel runtime
type Config struct{}

// ProjectConfig is the configuration for a keel project
type ProjectConfig struct {
	Environment EnvironmentConfig `yaml:"environment"`
	Secrets     []Input           `yaml:"secrets"`
	Auth        AuthConfig        `yaml:"auth"`
	DisableAuth bool              `yaml:"disableKeelAuth"`
}

func (p *ProjectConfig) GetEnvVars(env string) map[string]string {
	nameToValueMap := map[string]string{}

	var environmentInput []Input
	switch env {
	case "test":
		environmentInput = p.Environment.Test
	case "development":
		environmentInput = p.Environment.Development
	case "staging":
		environmentInput = p.Environment.Staging
	case "production":
		environmentInput = p.Environment.Production
	default:
		environmentInput = p.Environment.Default
	}

	// make a copy of the default env vars
	merged := append([]Input{}, p.Environment.Default...)
	// merge with specific environment env vars
	merged = append(merged, environmentInput...)

	for _, input := range merged {
		// override default env var with specific environment one, if it exists
		nameToValueMap[input.Name] = input.Value
	}

	return nameToValueMap
}

// AllEnvironmentVariables returns a slice of all of the unique environment variable key names
// defined across all environments
func (c *ProjectConfig) AllEnvironmentVariables() []string {
	var environmentVariables []string

	for _, envVar := range c.Environment.Default {
		environmentVariables = append(environmentVariables, envVar.Name)
	}

	for _, envVar := range c.Environment.Staging {
		environmentVariables = append(environmentVariables, envVar.Name)
	}

	for _, envVar := range c.Environment.Development {
		environmentVariables = append(environmentVariables, envVar.Name)
	}

	for _, envVar := range c.Environment.Production {
		environmentVariables = append(environmentVariables, envVar.Name)
	}

	duplicateKeys := make(map[string]bool)
	allEnvironmentVariables := []string{}
	for _, item := range environmentVariables {
		if _, value := duplicateKeys[item]; !value {
			duplicateKeys[item] = true
			allEnvironmentVariables = append(allEnvironmentVariables, item)
		}
	}

	return allEnvironmentVariables
}

func (c *ProjectConfig) AllSecrets() []string {
	var secrets []string

	for _, secret := range c.Secrets {
		secrets = append(secrets, secret.Name)
	}

	return secrets
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
	Name     string   `yaml:"name"`
	Value    string   `yaml:"value,omitempty"`
	Required []string `yaml:"required,omitempty"`
}

type ConfigError struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}

const (
	ConfigDuplicateErrorString                       = "environment variable %s has a duplicate set in environment: %s"
	ConfigRequiredErrorString                        = "environment variable %s is required but not defined in the following environments: %s"
	ConfigIncorrectNamingErrorString                 = "%s must be written in upper snakecase"
	ConfigReservedNameErrorString                    = "environment variable %s cannot start with %s as it is reserved"
	ConfigAuthTokenExpiryMustBePositive              = "%s token lifespan cannot be negative or zero for field: %s"
	ConfigAuthProviderMissingFieldAtIndexErrorString = "auth provider at index %v is missing field: %s"
	ConfigAuthProviderMissingFieldErrorString        = "auth provider '%s' is missing field: %s"
	ConfigAuthProviderInvalidTypeErrorString         = "auth provider '%s' has invalid type '%s' which must be one of: %s"
	ConfigAuthProviderDuplicateErrorString           = "auth provider name '%s' has been defined more than once, but must be unique"
	ConfigAuthProviderInvalidHttpUrlErrorString      = "auth provider '%s' has missing or invalid https url for field: %s"
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

	var config ProjectConfig
	err = yaml.Unmarshal(loadConfig, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %w", err)
	}

	validationErrors := Validate(&config)
	if validationErrors != nil {
		return &config, validationErrors
	}

	return &config, nil
}

func LoadFromBytes(data []byte) (*ProjectConfig, error) {
	var config ProjectConfig

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %w", err)
	}

	validationErrors := Validate(&config)
	if validationErrors != nil {
		return &config, validationErrors
	}

	return &config, nil
}

var reservedEnvVarRegex = regexp.MustCompile(`^AWS_|^_|^OTEL_|^OPENCOLLECTOR_CONFIG|^KEEL_`)

func Validate(config *ProjectConfig) *ConfigErrors {
	var errors []*ConfigError

	duplicates, results := checkForDuplicates(config)
	if duplicates {
		for duplicatedEnvVarName, environmentNames := range results {
			errors = append(errors, &ConfigError{
				Type:    "duplicate",
				Message: fmt.Sprintf(ConfigDuplicateErrorString, duplicatedEnvVarName, environmentNames),
			})
		}
	}

	missingKeys, keys := requiredValuesKeys(config)
	if missingKeys {
		for requiredValueName, environmentNames := range keys {
			errors = append(errors, &ConfigError{
				Type:    "missing",
				Message: fmt.Sprintf(ConfigRequiredErrorString, requiredValueName, environmentNames),
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

	if config.Auth.Tokens != nil && config.Auth.Tokens.AccessTokenExpiry <= 0 {
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthTokenExpiryMustBePositive, "access", "accessTokenExpiry"),
		})
	}

	if config.Auth.Tokens != nil && config.Auth.Tokens.RefreshTokenExpiry <= 0 {
		errors = append(errors, &ConfigError{
			Type:    "invalid",
			Message: fmt.Sprintf(ConfigAuthTokenExpiryMustBePositive, "refresh", "refreshTokenExpiry"),
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

	if len(errors) == 0 {
		return nil
	}

	return &ConfigErrors{
		Errors: errors,
	}
}

// checkForDuplicates checks for duplicate environment variables in a keel project
// We assume that any environment variable that is defined in staging/production will override the default
func checkForDuplicates(config *ProjectConfig) (bool, map[string][]string) {
	results := make(map[string][]string, 2)
	stagingDuplicates, staging := findDuplicates(config.Environment.Staging)

	if len(staging) > 0 {
		for _, key := range staging {
			results[key] = append(results[key], "staging")
		}
	}
	if len(config.Environment.Production) == 0 {
		return stagingDuplicates, results
	}

	productionDuplicates, production := findDuplicates(config.Environment.Production)
	if len(production) > 0 {
		for _, key := range production {
			results[key] = append(results[key], "production")
		}
	}

	secretDuplicates, secrets := findDuplicates(config.Secrets)
	if len(secrets) > 0 {
		for _, key := range secrets {
			results[key] = append(results[key], "secrets")
		}
	}

	if stagingDuplicates || productionDuplicates || secretDuplicates {
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

// requiredValuesKeys checks for required environment variables in a keel project defined in the default block
// A required environment variable must be defined in either staging or production
func requiredValuesKeys(config *ProjectConfig) (bool, map[string][]string) {
	results := make(map[string][]string, 2)

	for _, v := range config.Environment.Default {
		if v.Required == nil {
			continue
		}

		for _, required := range v.Required {
			if required == "staging" {
				if !contains(config.Environment.Staging, v.Name) {
					results[v.Name] = append(results[v.Name], "staging")
				}
			}
			if required == "production" {
				if !contains(config.Environment.Production, v.Name) {
					results[v.Name] = append(results[v.Name], "production")
				}
			}
		}
	}

	return len(results) > 0, results
}

func contains(s []Input, e string) bool {
	for _, input := range s {
		if input.Name == e {
			return true
		}
	}

	return false
}

// validateFormat checks if any secret name or environment variables in default, staging and production environments
// are written in the wrong format. Must be screaming snakecase or, if an environment variable, a non-reserved name.
func validateFormat(config *ProjectConfig, formatType string) (bool, map[string]bool) {
	defaultEnv := config.Environment.Default
	stagingEnv := config.Environment.Staging
	productionEnv := config.Environment.Production
	testEnv := config.Environment.Test
	secrets := config.Secrets

	envsToCheck := [][]Input{defaultEnv, stagingEnv, productionEnv, testEnv}

	incorrectNamesMap := make(map[string]bool)

	if formatType == "snakecase" {
		for _, secret := range secrets {
			if ok := incorrectNamesMap[secret.Name]; ok {
				continue
			}

			ssName := casing.ToScreamingSnake(secret.Name)

			if secret.Name != ssName {
				incorrectNamesMap[secret.Name] = true
			}
			continue
		}
	}

	for _, environment := range envsToCheck {
		for _, envVar := range environment {
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
