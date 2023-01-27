package config

import (
	"fmt"
	"os"
	"path"

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

func Load(configFileName string) (*ProjectConfig, error) {
	dir, err := os.Getwd()
	if err != nil {
		color.Red("cannot read config file")
		os.Exit(1)
	}

	loadConfig, err := os.ReadFile(path.Join(dir, configFileName))
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config ProjectConfig
	err = yaml.Unmarshal(loadConfig, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %w", err)
	}

	return &config, nil
}
