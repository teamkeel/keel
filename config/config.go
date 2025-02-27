package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/samber/lo"
	"github.com/xeipuuv/gojsonschema"
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
	Console       ConsoleConfig         `yaml:"console"`
	Deploy        *DeployConfig         `yaml:"deploy,omitempty"`
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
	Filename        string    `json:"filename"`
	Type            string    `json:"type"`
	Message         string    `json:"message,omitempty"`
	Field           string    `json:"field"`
	Pos             *Position `json:"pos"`
	EndPos          *Position `json:"endPos"`
	AnnotatedSource string    `json:"-"`
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
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

	return parseAndValidate(loadConfig, dir)
}

func LoadFromBytes(data []byte, filename string) (*ProjectConfig, error) {
	return parseAndValidate(data, filename)
}

func parseAndValidate(data []byte, filename string) (*ProjectConfig, error) {
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
		return nil, err
	}

	errors := &ConfigErrors{}

	for _, err := range result.Errors() {
		errors.Errors = append(errors.Errors, &ConfigError{
			Filename: filename,
			Message:  err.String(),
			Field:    err.Field(),
			Type:     err.Type(),
		})
	}

	for _, fn := range validators {
		errs := fn(&config)
		for _, e := range errs {
			e.Filename = filename
		}
		errors.Errors = append(errors.Errors, errs...)
	}

	if len(errors.Errors) == 0 {
		return &config, nil
	}

	err = annotateErrors(data, errors)
	if err != nil {
		return nil, err
	}

	return &config, errors
}

type ValidationFunc func(c *ProjectConfig) []*ConfigError

var validators = []ValidationFunc{
	validateUniqueNames,
	validateReservedPrefixes,
	validateAuthProviders,
	validateDatabase,
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
					Field:   fmt.Sprintf(path, i),
					Type:    "reserved-prefix",
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
				Field:   fmt.Sprintf(path, i),
				Type:    "duplicate-value",
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
				Field:   fmt.Sprintf("auth.providers.%d.name", i),
				Type:    "reserved-prefix",
			})
		}
		if p.Type == "oidc" && p.IssuerUrl == "" {
			errors = append(errors, &ConfigError{
				Message: fmt.Sprintf("auth.providers.%d: 'issuerUrl' is required if 'type' is 'oidc'", i),
				Field:   fmt.Sprintf("auth.providers.%d", i),
				Type:    "required",
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

var arrayIndexRegex = regexp.MustCompile(`\.(\d+)(\.){0,1}`)
var additionalPropertyRegex = regexp.MustCompile(`Additional property (\w+) is not allowed`)

func annotateErrors(src []byte, errors *ConfigErrors) error {
	file, err := parser.ParseBytes(src, 0)
	if err != nil {
		return err
	}

	for _, e := range errors.Errors {
		field := e.Field

		// Get rid of "(root)" which isn't valid in YAML path
		if field == "(root)" {
			field = ""
		}

		// change "foo.0.baz' into "foo[0].baz"
		field = arrayIndexRegex.ReplaceAllString(field, "[$1]$2")
		field = fmt.Sprintf("$.%s", field)

		// Special case for additional properties - add that property to the end so
		// we point to it
		if e.Type == "additional_property_not_allowed" {
			prop := additionalPropertyRegex.FindStringSubmatch(e.Message)
			field = fmt.Sprintf("%s.%s", field, prop[1])
		}

		path, err := yaml.PathString(field)
		if err != nil {
			// If the YAML path ends up being invalid we'll just continue, don't want to
			// blow up on this
			fmt.Println("invalid YAMLPath:", field, err.Error())
			continue
		}

		node, err := path.FilterFile(file)
		if err != nil {
			// This shouldn't really happen if the YAML path is valid but we'll just continue if it does
			fmt.Println("unable to filter YAML file", err.Error())
			continue
		}

		pos := node.GetToken().Position
		e.Pos = &Position{
			Line:   pos.Line,
			Column: pos.Column,
		}
		e.EndPos = &Position{
			Line:   pos.Line,
			Column: pos.Column + len(node.GetToken().Value),
		}

		annotated, err := path.AnnotateSource(src, true)
		if err != nil {
			// Again not sure what kind of error can happen here but we'll just continue
			fmt.Println("error annotating soure", err.Error())
			continue
		}

		e.AnnotatedSource = string(annotated)
	}

	return nil
}
