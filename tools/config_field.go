package tools

import (
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

type FieldConfigs []*FieldConfig

// applyOnTools will apply all user field configs to the response fields relevant within the given tools.
func (f FieldConfigs) applyOnTools(tools []*toolsproto.Tool) {
	for _, t := range tools {
		f.applyOnTool(t)
	}
}

func (f FieldConfigs) applyOnTool(t *toolsproto.Tool) {
	// skip if the tool is not action based
	if !t.IsActionBased() {
		return
	}

	for _, response := range t.GetActionConfig().GetResponse() {
		// if this is a model field
		if modelName := response.GetModelName(); modelName != "" {
			// .. and we have a field config
			if fieldCfg := f.find(modelName + "." + response.GetFieldName()); fieldCfg != nil {
				// .. apply it on the response
				fieldCfg.applyOnResponseField(response)
			}
		}
	}

}

// haveChanges checks if the fields have any config changes compared to the generated one.
func (f FieldConfigs) haveChanges() bool {
	for _, c := range f {
		if c.hasChanges() {
			return true
		}
	}

	return false
}

// find returns the field with the given id if any.
func (f FieldConfigs) find(id string) *FieldConfig {
	for _, field := range f {
		if field.ID == id {
			return field
		}
	}

	return nil
}

// changed returns a subset of FieldConfigs including just the fields with changes.
func (f FieldConfigs) changed() FieldConfigs {
	changed := FieldConfigs{}
	for _, c := range f {
		if c.hasChanges() {
			changed = append(changed, c)
		}
	}

	return changed
}

func (f FieldConfigs) applyOn(fields []*toolsproto.Field) {
	for _, cfg := range f.changed() {
		for _, field := range fields {
			if cfg.ID == field.GetID() {
				cfg.applyOn(field)
			}
		}
	}
}

type FieldConfig struct {
	ID           string        `json:"id"`
	Format       *FormatConfig `json:"format,omitempty"`
	DisplayName  *string       `json:"display_name,omitempty"`
	Visible      *bool         `json:"visible,omitempty"`
	HelpText     *string       `json:"help_text,omitempty"`
	ImagePreview *bool         `json:"image_preview,omitempty"`
}

func (f *FieldConfig) hasChanges() bool {
	return f.Format != nil && f.Format.hasChanges() ||
		f.DisplayName != nil ||
		f.Visible != nil ||
		f.HelpText != nil ||
		f.ImagePreview != nil
}

func (f *FieldConfig) applyOn(field *toolsproto.Field) {
	if field == nil {
		return
	}

	if f.Format != nil {
		field.Format = f.Format.applyOn(field.GetFormat())
	}
	if f.DisplayName != nil {
		field.DisplayName = f.DisplayName
	}
	if f.Visible != nil {
		field.Visible = f.Visible
	}
	if f.HelpText != nil {
		field.HelpText = makeStringTemplate(f.HelpText)
	}
	if f.ImagePreview != nil {
		field.ImagePreview = f.ImagePreview
	}
}

// applyOnResponseField will apply this Field configuration onto a tool's ResponseFieldConfig.
func (f *FieldConfig) applyOnResponseField(response *toolsproto.ResponseFieldConfig) {
	if response == nil {
		return
	}

	if f.Format != nil {
		response.Format = f.Format.applyOn(response.GetFormat())
	}
	if f.DisplayName != nil {
		response.DisplayName = *f.DisplayName
	}
	if f.Visible != nil {
		response.Visible = *f.Visible
	}
	if f.HelpText != nil {
		response.HelpText = makeStringTemplate(f.HelpText)
	}
	if f.ImagePreview != nil {
		response.ImagePreview = *f.ImagePreview
	}
}

type FormatConfig struct {
	EnumConfig   *EnumFormatConfig   `json:"enum_config,omitempty"`
	NumberConfig *NumberFormatConfig `json:"number_config,omitempty"`
	StringConfig *StringFormatConfig `json:"string_config,omitempty"`
	BoolConfig   *BoolFormatConfig   `json:"bool_config,omitempty"`
}

func (f *FormatConfig) GetType() toolsproto.FormatConfig_Type {
	switch {
	case f.EnumConfig != nil:
		return toolsproto.FormatConfig_ENUM
	case f.NumberConfig != nil:
		return toolsproto.FormatConfig_NUMBER
	case f.StringConfig != nil:
		return toolsproto.FormatConfig_STRING
	case f.BoolConfig != nil:
		return toolsproto.FormatConfig_BOOL
	default:
		return toolsproto.FormatConfig_UNKNOWN
	}
}

func (f *FormatConfig) hasChanges() bool {
	return (f.EnumConfig != nil && f.EnumConfig.hasChanges()) ||
		(f.NumberConfig != nil && f.NumberConfig.hasChanges()) ||
		(f.StringConfig != nil && f.StringConfig.hasChanges()) ||
		(f.BoolConfig != nil && f.BoolConfig.hasChanges())
}

func (f *FormatConfig) applyOn(cfg *toolsproto.FormatConfig) *toolsproto.FormatConfig {
	if cfg == nil {
		return &toolsproto.FormatConfig{
			Type:         f.GetType(),
			EnumConfig:   nil, //TODO
			NumberConfig: f.NumberConfig.applyOn(nil),
			StringConfig: f.StringConfig.applyOn(nil),
			BoolConfig:   f.BoolConfig.applyOn(nil),
		}
	}
	// TODO: implement
	// if f.EnumConfig != nil {
	// cfg.EnumConfig = f.EnumConfig.applyOn(cfg.EnumConfig)
	// }

	if f.NumberConfig != nil {
		cfg.NumberConfig = f.NumberConfig.applyOn(cfg.GetNumberConfig())
	}
	if f.StringConfig != nil {
		cfg.StringConfig = f.StringConfig.applyOn(cfg.GetStringConfig())
	}
	if f.BoolConfig != nil {
		cfg.BoolConfig = f.BoolConfig.applyOn(cfg.GetBoolConfig())
	}

	return cfg
}

type EnumFormatConfig struct {
	//TODO:
}

func (e *EnumFormatConfig) hasChanges() bool {
	// TODO:
	return false
}

type NumberFormatConfig struct {
	Prefix       *string `json:"prefix,omitempty"`
	Suffix       *string `json:"suffix,omitempty"`
	Sensitive    *bool   `json:"sensitive,omitempty"` // hidden by default, hover to show
	Mode         *string `json:"mode,omitempty"`
	CurrencyCode *string `json:"currency_code,omitempty"`
	UnitCode     *string `json:"unit_code,omitempty"`
	Locale       *string `json:"locale,omitempty"`
}

func (n *NumberFormatConfig) hasChanges() bool {
	return n.Prefix != nil ||
		n.Suffix != nil ||
		n.Sensitive != nil ||
		n.Mode != nil ||
		n.CurrencyCode != nil ||
		n.UnitCode != nil ||
		n.Locale != nil
}

func (n *NumberFormatConfig) applyOn(cfg *toolsproto.NumberFormatConfig) *toolsproto.NumberFormatConfig {
	if n == nil {
		return cfg
	}

	if cfg == nil {
		return &toolsproto.NumberFormatConfig{
			Prefix:    n.Prefix,
			Suffix:    n.Suffix,
			Sensitive: n.Sensitive,
			Mode: func() toolsproto.NumberFormatConfig_Mode {
				if n.Mode == nil {
					return toolsproto.NumberFormatConfig_DECIMAL
				}
				return toolsproto.NumberFormatConfig_Mode(toolsproto.NumberFormatConfig_Mode_value[*n.Mode])
			}(),
			CurrencyCode: n.CurrencyCode,
			UnitCode:     n.UnitCode,
			Locale:       n.Locale,
		}
	}

	if n.Prefix != nil {
		cfg.Prefix = n.Prefix
	}
	if n.Suffix != nil {
		cfg.Suffix = n.Suffix
	}
	if n.Sensitive != nil {
		cfg.Sensitive = n.Sensitive
	}
	if n.Mode != nil {
		cfg.Mode = toolsproto.NumberFormatConfig_Mode(toolsproto.NumberFormatConfig_Mode_value[*n.Mode])
	}
	if n.CurrencyCode != nil {
		cfg.CurrencyCode = n.CurrencyCode
	}
	if n.UnitCode != nil {
		cfg.UnitCode = n.UnitCode
	}
	if n.Locale != nil {
		cfg.Locale = n.Locale
	}

	return cfg
}

type StringFormatConfig struct {
	Prefix         *string `json:"prefix,omitempty"`
	Suffix         *string `json:"suffix,omitempty"`
	ShowURLPreview *bool   `json:"show_url_preview,omitempty"`
	Sensitive      *bool   `json:"sensitive,omitempty"` // hidden by default, hover to show
	TextColour     *string `json:"text_colour,omitempty"`
}

func (s *StringFormatConfig) hasChanges() bool {
	return s.Prefix != nil ||
		s.Suffix != nil ||
		s.Sensitive != nil ||
		s.TextColour != nil ||
		s.ShowURLPreview != nil
}

func (s *StringFormatConfig) applyOn(cfg *toolsproto.StringFormatConfig) *toolsproto.StringFormatConfig {
	if s == nil {
		return cfg
	}

	if cfg == nil {
		return &toolsproto.StringFormatConfig{
			Prefix:         s.Prefix,
			Suffix:         s.Suffix,
			ShowUrlPreview: s.ShowURLPreview,
			Sensitive:      s.Sensitive,
			TextColour:     s.TextColour,
		}
	}

	if s.Prefix != nil {
		cfg.Prefix = s.Prefix
	}
	if s.Suffix != nil {
		cfg.Suffix = s.Suffix
	}
	if s.ShowURLPreview != nil {
		cfg.ShowUrlPreview = s.ShowURLPreview
	}
	if s.Sensitive != nil {
		cfg.Sensitive = s.Sensitive
	}
	if s.TextColour != nil {
		cfg.TextColour = s.TextColour
	}

	return cfg
}

type BoolFormatConfig struct {
	PositiveColour *string `json:"positive_colour,omitempty"`
	PositiveValue  *string `json:"positive_value,omitempty"`
	NegativeColour *string `json:"negative_colour,omitempty"`
	NegativeValue  *string `json:"negative_value,omitempty"`
}

func (b *BoolFormatConfig) hasChanges() bool {
	return b.PositiveColour != nil ||
		b.PositiveValue != nil ||
		b.NegativeColour != nil ||
		b.NegativeValue != nil
}

func (b *BoolFormatConfig) applyOn(cfg *toolsproto.BoolFormatConfig) *toolsproto.BoolFormatConfig {
	if b == nil {
		return cfg
	}

	if cfg == nil {
		return &toolsproto.BoolFormatConfig{
			PositiveColour: b.PositiveColour,
			PositiveValue:  b.PositiveValue,
			NegativeColour: b.NegativeColour,
			NegativeValue:  b.NegativeValue,
		}
	}

	if b.PositiveColour != nil {
		cfg.PositiveColour = b.PositiveColour
	}
	if b.PositiveValue != nil {
		cfg.PositiveValue = b.PositiveValue
	}
	if b.NegativeColour != nil {
		cfg.NegativeColour = b.NegativeColour
	}
	if b.NegativeValue != nil {
		cfg.NegativeValue = b.NegativeValue
	}

	return cfg
}
