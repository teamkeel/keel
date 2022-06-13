package migrations

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestCreateTable(t *testing.T) {
	output := createTable(exampleModel)
	require.True(t, len(output) > 20)
	if os.Getenv("DEBUG") != "" {
		t.Logf("\n\n%s\n\n", output)
	}
	require.Equal(t, expectedCreateTable, output)
}

var exampleModel *proto.Model = &proto.Model{
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
}

const expectedCreateTable string = `CREATE TABLE Person(
Name TEXT,
Age integer
);`
