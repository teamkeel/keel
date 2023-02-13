package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// Config is the configuration for the Keel runtime
type Config struct{}

// ProjectConfig is the configuration for a keel project
type ProjectConfig struct {
	Environment EnvironmentConfig `yaml:"environment"`
	Secrets     []Input           `yaml:"secrets"`
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

func SetEnvVars(directory, environment string) {
	config, err := Load(directory)
	if err != nil {
		panic(err)
	}

	// Find another way to get the environment
	envVars := config.GetEnvVars(environment)
	for key, value := range envVars {
		os.Setenv(key, value)
	}

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

type ConfigErrors struct {
	Type         string   `json:"type,omitempty"`
	Key          string   `json:"key,omitempty"`
	Environments []string `json:"environments,omitempty"`
}

const (
	DuplicateErrorString = " - environment variable %s has a duplicate set in environment: %s\n"
	RequiredErrorString  = " - environment variable %s is required but not defined in the following environments: %s\n"
)

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
		return &config, generateOutput(validationErrors)
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
		return &config, generateOutput(validationErrors)
	}

	return &config, nil
}

func Validate(config *ProjectConfig) []ConfigErrors {
	var errors []ConfigErrors

	duplicates, results := checkForDuplicates(config)
	if duplicates {
		for k, v := range results {
			errors = append(errors, ConfigErrors{
				Type:         "duplicate",
				Key:          k,
				Environments: v,
			})
		}
	}

	missingKeys, keys := requiredValuesKeys(config)
	if missingKeys {
		for k, v := range keys {
			errors = append(errors, ConfigErrors{
				Type:         "missing",
				Key:          k,
				Environments: v,
			})
		}
	}

	return errors
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

	if stagingDuplicates || productionDuplicates {
		return true, results
	}

	return false, results
}

// findDuplicates checks for duplicate environment variables for a given environment
func findDuplicates(environment []Input) (bool, []string) {
	envVarKeys := make(map[string]bool)

	duplicateEnvVars := []string{}
	for _, envVar := range environment {
		if _, value := envVarKeys[envVar.Name]; !value {
			envVarKeys[envVar.Name] = true
		} else {
			duplicateEnvVars = append(duplicateEnvVars, envVar.Name)
		}
	}

	return len(duplicateEnvVars) > 0, duplicateEnvVars
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

func generateOutput(validationErrors []ConfigErrors) error {
	var errorString string
	for _, v := range validationErrors {
		if v.Type == "duplicate" {
			errorString = errorString + fmt.Sprintf(DuplicateErrorString, color.RedString(v.Key), v.Environments)
		}
		if v.Type == "missing" {
			errorString = errorString + fmt.Sprintf(RequiredErrorString, color.RedString(v.Key), v.Environments)
		}
	}

	return errors.New(errorString)
}
