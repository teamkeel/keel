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
