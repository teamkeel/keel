package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
)

type Config struct {
	configPath string
	viper      *viper.Viper
}

type RootConfig struct {
	Context  string                   `json:"context"`
	User     UserConfig               `json:"user"`
	Projects map[string]ProjectConfig `json:"projects"`
}

type UserConfig struct {
	Token string `json:"token"`
}

type ProjectConfig struct {
	Project      string                  `json:"project,omitempty"`
	Environments map[string]Environments `json:"environments,omitempty"`
}

type Environments struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func New() *Config {
	rootViper := viper.New()
	rootConfigPartialPath := ".keel/config.json"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	rootConfigPath := path.Join(homeDir, rootConfigPartialPath)

	rootViper.SetConfigFile(rootConfigPath)
	err = rootViper.ReadInConfig()
	if os.IsNotExist(err) {
	} else if err != nil {
		fmt.Printf("Unable to parse keel config %s\n", err)
	}

	return &Config{
		viper:      rootViper,
		configPath: rootConfigPath,
	}
}

func (c *Config) GetConfig() (*RootConfig, error) {
	var cfg RootConfig
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(homedir + "/.keel/config.json")
	if os.IsNotExist(err) {
		return nil, errors.New("config not found")
	} else if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &cfg)
	return &cfg, err
}

func (c *Config) SetConfig(cfg *RootConfig) error {
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]ProjectConfig)
	}

	return c.generateConfig(c, *cfg)
}

func (c *Config) generateConfig(config *Config, cfg interface{}) error {
	reflectCfg := reflect.ValueOf(cfg)
	for i := 0; i < reflectCfg.NumField(); i++ {
		k := reflectCfg.Type().Field(i).Name
		v := reflectCfg.Field(i).Interface()
		config.viper.Set(k, v)
	}

	err := c.CreatePathIfNotExist(config.configPath)
	if err != nil {
		return err
	}

	return config.viper.WriteConfig()
}

func (c *Config) CreatePathIfNotExist(path string) error {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
