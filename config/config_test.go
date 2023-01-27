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
