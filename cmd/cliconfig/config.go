package cliconfig

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	configPath string
	viper      *viper.Viper
}

type UserConfig struct {
	Projects map[string]Project `yaml:"projects"`
}

type Project struct {
	Secrets Environments `yaml:"secrets"`
}

type Environments struct {
	Development map[string]string `yaml:"development"`
	Staging     map[string]string `yaml:"staging"`
	Production  map[string]string `yaml:"production"`
	Test        map[string]string `yaml:"test"`
}

type Options struct {
	Path string
}

func New(options *Options) *Config {
	rootViper := viper.New()

	if options != nil {
		rootViper.SetConfigFile(options.Path)
		err := rootViper.ReadInConfig()
		if os.IsNotExist(err) {
		} else if err != nil {
			panic(err)
		}
		return &Config{
			viper:      rootViper,
			configPath: options.Path,
		}
	}

	UserConfigPartialPath := ".keel/config.yaml"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	userConfigPath := path.Join(homeDir, UserConfigPartialPath)

	rootViper.SetConfigFile(userConfigPath)
	err = rootViper.ReadInConfig()
	if os.IsNotExist(err) {
	} else if err != nil {
		panic(err)
	}

	return &Config{
		viper:      rootViper,
		configPath: userConfigPath,
	}
}

func (c *Config) GetConfig() (*UserConfig, error) {
	var cfg UserConfig

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(c.configPath)
	if os.IsNotExist(err) {
		projects := make(map[string]Project)
		project := Project{
			Secrets: createEnvironments(),
		}

		projects[wd] = project
		c.viper.Set("projects", projects)

		err = c.viper.WriteConfig()
		if err != nil {
			return nil, err
		}

		return &UserConfig{
			Projects: projects,
		}, nil
	} else if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &cfg)
	return &cfg, err
}

func (c *Config) GetProject() (*Project, error) {
	cfg, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	project, ok := cfg.Projects[wd]
	if !ok {
		return nil, errors.New("project not found")
	}

	return &project, nil
}

func (c *Config) SetSecret(environment, key, value string) error {
	cfg, err := c.GetConfig()
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	currentSecrets := cfg.Projects[wd].Secrets
	secrets := createSecretEnvironments(environment, key, value, &currentSecrets)

	cfg.Projects[wd] = Project{
		Secrets: secrets,
	}

	return c.writeConfig(*cfg)
}

func (c *Config) writeConfig(cfg interface{}) error {
	reflectCfg := reflect.ValueOf(cfg)
	for i := 0; i < reflectCfg.NumField(); i++ {
		k := reflectCfg.Type().Field(i).Name
		v := reflectCfg.Field(i).Interface()

		c.viper.Set(k, v)
	}

	err := c.CreatePathIfNotExist(c.configPath)
	if err != nil {
		return err
	}

	return c.viper.WriteConfig()
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

func createSecretEnvironments(environment, key, value string, secrets *Environments) Environments {
	var environments Environments

	if secrets == nil {
		environments = createEnvironments()
		return environments
	} else {
		environments = *secrets
	}

	switch environment {
	case "run":
		environments.Development[key] = value
	case "test":
		environments.Test[key] = value
	case "staging":
		environments.Staging[key] = value
	case "production":
		environments.Production[key] = value
	default:
		panic("invalid environment")

	}

	return environments
}

func createEnvironments() Environments {
	return Environments{
		Development: make(map[string]string),
		Test:        make(map[string]string),
		Staging:     make(map[string]string),
		Production:  make(map[string]string),
	}
}
