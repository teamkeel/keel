package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestCreateTable(t *testing.T) {
	require.Equal(t, expectedCreateTable, createTable(exampleModel))
}

func TestDropTable(t *testing.T) {
	require.Equal(t, expectedDropTable, dropTable("Person"))
}

func TestCreateField(t *testing.T) {
	require.Equal(t, expectedCreateField, createField(
		exampleModel.Name,
		&proto.Field{
			Name: "myNewField",
			Type: proto.FieldType_FIELD_TYPE_DATE,
		}))
}

func TestDropField(t *testing.T) {
	require.Equal(t, expectedDropField, dropField("Person", "Age"))
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

const expectedCreateTable string = `CREATE TABLE "Person"(
"Name" TEXT,
"Age" integer
);`

const expectedDropTable string = `DROP TABLE "Person";`

const expectedCreateField string = `ALTER TABLE "Person"
ADD "myNewField" DATE;`

const expectedDropField string = `ALTER TABLE "Person"
DROP "Age";`
