package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestModelsDroppedOrAdded(t *testing.T) {
	differ := NewProtoDiffer(modelsABC, modelsBCD)
	diffs, err := differ.Analyse()
	require.NoError(t, err)
	require.Equal(t, []string{"D"}, diffs.ModelsAdded)
	require.Equal(t, []string{"A"}, diffs.ModelsRemoved)
}

var modelsABC *proto.Schema = &proto.Schema{
	Models: []*proto.Model{
		{
			Name: "A",
		},
		{
			Name: "B",
		},
		{
			Name: "C",
		},
	},
}

var modelsBCD *proto.Schema = &proto.Schema{
	Models: []*proto.Model{
		{
			Name: "B",
		},
		{
			Name: "C",
		},
		{
			Name: "D",
		},
	},
}
