package proto

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestModelNames(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"ModelA", "ModelB", "ModelC"}, ModelNames(referenceSchema))
}

func TestFieldNames(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"Field1", "Field2"}, FieldNames(referenceSchema.Models[0]))
}

func TestFindModel(t *testing.T) {
	t.Parallel()
	require.Equal(t, "ModelA", FindModel(referenceSchema.Models, "ModelA").Name)
}

func TestFindModels(t *testing.T) {
	t.Parallel()
	modelsFound := FindModels(referenceSchema.Models, []string{"ModelA", "ModelC"})
	namesOfFoundModels := lo.Map(modelsFound, func(m *Model, _ int) string {
		return m.Name
	})
	require.Equal(t, []string{"ModelA", "ModelC"}, namesOfFoundModels)
}

func TestFindField(t *testing.T) {
	t.Parallel()
	require.Equal(t, "Field2", FindField(referenceSchema.Models, "ModelA", "Field2").Name)
}

func TestModelExists(t *testing.T) {
	t.Parallel()
	require.True(t, ModelExists(referenceSchema.Models, "ModelA"))
	require.False(t, ModelExists(referenceSchema.Models, "ModelZ"))
}

var referenceSchema *Schema = &Schema{
	Models: []*Model{
		{
			Name: "ModelA",
			Fields: []*Field{
				{Name: "Field1"},
				{Name: "Field2"},
			},
		},
		{
			Name: "ModelB",
			Fields: []*Field{
				{Name: "Field2"},
				{Name: "Field3"},
			},
		},
		{
			Name: "ModelC",
			Fields: []*Field{
				{Name: "Field42"},
				{Name: "Field43"},
			},
		},
	},
}

func TestSchema_HasFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		schema *Schema
		want   bool
	}{
		{
			name: "schema with files",
			schema: &Schema{
				Models: []*Model{
					{
						Name: "Model",
						Fields: []*Field{
							{Name: "field_1"},
							{Name: "image", Type: &TypeInfo{Type: Type_TYPE_INLINE_FILE}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "schema without files",
			schema: &Schema{
				Models: []*Model{
					{
						Name: "Model",
						Fields: []*Field{
							{Name: "field_1"},
							{Name: "image", Type: &TypeInfo{Type: Type_TYPE_STRING}},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.schema.HasFiles(); got != tt.want {
				t.Errorf("Schema.HasFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
