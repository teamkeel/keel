package migrations

import (
	"fmt"

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

func createTableStmt(model *proto.Model) string {
	output := fmt.Sprintf("CREATE TABLE \"%s\"(\n", model.Name)
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
	output := fmt.Sprintf("CREATE TABLE if not exists \"%s\"(\n", name)
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
	return fmt.Sprintf("DROP TABLE \"%s\";", name)
}

func addColumnStmt(modelName string, field *proto.Field) string {
	output := fmt.Sprintf("ALTER TABLE %s ADD COLUMN ", pq.QuoteIdentifier(modelName))
	output += fieldDefinition(field) + ";"
	return output
}

func fieldDefinition(field *proto.Field) string {
	output := fmt.Sprintf("%s %s", pq.QuoteIdentifier(field.Name), PostgresFieldTypes[field.Type.Type])
	if !field.Optional {
		output += " NOT NULL"
	}

	if field.Unique {
		output += " UNIQUE"
	}

	return output
}

func dropColumnStmt(modelName string, fieldName string) string {
	output := fmt.Sprintf("ALTER TABLE \"%s\" ", modelName)
	output += fmt.Sprintf("DROP COLUMN \"%s\";", fieldName)
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
