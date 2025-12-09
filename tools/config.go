package tools

// UserConfig represents all user configurations; this includes tool configurations and model fields.
type UserConfig struct {
	Tools  ToolConfigs
	Fields FieldConfigs
	Spaces SpaceConfigs
}
