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
	Production []Input `yaml:"production"`
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
func checkForDuplicates(config *ProjectConfig) (bool, map[string][]string) {
	stagingDuplicates, staging := getDuplicates(config.Environment.Staging)
	productionDuplicates, production := getDuplicates(config.Environment.Production)

	results := make(map[string][]string, 2)

	if len(staging) > 0 {
		results["staging"] = staging
	}
	if len(production) > 0 {
		results["production"] = production
	}

	if stagingDuplicates || productionDuplicates {
		return true, results
	}

	return false, results
}

func getDuplicates(environment []Input) (bool, []string) {
	allKeys := make(map[string]bool)

	dupes := []string{}
	for _, item := range environment {
		if _, value := allKeys[item.Name]; !value {
			allKeys[item.Name] = true
		} else {
			dupes = append(dupes, item.Name)
		}
	}

	return len(dupes) > 0, dupes
}

func requiredValuesKeys(config *ProjectConfig) (bool, map[string][]string) {
	var staging []string
	var production []string

	for _, v := range config.Environment.Default {
		if v.Required == nil {
			continue
		}

		for _, required := range v.Required {
			if required == "staging" {
				if !contains(config.Environment.Staging, v.Name) {
					staging = append(staging, v.Name)
				}
			}
			if required == "production" {
				if !contains(config.Environment.Production, v.Name) {
					production = append(production, v.Name)
				}
			}
		}
	}

	results := make(map[string][]string, 2)

	if len(staging) > 0 {
		results["staging"] = staging
	}

	if len(production) > 0 {
		results["production"] = production
	}

	return len(results) > 0, results
}

func contains(s []Input, e string) bool {
	for _, a := range s {
		if a.Name == e {
			return true
		}
	}
	return false
}
