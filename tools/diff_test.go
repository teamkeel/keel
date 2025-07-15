package tools

import (
	"reflect"
	"testing"

	toolsproto "github.com/teamkeel/keel/tools/proto"
)

var emptyStr = ""

func Test_diffString(t *testing.T) {
	type args struct {
		old     string
		updated string
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{
			name: "no change",
			args: args{
				old:     "foo",
				updated: "foo",
			},
			want: nil,
		},
		{
			name: "change value",
			args: args{
				old:     "foo",
				updated: "bar",
			},
			want: stringPointer("bar"),
		},
		{
			name: "change from empty",
			args: args{
				old:     "",
				updated: "bar",
			},
			want: stringPointer("bar"),
		},
		{
			name: "change from value to empty",
			args: args{
				old:     "bar",
				updated: "",
			},
			want: &emptyStr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffString(tt.args.old, tt.args.updated)
			if (got == nil) != (tt.want == nil) || (got != nil && tt.want != nil && *got != *tt.want) {
				t.Errorf("diffString() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func Test_diffStringTemplate(t *testing.T) {
	type args struct {
		old     *toolsproto.StringTemplate
		updated *toolsproto.StringTemplate
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{
			name: "both nil",
			args: args{
				old:     nil,
				updated: nil,
			},
			want: nil,
		},
		{
			name: "old nil, updated non-nil",
			args: args{
				old:     nil,
				updated: &toolsproto.StringTemplate{Template: "foo"},
			},
			want: stringPointer("foo"),
		},
		{
			name: "old non-nil, updated nil",
			args: args{
				old:     &toolsproto.StringTemplate{Template: "foo"},
				updated: nil,
			},
			want: nil,
		},
		{
			name: "same template value",
			args: args{
				old:     &toolsproto.StringTemplate{Template: "foo"},
				updated: &toolsproto.StringTemplate{Template: "foo"},
			},
			want: nil,
		},
		{
			name: "different template value",
			args: args{
				old:     &toolsproto.StringTemplate{Template: "foo"},
				updated: &toolsproto.StringTemplate{Template: "bar"},
			},
			want: stringPointer("bar"),
		},
		{
			name: "old empty, updated non-empty",
			args: args{
				old:     &toolsproto.StringTemplate{Template: ""},
				updated: &toolsproto.StringTemplate{Template: "bar"},
			},
			want: stringPointer("bar"),
		},
		{
			name: "old non-empty, updated empty",
			args: args{
				old:     &toolsproto.StringTemplate{Template: "bar"},
				updated: &toolsproto.StringTemplate{Template: ""},
			},
			want: &emptyStr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := diffStringTemplate(tt.args.old, tt.args.updated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("diffStringTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractFieldConfig(t *testing.T) {
	tests := []struct {
		name      string
		generated *toolsproto.Field
		updated   *toolsproto.Field
		want      *FieldConfig
	}{
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
			if got := extractFieldConfig(tt.generated, tt.updated); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractFieldConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
