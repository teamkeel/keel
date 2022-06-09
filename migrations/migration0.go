package migrations

import (
	"fmt"
	"os"

	"github.com/teamkeel/keel/proto"
)

func GenerateAllTables(models []*proto.Model) string {
	output := ""
	for _, model := range models {
		output += createTable(model)
	}
	// todo - similar for API's, Enums, etc.

	if os.Getenv("DEBUG") != "" {
		fmt.Printf("\n%s\n\n", output)
	}
	return output
}

func createTable(model *proto.Model) string {

	output := fmt.Sprintf("\nCREATE TABLE %s(\n", model.Name) // Should we apply transformation proto model name?
	for i, field := range model.Fields {
		f := fmt.Sprintf("%s %s", field.Name, PostgresFieldTypes[field.Type])
		if i != len(model.Fields)-1 {
			f += ","
		}
		f += "\n"
		output += f
	}
	output += ");"
	return output
}

// todos:
// add [] when field type is a list
// add field constraints - particularly not null is !OPTIONAL
