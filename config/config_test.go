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
	_, err := Load("fixtures/test_basic_failing_config.yaml")
	assert.Error(t, err)

	assert.Equal(t, "could not unmarshal config file: yaml: unmarshal errors:\n  line 5: cannot unmarshal !!seq into string", err.Error())
}

func TestDuplicates(t *testing.T) {
	config, err := Load("fixtures/test_basic_config_duplicates.yaml")
	assert.Error(t, err)
	assert.Equal(t, config.Environment.Default[0].Name, "TEST")
	assert.Equal(t, "could not validate config file: duplicate environment variables found in staging: TEST", err.Error())
}

func TestRequiredFail(t *testing.T) {
	config, err := Load("fixtures/test_basic_config_required_fail.yaml")
	assert.Error(t, err)
	assert.Equal(t, config.Environment.Staging[0].Name, "TEST")
	assert.Equal(t, "could not validate config file: missing required environment variables in production: TEST", err.Error())
}

func TestRequired(t *testing.T) {
	config, err := Load("fixtures/test_basic_config_required.yaml")
	assert.NoError(t, err)
	assert.Equal(t, config.Environment.Staging[0].Name, "TEST")
	assert.Equal(t, config.Environment.Production[0].Name, "TEST")
}

func TestRequiredValuesKeys(t *testing.T) {
	config, err := Load("fixtures/test_basic_config_required.yaml")
	assert.NoError(t, err)

	notOk, keys := requiredValuesKeys(config)
	assert.False(t, notOk)
	var nil []string
	assert.Equal(t, nil, keys["staging"])

	config, err = Load("fixtures/test_basic_config_required_fail.yaml")
	assert.Error(t, err)

	notOk, keys = requiredValuesKeys(config)
	assert.True(t, notOk)
	assert.Equal(t, []string{"TEST"}, keys["production"])
}
