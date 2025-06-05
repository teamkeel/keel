package proto

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestFindModel(t *testing.T) {
	t.Parallel()
	require.Equal(t, "ModelA", FindModel(referenceSchema.GetModels(), "ModelA").GetName())
}

func TestFindModels(t *testing.T) {
	t.Parallel()
	modelsFound := FindModels(referenceSchema.GetModels(), []string{"ModelA", "ModelC"})
	namesOfFoundModels := lo.Map(modelsFound, func(m *Model, _ int) string {
		return m.GetName()
	})
	require.Equal(t, []string{"ModelA", "ModelC"}, namesOfFoundModels)
}

func TestFindField(t *testing.T) {
	t.Parallel()
	require.Equal(t, "Field2", FindField(referenceSchema.GetModels(), "ModelA", "Field2").GetName())
}

func TestModelExists(t *testing.T) {
	t.Parallel()
	require.True(t, ModelExists(referenceSchema.GetModels(), "ModelA"))
	require.False(t, ModelExists(referenceSchema.GetModels(), "ModelZ"))
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
