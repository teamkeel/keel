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
	Test        map[string]string `yaml:"test"`
}

type Options struct {
	FileName   string
	WorkingDir string
}

func New(options *Options) *Config {
	viper := viper.New()

	if options != nil && options.FileName != "" {
		viper.SetConfigFile(options.FileName)
		err := checkConfigFileExists(viper, options.WorkingDir)
		if err != nil {
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

	err = checkConfigFileExists(viper, options.WorkingDir)
	if err != nil {
		panic(err)
	}

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	return &Config{
		viper:      viper,
		configPath: userConfigPath,
	}
}

func (c *Config) GetConfig(path string) (*UserConfig, error) {
	var cfg UserConfig

	b, err := os.ReadFile(c.configPath)
	if os.IsNotExist(err) {
		return createEmptyConfig(c.viper, path)
	} else if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &cfg)
	return &cfg, err
}

func (c *Config) GetProject(path string) (*Project, error) {
	cfg, err := c.GetConfig(path)
	if err != nil {
		return nil, err
	}

	project, ok := cfg.Projects[path]
	if !ok {
		return nil, errors.New("project not found")
	}

	return &project, nil
}

func (c *Config) SetSecret(path, environment, key, value string) error {
	cfg, err := c.GetConfig(path)
	if err != nil {
		return err
	}

	currentSecrets := cfg.Projects[path].Secrets
	secrets := createSecretEnvironments(environment, key, value, &currentSecrets)

	cfg.Projects[path] = Project{
		Secrets: secrets,
	}

	return c.writeConfig(*cfg)
}

func (c *Config) GetSecrets(path, environment string) (map[string]string, error) {
	project, err := c.GetProject(path)
	if err != nil {
		return nil, err
	}

	switch environment {
	case "development":
		return project.Secrets.Development, nil
	case "test":
		return project.Secrets.Test, nil
	default:
		return nil, errors.New("invalid environment")
	}
}

func (c *Config) writeConfig(cfg interface{}) error {
	reflectCfg := reflect.ValueOf(cfg)
	for i := 0; i < reflectCfg.NumField(); i++ {
		k := reflectCfg.Type().Field(i).Name
		v := reflectCfg.Field(i).Interface()

		c.viper.Set(k, v)
	}

	err := c.createPathIfNotExist(c.configPath)
	if err != nil {
		return err
	}

	return c.viper.WriteConfig()
}

func (c *Config) createPathIfNotExist(path string) error {
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
	default:
		panic("invalid environment")

	}

	return environments
}

func createEnvironments() EnvironmentSecret {
	return EnvironmentSecret{
		Development: make(map[string]string),
		Test:        make(map[string]string),
	}
}

func checkConfigFileExists(viper *viper.Viper, path string) error {
	err := viper.ReadInConfig()
	if os.IsNotExist(err) {
		_, err = createEmptyConfig(viper, path)
		if err != nil {
			return err
		}
	}
	return nil
}
