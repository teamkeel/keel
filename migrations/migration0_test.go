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

	fmt.Printf("XXXX generated SQL so far...\n\n%s\n", m0.SQL)
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
