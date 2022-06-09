package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestFirstBabySteps(t *testing.T) {
	output := GenerateAllTables(referenceSchema.Models)
	require.True(t, len(output) > 20)
	// todo put in proper assertions here
	// - I tried using expected output string constants - but vscode auto
	// indents them which break the tests.
}

var referenceSchema proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "Person",
			Fields: []*proto.Field{
				{
					Name: "Name",
					Type: proto.FieldType_FIELD_TYPE_STRING,
				},
				{
					Name: "Age",
					Type: proto.FieldType_FIELD_TYPE_INT,
				},
			},
		},
		{
			Name: "Vehicle",
			Fields: []*proto.Field{
				{
					Name: "Make",
					Type: proto.FieldType_FIELD_TYPE_STRING,
				},
				{
					Name: "PriceNew",
					Type: proto.FieldType_FIELD_TYPE_CURRENCY,
				},
			},
		},
	},
}
