package proto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelNames(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"ModelA", "ModelB", "ModelC"}, referenceSchema.ModelNames())
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
