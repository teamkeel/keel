package config

type ConsoleConfig struct {
	Api *string `yaml:"api,omitempty"`
}

// AccessTokenExpiry retrieves the configured or default access token expiry
func (c *ConsoleConfig) ToolsApi() string {
	if c.Api != nil {
		return *c.Api
	} else {
		return "api"
	}
}
