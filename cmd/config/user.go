package config

func (c *Config) SetUserConfig(userConfig *UserConfig) error {
	var config *RootConfig

	config, err := c.GetConfig()
	if err != nil {
		config = &RootConfig{}
	}

	config.User = *userConfig

	return c.SetConfig(config)
}
