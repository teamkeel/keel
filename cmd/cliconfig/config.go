package cliconfig

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	configFile string
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

	absolutePath, err := filepath.Abs(options.WorkingDir)
	if err != nil {
		panic(err)
	}

	if options != nil && options.FileName != "" {
		viper.SetConfigFile(options.FileName)
		err := checkConfigFileExists(viper, options.FileName, absolutePath)
		if err != nil {
			panic(err)
		}

		return &Config{
			viper:      viper,
			configFile: options.FileName,
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	userConfigPath := path.Join(homeDir, ".keel", "config.yaml")

	viper.SetConfigFile(userConfigPath)

	err = checkConfigFileExists(viper, userConfigPath, absolutePath)
	if err != nil {
		panic(err)
	}

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	return &Config{
		viper:      viper,
		configFile: userConfigPath,
	}
}

func (c *Config) GetConfig(path string) (*UserConfig, error) {
	var cfg UserConfig

	b, err := os.ReadFile(c.configFile)
	if os.IsNotExist(err) {
		return createEmptyConfig(c.viper, c.configFile, path)
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
		newProjectConfig, err := createProject(c, path)
		if err != nil {
			return nil, fmt.Errorf("Unable to setup config in project directory: %s", path)
		}
		useConfig := newProjectConfig.Projects[path]
		return &useConfig, nil
	}

	return &project, nil
}

func (c *Config) SetSecret(path, environment, key, value string) error {
	cfg, err := c.GetConfig(path)
	if err != nil {
		return err
	}

	currentSecrets := cfg.Projects[path].Secrets
	secrets, err := createSecretEnvironments(environment, key, value, &currentSecrets)
	if err != nil {
		return err
	}
	cfg.Projects[path] = Project{
		Secrets: secrets,
	}

	return c.writeConfig(*cfg)
}

func (c *Config) RemoveSecret(path, environment, key string) error {
	cfg, err := c.GetConfig(path)
	if err != nil {
		return err
	}

	currentSecrets := cfg.Projects[path].Secrets

	switch environment {
	case "development":
		delete(currentSecrets.Development, key)
	case "test":
		delete(currentSecrets.Test, key)
	default:
		return errors.New("invalid environment " + environment)
	}

	cfg.Projects[path] = Project{
		Secrets: currentSecrets,
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
		return nil, errors.New("invalid environment " + environment)
	}
}

func (c *Config) writeConfig(cfg interface{}) error {
	reflectCfg := reflect.ValueOf(cfg)
	for i := range reflectCfg.NumField() {
		k := reflectCfg.Type().Field(i).Name
		v := reflectCfg.Field(i).Interface()

		c.viper.Set(k, v)
	}

	err := createPathIfNotExist(c.configFile)
	if err != nil {
		return err
	}

	return c.viper.WriteConfig()
}

func createEmptyConfig(v *viper.Viper, configPath, wd string) (*UserConfig, error) {
	err := createPathIfNotExist(configPath)
	if err != nil {
		return nil, err
	}

	projects := make(map[string]Project)
	project := Project{
		Secrets: createEnvironments(),
	}

	projects[wd] = project
	v.Set("projects", projects)

	config := &UserConfig{
		Projects: projects,
	}

	encoded, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(configPath, encoded, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &UserConfig{
		Projects: projects,
	}, nil
}

func createProject(c *Config, wd string) (*UserConfig, error) {
	var cfg UserConfig

	b, err := os.ReadFile(c.configFile)
	if os.IsNotExist(err) {
		return createEmptyConfig(c.viper, c.configFile, wd)
	} else if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	projects := cfg.Projects
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
}

func createSecretEnvironments(environment, key, value string, secrets *EnvironmentSecret) (EnvironmentSecret, error) {
	var environments EnvironmentSecret

	if secrets.Development == nil || secrets.Test == nil {
		return createEnvironments(), nil
	} else {
		environments = *secrets
	}

	switch environment {
	case "development":
		environments.Development[key] = value
	case "test":
		environments.Test[key] = value
	default:
		return EnvironmentSecret{}, errors.New("invalid environment " + environment)
	}

	return environments, nil
}

func createEnvironments() EnvironmentSecret {
	return EnvironmentSecret{
		Development: make(map[string]string),
		Test:        make(map[string]string),
	}
}

func checkConfigFileExists(viper *viper.Viper, configPath, wd string) error {
	err := viper.ReadInConfig()
	if os.IsNotExist(err) {
		_, err = createEmptyConfig(viper, configPath, wd)
		if err != nil {
			return fmt.Errorf("Unable to create config file: %s", err)
		}
	}
	return nil
}

func createPathIfNotExist(path string) error {
	dir := filepath.Dir(path)

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
