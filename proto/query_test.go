package proto

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestModelNames(t *testing.T) {
	require.Equal(t, []string{"ModelA", "ModelB"}, ModelNames(referenceSchema))
}

func TestFieldNames(t *testing.T) {
	require.Equal(t, []string{"Field1", "Field2"}, FieldNames(referenceSchema.Models[0]))
}

func TestFindModel(t *testing.T) {
	require.Equal(t, "ModelA", FindModel(referenceSchema.Models, "ModelA").Name)
}

func TestFindModels(t *testing.T) {
	modelsFound := FindModels(referenceSchema.Models, []string{"ModelA", "ModelC"})
	namesOfFoundModels := lo.Map(modelsFound, func(m *Model, _ int) string {
		return m.Name
	})
	require.Equal(t, []string{"ModelA", "ModelC"}, namesOfFoundModels)
}

func TestFindField(t *testing.T) {
	require.Equal(t, "Field2", FindField(referenceSchema.Models, "ModelA", "Field2").Name)
}

func TestModelExists(t *testing.T) {
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
