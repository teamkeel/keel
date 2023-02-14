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
	Secrets EnvironmentSecret `yaml:"secrets"`
}

type EnvironmentSecret struct {
	Development map[string]string `yaml:"development"`
	Staging     map[string]string `yaml:"staging"`
	Production  map[string]string `yaml:"production"`
	Test        map[string]string `yaml:"test"`
}

type Options struct {
	FileName string
}

func New(options *Options) *Config {
	viper := viper.New()

	if options != nil {
		viper.SetConfigFile(options.FileName)
		err := viper.ReadInConfig()
		if os.IsNotExist(err) {
			wd, err := os.Getwd()
			if err != nil {
				panic(err)
			}

			_, err = createEmptyConfig(viper, wd)
			if err != nil {
				panic(err)
			}
		} else if err != nil {
			panic(err)
		}

		return &Config{
			viper:      viper,
			configPath: options.FileName,
		}
	}

	UserConfigPartialPath := ".keel/config.yaml"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	userConfigPath := path.Join(homeDir, UserConfigPartialPath)

	viper.SetConfigFile(userConfigPath)

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	return &Config{
		viper:      viper,
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
		return createEmptyConfig(c.viper, wd)
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

func createEmptyConfig(v *viper.Viper, wd string) (*UserConfig, error) {
	projects := make(map[string]Project)
	project := Project{
		Secrets: createEnvironments(),
	}

	projects[wd] = project
	v.Set("projects", projects)

	err := v.WriteConfig()
	if err != nil {
		return nil, err
	}

	return &UserConfig{
		Projects: projects,
	}, nil
}

func createSecretEnvironments(environment, key, value string, secrets *EnvironmentSecret) EnvironmentSecret {
	var environments EnvironmentSecret

	if secrets == nil {
		environments = createEnvironments()
		return environments
	} else {
		environments = *secrets
	}

	switch environment {
	case "development":
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

func createEnvironments() EnvironmentSecret {
	return EnvironmentSecret{
		Development: make(map[string]string),
		Test:        make(map[string]string),
		Staging:     make(map[string]string),
		Production:  make(map[string]string),
	}
}
