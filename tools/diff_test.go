package tools

import (
	"reflect"
	"testing"

	toolsproto "github.com/teamkeel/keel/tools/proto"
)

var emptyStr = ""

func Test_diffString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		old     string
		updated string
		want    *string
	}{
		{
			name:    "no change",
			old:     "foo",
			updated: "foo",
			want:    nil,
		},
		{
			name:    "change value",
			old:     "foo",
			updated: "bar",
			want:    stringPointer("bar"),
		},
		{
			name:    "change from empty",
			old:     "",
			updated: "bar",
			want:    stringPointer("bar"),
		},
		{
			name:    "change from value to empty",
			old:     "bar",
			updated: "",
			want:    &emptyStr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := diffString(tt.old, tt.updated)
			if (got == nil) != (tt.want == nil) || (got != nil && tt.want != nil && *got != *tt.want) {
				t.Errorf("diffString() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func Test_diffStringTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		old     *toolsproto.StringTemplate
		updated *toolsproto.StringTemplate
		want    *string
	}{
		{
			name:    "both nil",
			old:     nil,
			updated: nil,
			want:    nil,
		},
		{
			name:    "old nil, updated non-nil",
			old:     nil,
			updated: &toolsproto.StringTemplate{Template: "foo"},
			want:    stringPointer("foo"),
		},
		{
			name:    "old non-nil, updated nil",
			old:     &toolsproto.StringTemplate{Template: "foo"},
			updated: nil,
			want:    nil,
		},
		{
			name:    "same template value",
			old:     &toolsproto.StringTemplate{Template: "foo"},
			updated: &toolsproto.StringTemplate{Template: "foo"},
			want:    nil,
		},
		{
			name:    "different template value",
			old:     &toolsproto.StringTemplate{Template: "foo"},
			updated: &toolsproto.StringTemplate{Template: "bar"},
			want:    stringPointer("bar"),
		},
		{
			name:    "old empty, updated non-empty",
			old:     &toolsproto.StringTemplate{Template: ""},
			updated: &toolsproto.StringTemplate{Template: "bar"},
			want:    stringPointer("bar"),
		},
		{
			name:    "old non-empty, updated empty",
			old:     &toolsproto.StringTemplate{Template: "bar"},
			updated: &toolsproto.StringTemplate{Template: ""},
			want:    &emptyStr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := diffStringTemplate(tt.old, tt.updated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("diffStringTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractFieldConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		generated *toolsproto.Field
		updated   *toolsproto.Field
		want      *FieldConfig
	}{
		{
			name: "no changes",
			generated: &toolsproto.Field{
				DisplayName:  stringPointer("old name"),
				ModelName:    stringPointer("model"),
				FieldName:    stringPointer("field"),
				Visible:      boolPointer(false),
				ImagePreview: boolPointer(true),
				HelpText:     &toolsproto.StringTemplate{Template: "old help"},
			},
			updated: &toolsproto.Field{
				DisplayName:  stringPointer("old name"),
				ModelName:    stringPointer("model"),
				FieldName:    stringPointer("field"),
				Visible:      boolPointer(false),
				ImagePreview: boolPointer(true),
				HelpText:     &toolsproto.StringTemplate{Template: "old help"},
			},
			want: nil,
		},
		{
			name: "all changed",
			generated: &toolsproto.Field{
				DisplayName:  stringPointer("old name"),
				ModelName:    stringPointer("model"),
				FieldName:    stringPointer("field"),
				Visible:      boolPointer(false),
				ImagePreview: boolPointer(true),
				HelpText:     &toolsproto.StringTemplate{Template: "old help"},
			},
			updated: &toolsproto.Field{
				DisplayName:  stringPointer("new name"),
				ModelName:    stringPointer("model"),
				FieldName:    stringPointer("field"),
				Visible:      boolPointer(true),
				ImagePreview: boolPointer(false),
				HelpText:     &toolsproto.StringTemplate{Template: "new help"},
			},
			want: &FieldConfig{
				ID:           "model.field",
				DisplayName:  stringPointer("new name"),
				Visible:      boolPointer(true),
				ImagePreview: boolPointer(false),
				HelpText:     stringPointer("new help"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := extractFieldConfig(tt.generated, tt.updated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractFieldConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_extractNumberFormatConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		generated *toolsproto.NumberFormatConfig
		updated   *toolsproto.NumberFormatConfig
		want      *NumberFormatConfig
	}{
		{
			name: "updated fields",
			generated: &toolsproto.NumberFormatConfig{
				Mode:         toolsproto.NumberFormatConfig_DECIMAL,
				CurrencyCode: stringPointer("GBP"),
				UnitCode:     stringPointer("CODE"),
				Sensitive:    boolPointer(false),
				Locale:       stringPointer("UK"),
				Prefix:       stringPointer("£"),
				Suffix:       stringPointer("pcm"),
			},
			updated: &toolsproto.NumberFormatConfig{
				Mode:         toolsproto.NumberFormatConfig_CURRENCY,
				CurrencyCode: stringPointer("USD"),
				UnitCode:     stringPointer("NEW CODE"),
				Sensitive:    boolPointer(true),
				Locale:       stringPointer("US"),
				Prefix:       stringPointer("$"),
				Suffix:       stringPointer("p/w"),
			},
			want: &NumberFormatConfig{
				Mode:         stringPointer("CURRENCY"),
				CurrencyCode: stringPointer("USD"),
				UnitCode:     stringPointer("NEW CODE"),
				Sensitive:    boolPointer(true),
				Locale:       stringPointer("US"),
				Prefix:       stringPointer("$"),
				Suffix:       stringPointer("p/w"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := extractNumberFormatConfig(tt.generated, tt.updated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractNumberFormatConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractStringFormatConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		generated *toolsproto.StringFormatConfig
		updated   *toolsproto.StringFormatConfig
		want      *StringFormatConfig
	}{
		{
			name: "all fields changed",
			generated: &toolsproto.StringFormatConfig{
				Prefix:         stringPointer("oldPrefix"),
				Suffix:         stringPointer("oldSuffix"),
				ShowUrlPreview: boolPointer(true),
				Sensitive:      boolPointer(false),
				TextColour:     stringPointer("red"),
			},
			updated: &toolsproto.StringFormatConfig{
				Prefix:         stringPointer("newPrefix"),
				Suffix:         stringPointer("newSuffix"),
				ShowUrlPreview: boolPointer(false),
				Sensitive:      boolPointer(true),
				TextColour:     stringPointer("blue"),
			},
			want: &StringFormatConfig{
				Prefix:         stringPointer("newPrefix"),
				Suffix:         stringPointer("newSuffix"),
				ShowURLPreview: boolPointer(false),
				Sensitive:      boolPointer(true),
				TextColour:     stringPointer("blue"),
			},
		},
		{
			name:      "generated is nil, updated has values",
			generated: nil,
			updated: &toolsproto.StringFormatConfig{
				Prefix:         stringPointer("prefix"),
				Suffix:         stringPointer("suffix"),
				ShowUrlPreview: boolPointer(true),
				Sensitive:      boolPointer(true),
				TextColour:     stringPointer("green"),
			},
			want: &StringFormatConfig{
				Prefix:         stringPointer("prefix"),
				Suffix:         stringPointer("suffix"),
				ShowURLPreview: boolPointer(true),
				Sensitive:      boolPointer(true),
				TextColour:     stringPointer("green"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := extractStringFormatConfig(tt.generated, tt.updated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractStringFormatConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_extractBoolFormatConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		generated *toolsproto.BoolFormatConfig
		updated   *toolsproto.BoolFormatConfig
		want      *BoolFormatConfig
	}{
		{
			name: "all fields changed",
			generated: &toolsproto.BoolFormatConfig{
				PositiveValue:  stringPointer("Yes"),
				NegativeValue:  stringPointer("No"),
				PositiveColour: stringPointer("green"),
				NegativeColour: stringPointer("red"),
			},
			updated: &toolsproto.BoolFormatConfig{
				PositiveValue:  stringPointer("True"),
				NegativeValue:  stringPointer("False"),
				PositiveColour: stringPointer("blue"),
				NegativeColour: stringPointer("yellow"),
			},
			want: &BoolFormatConfig{
				PositiveValue:  stringPointer("True"),
				NegativeValue:  stringPointer("False"),
				PositiveColour: stringPointer("blue"),
				NegativeColour: stringPointer("yellow"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := extractBoolFormatConfig(tt.generated, tt.updated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractBoolFormatConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
