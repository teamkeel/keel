package protoqry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestModelNames(t *testing.T) {
	require.Equal(t, []string{"ModelA", "ModelB"}, ModelNames(&referenceSchema))
}

func TestFieldNames(t *testing.T) {
	require.Equal(t, []string{"Field1", "Field2"}, FieldNames(referenceSchema.Models[0]))
}

func TestFindModel(t *testing.T) {
	require.Equal(t, "ModelA", FindModel(referenceSchema.Models, "ModelA").Name)
}

func TestFindField(t *testing.T) {
	require.Equal(t, "Field2", FindField(referenceSchema.Models, "ModelA", "Field2").Name)
}

func TestModelExists(t *testing.T) {
	require.True(t, ModelExists(referenceSchema.Models, "ModelA"))
	require.False(t, ModelExists(referenceSchema.Models, "ModelZ"))
}

var referenceSchema proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "ModelA",
			Fields: []*proto.Field{
				{Name: "Field1"},
				{Name: "Field2"},
			},
		},
		{
			Name: "ModelB",
			Fields: []*proto.Field{
				{Name: "Field2"},
				{Name: "Field3"},
			},
		},
	},
}
