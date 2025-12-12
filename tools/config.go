package tools

import toolsproto "github.com/teamkeel/keel/tools/proto"

// UserConfig represents all user configurations; this includes tool configurations and model fields.
type UserConfig struct {
	Tools  ToolConfigs
	Fields FieldConfigs
	Spaces SpaceConfigs
}

func makeStringTemplate(tmpl *string) *toolsproto.StringTemplate {
	if tmpl != nil {
		return &toolsproto.StringTemplate{Template: *tmpl}
	}

	return nil
}
