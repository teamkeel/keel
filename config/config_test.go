package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	config, err := Load("fixtures/test_basic_config.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "TEST", config.Environment.Default[0].Name)
	assert.Equal(t, "test", config.Environment.Default[0].Value)
	assert.Equal(t, "API_KEY", config.Secrets[0].Name)
}

func TestFailConfigValue(t *testing.T) {
	_, err := Load("fixtures/test_failing_config.yaml")
	assert.Error(t, err)

	assert.Equal(t, "could not unmarshal config file: yaml: unmarshal errors:\n  line 5: cannot unmarshal !!seq into string", err.Error())
}

func TestDuplicates(t *testing.T) {
	config, err := Load("fixtures/test_duplicates.yaml")
	assert.Error(t, err)

	assert.Equal(t, "TEST", config.Environment.Default[0].Name)
	assert.Equal(t, "environment variable TEST has a duplicate set in environment: [staging]\n", err.Error())
}

func TestRequiredFail(t *testing.T) {
	config, err := Load("fixtures/test_required_fail.yaml")
	assert.Error(t, err)

	assert.Equal(t, "TEST", config.Environment.Staging[0].Name)
	assert.Equal(t, "environment variable TEST is required but not defined in the following environments: [production]\n", err.Error())
}

func TestRequired(t *testing.T) {
	config, err := Load("fixtures/test_required.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "TEST", config.Environment.Staging[0].Name)
	assert.Equal(t, "TEST", config.Environment.Production[0].Name)
}

func TestRequiredValuesKeys(t *testing.T) {
	config, err := Load("fixtures/test_required.yaml")
	assert.NoError(t, err)

	notOk, keys := requiredValuesKeys(config)
	assert.False(t, notOk)

	var nil []string
	assert.Equal(t, nil, keys["staging"])

	config, err = Load("fixtures/test_required_fail.yaml")
	assert.Error(t, err)

	notOk, keys = requiredValuesKeys(config)
	assert.True(t, notOk)
	assert.Equal(t, []string{"production"}, keys["TEST"])
}

func TestEmptyConfig(t *testing.T) {
	config, err := Load("fixtures/test_required_fail_empty.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "TEST", config.Environment.Staging[0].Name)
	assert.Equal(t, "test2_duplicate", config.Environment.Staging[0].Value)
}

func TestAllEnvironmentVariables(t *testing.T) {
	config, err := Load("fixtures/test_required.yaml")
	assert.NoError(t, err)

	allEnvVars := config.AllEnvironmentVariables()
	assert.Equal(t, []string{"TEST", "TEST_API_KEY"}, allEnvVars)
}

func TestGetEnvVars(t *testing.T) {
	config, err := Load("fixtures/test_merge_envs_config.yaml")
	assert.NoError(t, err)

	envVars := config.GetEnvVars("test")
	assert.Equal(t, map[string]string{"DOG_NAME": "Peggy", "PERSON_NAME": "Louisa"}, envVars)
}

func TestSnakecaseValidateFormat(t *testing.T) {
	_, err := Load("fixtures/test_snakecase_config.yaml")
	assert.Error(t, err)

	assert.NotContains(t, err.Error(), "THIS_IS_ALLOWED")

	assert.Contains(t, err.Error(), "environment variable this_is_not_Allowed1 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "environment variable THIS_IS_NOT_ALLOWEd2 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "environment variable thisIsNotAllowed3 must be written in upper snakecase\n")
	assert.Contains(t, err.Error(), "environment variable This_IS_nOT_AlloWED4 must be written in upper snakecase\n")
}

func TestReservedNameValidateFormat(t *testing.T) {
	_, err := Load("fixtures/test_reserved_name_config.yaml")
	assert.Error(t, err)

	assert.NotContains(t, err.Error(), "THIS_IS_ALLOWED")

	assert.Contains(t, err.Error(), "environment variable AWS_NOT_ALLOWED1 cannot start with AWS_ as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable KEEL_NOT_ALLOWED2 cannot start with KEEL_ as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable OTEL_NOT_ALLOWED3 cannot start with OTEL_ as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable OPENCOLLECTOR_CONFIG_NOT_ALLOWED4 cannot start with OPENCOLLECTOR_CONFIG as it is reserved\n")
	assert.Contains(t, err.Error(), "environment variable _NOT_ALLOWED5 cannot start with _ as it is reserved\n")
}
