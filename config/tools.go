package config

type ToolsConfig struct {
	UseApi *string `yaml:"useApi,omitempty"`
}

// AccessTokenExpiry retrieves the configured or default access token expiry
func (c *ToolsConfig) ToolsApi() string {
	if c.UseApi != nil {
		return *c.UseApi
	} else {
		return "api"
	}
}
