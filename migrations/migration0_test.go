package migrations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestItCompiles(t *testing.T) {
	m0 := NewMigration0(&referenceSchema)
	m0.GenerateSQL()
	require.True(t, len(m0.SQL) > 0)

	fmt.Println()
	for _, statement := range m0.SQL {
		fmt.Printf("%s\n", statement)
	}
}

var referenceSchema proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "ModelA",
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
	},
}
