package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestModelsDroppedOrAdded(t *testing.T) {
	diffs, err := ProtoDeltas(modelsAB, modelsBC)
	require.NoError(t, err)
	require.Equal(t, []string{"ModelC"}, diffs.ModelsAdded)
	require.Equal(t, []string{"ModelA"}, diffs.ModelsRemoved)
}

func TestFieldsDroppedOrAdded(t *testing.T) {
	diffs, err := ProtoDeltas(fieldsAB, fieldsBC)
	require.NoError(t, err)
	require.Equal(t, []string{"FieldC"}, diffs.FieldsAdded["ModelA"])
	require.Equal(t, []string{"FieldA"}, diffs.FieldsRemoved["ModelA"])
}

var modelsAB *proto.Schema = &proto.Schema{
	Models: []*proto.Model{
		{
			Name: "ModelA",
		},
		{
			Name: "ModelB",
		},
	},
}

var modelsBC *proto.Schema = &proto.Schema{
	Models: []*proto.Model{
		{
			Name: "ModelB",
		},
		{
			Name: "ModelC",
		},
	},
}

var fieldsAB *proto.Schema = &proto.Schema{
	Models: []*proto.Model{
		{
			Name: "ModelA",
			Fields: []*proto.Field{
				{Name: "FieldA"},
				{Name: "FieldB"},
			},
		},
	},
}

var fieldsBC *proto.Schema = &proto.Schema{
	Models: []*proto.Model{
		{
			Name: "ModelA",
			Fields: []*proto.Field{
				{Name: "FieldB"},
				{Name: "FieldC"},
			},
		},
	},
}
