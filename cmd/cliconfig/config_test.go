package cliconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestSetAndUpdateConfig(t *testing.T) {
	filePath := "set_config.yaml"
	c := New(&Options{Path: filePath})
	cfg, err := c.GetConfig()
	assert.NoError(t, err)

	configFile, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var expected UserConfig
	err = yaml.Unmarshal(configFile, &expected)
	assert.NoError(t, err)

	assert.Equal(t, &expected, cfg)

	err = c.SetSecret("staging", "TEST_API_KEY", "test_secret")
	assert.NoError(t, err)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	cfg, err = c.GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, "test_secret", cfg.Projects[wd].Secrets.Staging["TEST_API_KEY"])

	err = c.SetSecret("staging", "TEST_API_KEY", "updated")
	assert.NoError(t, err)
	err = c.SetSecret("staging", "TEST_API_KEY_2", "updated2")
	assert.NoError(t, err)

	cfg, err = c.GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, "updated", cfg.Projects[wd].Secrets.Staging["TEST_API_KEY"])

	err = os.Remove(filePath)
	assert.NoError(t, err)
}

func TestGetConfig(t *testing.T) {
	filePath := "get_config.yaml"
	c := New(&Options{Path: filePath})
	cfg, err := c.GetConfig()
	assert.NoError(t, err)

	configFile, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var expected UserConfig
	err = yaml.Unmarshal(configFile, &expected)
	assert.NoError(t, err)

	assert.Equal(t, &expected, cfg)

	err = os.Remove(filePath)
	assert.NoError(t, err)
}
