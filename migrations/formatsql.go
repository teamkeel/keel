package migrations

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
)

func createTable(model *proto.Model) string {
	output := fmt.Sprintf("CREATE TABLE \"%s\"(\n", model.Name) // todo - we should normalise model names
	for i, field := range model.Fields {
		// todo: field names need to be normalised / standardised for use in the database.
		f := fmt.Sprintf("\"%s\" %s", field.Name, PostgresFieldTypes[field.Type])
		if i != len(model.Fields)-1 {
			f += ","
		}
		f += "\n"
		output += f
	}
	output += ");"
	return output
}

func CreateTableIfNotExists(name string, fields []*proto.Field) string {
	output := fmt.Sprintf("CREATE TABLE if not exists \"%s\"(\n", name) // todo - we should normalise model names
	for i, field := range fields {
		// todo: field names need to be normalised / standardised for use in the database.
		f := fmt.Sprintf("\"%s\" %s", field.Name, PostgresFieldTypes[field.Type])
		if i != len(fields)-1 {
			f += ","
		}
		f += "\n"
		output += f
	}
	output += ");"
	return output
}

func dropTable(name string) string {
	return fmt.Sprintf("DROP TABLE \"%s\";", name)
}

func createField(modelName string, field *proto.Field) string {
	output := fmt.Sprintf("ALTER TABLE \"%s\"\n", modelName)
	output += fmt.Sprintf("ADD \"%s\" %s;", field.Name, PostgresFieldTypes[field.Type])
	return output
}

func dropField(modelName string, fieldName string) string {
	output := fmt.Sprintf("ALTER TABLE \"%s\"\n", modelName)
	output += fmt.Sprintf("DROP \"%s\";", fieldName)
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
