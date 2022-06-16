package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestCreateTable(t *testing.T) {
	require.Equal(t, expectedCreateTable, createTable(exampleModel))
}

func TestCreateTableIfNotExists(t *testing.T) {
	fields := []*proto.Field{
		{
			Name: "my-field-name",
			Type: proto.FieldType_FIELD_TYPE_BOOL,
		},
	}
	require.Equal(t, expectedCreateTableIfNotExists, CreateTableIfNotExists("my-model-name", fields))
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

func TestInsertRowComprisingSingleString(t *testing.T) {
	require.Equal(t, expectedInsertRowComprisingSingleString, InsertRowComprisingSingleString("my-table", "my string value"))
}

func TestUpdateSingleStringColumn(t *testing.T) {
	require.Equal(t, expectedTestUpdateSingleStringColumn, UpdateSingleStringColumn("my-table", "my_column", "my string value"))
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

func TestSelectSingleColumn(t *testing.T) {
	require.Equal(t, expectedSingleColumn, SelectSingleColumn("my-table", "some-column"))
}

const expectedCreateTable string = `CREATE TABLE "Person"(
"Name" TEXT,
"Age" integer
);`

const expectedCreateTableIfNotExists string = `CREATE TABLE if not exists "my-model-name"(
"my-field-name" bool
);`

const expectedDropTable string = `DROP TABLE "Person";`

const expectedCreateField string = `ALTER TABLE "Person"
ADD "myNewField" DATE;`

const expectedDropField string = `ALTER TABLE "Person"
DROP "Age";`

const expectedSingleColumn string = `SELECT "some-column" FROM "my-table";`

const expectedInsertRowComprisingSingleString = `INSERT INTO "my-table"
VALUES ('my string value');`

const expectedTestUpdateSingleStringColumn string = `UPDATE "my-table" SET "my_column"='my string value';`
