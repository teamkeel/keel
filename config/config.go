package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/samber/lo"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

const Empty = ""

//go:embed schema.json
var jsonSchema string

type ConfigFile struct {
	Filename string
	Env      string
	Config   *ProjectConfig
	Errors   *ConfigErrors
}

// ProjectConfig is the configuration for a keel project
type ProjectConfig struct {
	Environment   []EnvironmentVariable `yaml:"environment"`
	UseDefaultApi *bool                 `yaml:"useDefaultApi,omitempty"`
	Secrets       []Secret              `yaml:"secrets"`
	Auth          AuthConfig            `yaml:"auth"`
	DisableAuth   bool                  `yaml:"disableKeelAuth"`
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

func (c *ProjectConfig) UsesAuthHook(hook FunctionHook) bool {
	return slices.Contains(c.Auth.Hooks, hook)
}

// EnvironmentVariable is the configuration for a keel environment variable or secret
type EnvironmentVariable struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value,omitempty"`
}

type Secret struct {
	Name string `yaml:"name"`
}

type ConfigError struct {
	Message string `json:"message,omitempty"`
}

const (
	ConfigAuthProviderInvalidName          = "auth provider name '%s' must only include alphanumeric characters and underscores, and cannot start with a number"
	ConfigAuthProviderDuplicateErrorString = "auth provider name '%s' has been defined more than once, but must be unique"
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

func ToConfigErrors(err error) *ConfigErrors {
	v, ok := err.(*ConfigErrors)
	if !ok {
		return nil
	}
	return v
}

func LoadAll(dir string) ([]*ConfigFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := []*ConfigFile{}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "keelconfig") && strings.HasSuffix(entry.Name(), ".yaml") {
			c, err := Load(filepath.Join(dir, entry.Name()))
			if err != nil && ToConfigErrors(err) == nil {
				return nil, err
			}

			parts := strings.Split(entry.Name(), ".")
			env := ""
			if len(parts) == 3 {
				env = parts[1]
			}

			files = append(files, &ConfigFile{
				Filename: entry.Name(),
				Env:      env,
				Config:   c,
				Errors:   ToConfigErrors(err),
			})
		}
	}

	return files, nil
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
					Message: fmt.Sprintf("could not unmarshal config file: %s", err.Error()),
				},
			},
		}
	}

	var yamlData map[string]interface{}
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return nil, &ConfigErrors{
			Errors: []*ConfigError{
				{
					Message: fmt.Sprintf("could not unmarshal config file: %s", err.Error()),
				},
			},
		}
	}

	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return nil, &ConfigErrors{
			Errors: []*ConfigError{
				{
					Message: fmt.Sprintf("error converting YAML to JSON for validation: %s", err.Error()),
				},
			},
		}
	}

	// Special case - if the config is empty then we'll end up with null here. Since an empty
	// config file is ok we can just return a plain config here
	if string(jsonData) == "null" {
		return &config, nil
	}

	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Fatalf("Error validating JSON: %v", err)
	}

	errors := &ConfigErrors{}

	for _, err := range result.Errors() {
		errors.Errors = append(errors.Errors, &ConfigError{
			Message: err.String(),
		})
	}

	for _, fn := range validators {
		errors.Errors = append(errors.Errors, fn(&config)...)
	}

	if len(errors.Errors) == 0 {
		return &config, nil
	}

	return &config, errors
}

type ValidationFunc func(c *ProjectConfig) []*ConfigError

var validators = []ValidationFunc{
	validateUniqueNames,
	validateReservedPrefixes,
	validateAuthProviders,
}

func validateReservedPrefixes(c *ProjectConfig) []*ConfigError {
	errors := []*ConfigError{}

	values := lo.Map(c.Environment, func(v EnvironmentVariable, _ int) string {
		return v.Name
	})
	errors = append(errors, validateReserved(values, "environment.%d.name")...)

	values = lo.Map(c.Secrets, func(v Secret, _ int) string {
		return v.Name
	})
	errors = append(errors, validateReserved(values, "secrets.%d.name")...)

	return errors
}

var ReservedPrefixes = []string{"KEEL_", "OTEL_", "AWS_"}

func validateReserved(values []string, path string) []*ConfigError {
	errors := []*ConfigError{}

	for i, v := range values {
		for _, p := range ReservedPrefixes {
			if strings.HasPrefix(v, p) {
				errors = append(errors, &ConfigError{
					Message: fmt.Sprintf("%s: The '%s' prefix is not allowed", fmt.Sprintf(path, i), p),
				})
			}
		}
	}

	return errors
}

func validateUniqueNames(c *ProjectConfig) []*ConfigError {
	errors := []*ConfigError{}

	values := lo.Map(c.Environment, func(v EnvironmentVariable, _ int) string {
		return v.Name
	})
	errors = append(errors, validateUnique(values, "environment.%d.name")...)

	values = lo.Map(c.Secrets, func(v Secret, _ int) string {
		return v.Name
	})
	errors = append(errors, validateUnique(values, "secrets.%d.name")...)

	values = lo.Map(c.Auth.Providers, func(p Provider, _ int) string {
		return p.Name
	})
	errors = append(errors, validateUnique(values, "auth.providers.%d.name")...)

	return errors
}

func validateUnique(values []string, path string) []*ConfigError {
	seen := map[string]bool{}
	errors := []*ConfigError{}
	for i, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			key := strings.Split(path, ".")
			errors = append(errors, &ConfigError{
				Message: fmt.Sprintf("%s: Duplicate %s %s", fmt.Sprintf(path, i), key[len(key)-1], v),
			})
		}
		seen[v] = true
	}
	return errors
}

func validateAuthProviders(c *ProjectConfig) []*ConfigError {
	errors := []*ConfigError{}
	for i, p := range c.Auth.Providers {
		if strings.HasPrefix(strings.ToLower(p.Name), "keel_") {
			errors = append(errors, &ConfigError{
				Message: fmt.Sprintf("auth.providers.%d.name: Cannot start with '%s'", i, p.Name[0:5]),
			})
		}
		if p.Type == "oidc" && p.AuthorizationUrl == "" {
			errors = append(errors, &ConfigError{
				Message: fmt.Sprintf("auth.providers.%d: 'authorizationUrl' is required if 'type' is 'oidc'", i),
			})
		}
		if p.Type == "oidc" && p.IssuerUrl == "" {
			errors = append(errors, &ConfigError{
				Message: fmt.Sprintf("auth.providers.%d: 'issuerUrl' is required if 'type' is 'oidc'", i),
			})
		}
		if p.Type == "oidc" && p.TokenUrl == "" {
			errors = append(errors, &ConfigError{
				Message: fmt.Sprintf("auth.providers.%d: 'tokenUrl' is required if 'type' is 'oidc'", i),
			})
		}
	}

	return errors
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
