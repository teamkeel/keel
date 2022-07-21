package migrations

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/lib/pq"
	"github.com/teamkeel/keel/proto"
)

var PostgresFieldTypes map[proto.Type]string = map[proto.Type]string{
	proto.Type_TYPE_ID:        "TEXT",
	proto.Type_TYPE_STRING:    "TEXT",
	proto.Type_TYPE_INT:       "INTEGER",
	proto.Type_TYPE_BOOL:      "BOOL",
	proto.Type_TYPE_TIMESTAMP: "TIMESTAMP",
	proto.Type_TYPE_DATETIME:  "TIMESTAMP",
	proto.Type_TYPE_DATE:      "DATE",
	proto.Type_TYPE_MODEL:     "TEXT", // id of the target
	proto.Type_TYPE_ENUM:      "TEXT",
}

// Identifier converts v into an identifier that can be used
// for table or column names in Postgres. The value is converted
// to snake case and then quoted. The former is done to create
// a more idiomatic postgres schema and the latter is so you
// can have a table name called "select" that would otherwise
// not be allowed as it clashes with the keyword.
func Identifier(v string) string {
	return pq.QuoteIdentifier(strcase.ToSnake(v))
}

func createTableStmt(model *proto.Model) string {
	output := fmt.Sprintf("CREATE TABLE %s (\n", Identifier(model.Name))
	for i, field := range model.Fields {
		output += fieldDefinition(field)
		if i < len(model.Fields)-1 {
			output += ","
		}
		output += "\n"
	}
	output += ");"
	return output
}

func createTableIfNotExistsStmt(name string, fields []*proto.Field) string {
	output := fmt.Sprintf("CREATE TABLE if not exists %s (\n", Identifier(name))
	for i, field := range fields {
		output += fieldDefinition(field)
		if i < len(fields)-1 {
			output += ","
		}
		output += "\n"
	}
	output += ");"
	return output
}

func dropTableStmt(name string) string {
	return fmt.Sprintf("DROP TABLE %s;", Identifier(name))
}

func addColumnStmt(modelName string, field *proto.Field) string {
	output := fmt.Sprintf("ALTER TABLE %s ADD COLUMN ", Identifier(modelName))
	output += fieldDefinition(field) + ";"
	return output
}

func fieldDefinition(field *proto.Field) string {
	output := fmt.Sprintf("%s %s", Identifier(field.Name), PostgresFieldTypes[field.Type.Type])
	if !field.Optional {
		output += " NOT NULL"
	}

	if field.Unique {
		output += " UNIQUE"
	}

	return output
}

func dropColumnStmt(modelName string, fieldName string) string {
	output := fmt.Sprintf("ALTER TABLE %s ", Identifier(modelName))
	output += fmt.Sprintf("DROP COLUMN %s;", Identifier(fieldName))
	return output
}

func SelectSingleColumn(tableName string, columnName string) string {
	return fmt.Sprintf("SELECT \"%s\" FROM \"%s\";", columnName, tableName)
}

func InsertRowComprisingSingleString(tableName string, theString string) string {
	output := fmt.Sprintf("INSERT INTO \"%s\"\n", tableName)
	output += fmt.Sprintf("VALUES ('%s');", theString)
	return output
}

func UpdateSingleStringColumn(tableName string, column string, newValue string) string {
	output := fmt.Sprintf("UPDATE \"%s\" SET \"%s\"='%s';", tableName, column, newValue)
	return output
}

// todo - export all these functions and needs making more coherent - a right raggle taggle mix it's become
