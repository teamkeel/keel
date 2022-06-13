package migrations

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
)

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
