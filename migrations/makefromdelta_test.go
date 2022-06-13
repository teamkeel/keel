package migrations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestYouGetOneCreatedTableAndOneDroppedTableIfYouChangeAModelName(t *testing.T) {
	// The only difference between these two schemas are that the name
	// of one of the models has changed. So we should get a new table created,
	// and one table dropped.
	generatedSQL, err := MakeMigrationsFromSchemaDifference(&oldProto, &newProto)
	fmt.Printf("TestChangedMode generated SQL...\n\n%s\n\n", generatedSQL)
	require.NoError(t, err)
	require.Equal(t, expectedChangedNameSQL, generatedSQL)
}

var oldProto proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "Person",
			Fields: []*proto.Field{
				{
					Name: "Name",
					Type: proto.FieldType_FIELD_TYPE_STRING,
				},
			},
		},
	},
}

var newProto proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "Human",
			Fields: []*proto.Field{
				{
					Name: "Name",
					Type: proto.FieldType_FIELD_TYPE_STRING,
				},
			},
		},
	},
}

const expectedChangedNameSQL string = `
CREATE TABLE Human(
Name TEXT
);
DROP TABLE Person;`
