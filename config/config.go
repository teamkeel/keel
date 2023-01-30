package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for the Keel runtime
type Config struct{}

// ProjectConfig is the configuration for a keel project
type ProjectConfig struct {
	Environment EnvironmentConfig `yaml:"environment"`
	Secrets     []Input           `yaml:"secrets"`
}

// EnvironmentConfig is the configuration for a keel environment default, staging, production
type EnvironmentConfig struct {
	Default    []Input `yaml:"default"`
	Staging    []Input `yaml:"staging"`
	Production []Input `yaml:"production,omitempty"`
}

// Input is the configuration for a keel environment variable or secret
type Input struct {
	Name     string   `yaml:"name"`
	Value    string   `yaml:"value,omitempty"`
	Required []string `yaml:"required,omitempty"`
}

func Load(dir string) (*ProjectConfig, error) {
	loadConfig, err := os.ReadFile(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config ProjectConfig
	err = yaml.Unmarshal(loadConfig, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %w", err)
	}

	err = Validate(&config)
	if err != nil {
		return &config, fmt.Errorf("could not validate config file: %w", err)
	}

	return &config, nil
}

func Validate(config *ProjectConfig) error {
	duplicates, results := checkForDuplicates(config)
	if duplicates {
		for k, v := range results {
			return fmt.Errorf("duplicate environment variables found in %s: %v", k, strings.Join(v[:], ","))
		}
	}

	missingKeys, keys := requiredValuesKeys(config)
	if missingKeys {
		for k, v := range keys {
			return fmt.Errorf("missing required environment variables in %s: %v", k, strings.Join(v[:], ","))
		}
	}

	return nil
}

// checkForDuplicates checks for duplicate environment variables in a keel project
// We assume that any environment variable that is defined in staging/production will override the default
func checkForDuplicates(config *ProjectConfig) (bool, map[string][]string) {
	results := make(map[string][]string, 2)
	stagingDuplicates, staging := findDuplicates(config.Environment.Staging)

	if len(staging) > 0 {
		results["staging"] = staging
	}
	if len(config.Environment.Production) == 0 {
		return stagingDuplicates, results
	}

	productionDuplicates, production := findDuplicates(config.Environment.Production)
	if len(production) > 0 {
		results["production"] = production
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
	var stagingMissing []string
	var productionMissing []string

	for _, v := range config.Environment.Default {
		if v.Required == nil {
			continue
		}

		for _, required := range v.Required {
			if required == "staging" {
				if !contains(config.Environment.Staging, v.Name) {
					stagingMissing = append(stagingMissing, v.Name)
				}
			}
			if required == "production" {
				if !contains(config.Environment.Production, v.Name) {
					productionMissing = append(productionMissing, v.Name)
				}
			}
		}
	}

	results := make(map[string][]string, 2)

	if len(stagingMissing) > 0 {
		results["staging"] = stagingMissing
	}

	if len(productionMissing) > 0 {
		results["production"] = productionMissing
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
