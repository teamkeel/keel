package tools

import toolsproto "github.com/teamkeel/keel/tools/proto"

type FieldConfigs []*FieldConfig

// haveChanges checks if the fields have any config changes compared to the generated one.
func (f FieldConfigs) haveChanges() bool {
	for _, c := range f {
		if c.hasChanges() {
			return true
		}
	}

	return false
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
	ID     string
	Format *FormatConfig
}

func (f *FieldConfig) hasChanges() bool {
	return f.Format != nil
}

func (f *FieldConfig) applyOn(field *toolsproto.Field) {
	if field == nil {
		return
	}

	field.Format = f.Format.applyOn(field.GetFormat())
}

type FormatConfig struct {
	EnumConfig   *EnumFormatConfig
	NumberConfig *NumberFormatConfig
	StringConfig *StringFormatConfig
	BoolConfig   *BoolFormatConfig
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
			NumberConfig: nil, //TODO
			StringConfig: nil, //TODO
			BoolConfig:   f.BoolConfig.applyOn(nil),
		}
	}
	if f.EnumConfig != nil {
		// TODO: Implement applyOn for EnumConfig when ready
		// cfg.EnumConfig = f.EnumConfig.applyOn(cfg.EnumConfig)
	}
	if f.NumberConfig != nil {
		// TODO: Implement applyOn for NumberConfig when ready
		// cfg.NumberConfig = f.NumberConfig.applyOn(cfg.NumberConfig)
	}
	if f.StringConfig != nil {
		// TODO: Implement applyOn for StringConfig when ready
		// cfg.StringConfig = f.StringConfig.applyOn(cfg.StringConfig)
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
	Prefix     *string
	Suffix     *string
	Precision  *int32
	Sensitive  *bool // hidden by default, hover to show
	TextColour *string
}

func (n *NumberFormatConfig) hasChanges() bool {
	return n.Prefix != nil ||
		n.Suffix != nil ||
		n.Precision != nil ||
		n.Sensitive != nil ||
		n.TextColour != nil
}

type StringFormatConfig struct {
	Prefix         *string
	Suffix         *string
	ShowURLPreview *bool
	Sensitive      *bool // hidden by default, hover to show
	TextColour     *string
}

func (s *StringFormatConfig) hasChanges() bool {
	return s.Prefix != nil ||
		s.Suffix != nil ||
		s.Sensitive != nil ||
		s.TextColour != nil ||
		s.ShowURLPreview != nil
}

type BoolFormatConfig struct {
	PositiveColour *string
	PositiveValue  *string
	NegativeColour *string
	NegativeValue  *string
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
