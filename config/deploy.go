package config

type DeployConfig struct {
	ProjectName string           `yaml:"projectName"`
	Region      string           `yaml:"region"`
	Database    *DatabaseConfig  `yaml:"database,omitempty"`
	Jobs        *JobsConfig      `yaml:"jobs,omitempty"`
	Telemetry   *TelemetryConfig `yaml:"telemetry,omitempty"`
}

type DatabaseConfig struct {
	Provider string     `yaml:"provider"`
	RDS      *RDSConfig `yaml:"rds,omitempty"`
}

type RDSConfig struct {
	Instance *string `yaml:"instance,omitempty"`
	MultiAZ  *bool   `yaml:"multiAz,omitempty"`
	Storage  *int    `yaml:"storage,omitempty"`
}

type JobsConfig struct {
	WebhookURL string `yaml:"webhookUrl,omitempty"`
}

type TelemetryConfig struct {
	Collector string `yaml:"collector,omitempty"`
}

func validateDatabase(c *ProjectConfig) []*ConfigError {
	errors := []*ConfigError{}

	if c.Deploy == nil {
		return errors
	}

	if c.Deploy.Database == nil {
		return errors
	}

	db := c.Deploy.Database

	if db.Provider != "rds" && db.RDS != nil {
		errors = append(errors, &ConfigError{
			Message: "deploy.database.rds: can only be provided if deploy.database.provider is 'rds'",
			Field:   "deploy.database.rds",
			Type:    "invalid-property",
		})
	}

	return errors
}
