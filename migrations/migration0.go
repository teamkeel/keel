package migrations

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/teamkeel/keel/proto"
)

const databaseName string = "keel"

// A Migration0 object knows how to generate the SQL to set up the database
// from scratch, i.e. starting from nothing.
// It is re-entrant in the sense that it removes the existing database
// before it starts if one exists already.
type Migration0 struct {
	schema *proto.Schema
	SQL    []string
}

func NewMigration0(schema *proto.Schema) *Migration0 {
	return &Migration0{
		schema: schema,
		SQL:    []string{},
	}
}

func (m0 *Migration0) GenerateSQL() {
	// First drop the entire existing database if it exists.
	s := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", databaseName)
	m0.SQL = append(m0.SQL, s)

	// todo add db creation and selection

	// now the model tables
	for _, model := range m0.schema.Models {
		m0.appendCreateModel(model)
	}
}

func (m0 *Migration0) appendCreateModel(model *proto.Model) {

	templateData := table{
		Name:    model.Name, // Todo can we use the proto model names as they stand?
		Columns: []*column{},
	}
	for _, field := range model.Fields {
		templateData.Columns = append(templateData.Columns, m0.column(field))
	}

	tmpl, err := template.New("foo").Parse(templateCreateTable)
	if err != nil {
		panic(fmt.Sprintf("error parsing template: %v", err))
	}
	output := bytes.Buffer{}
	err = tmpl.Execute(&output, templateData)
	if err != nil {
		panic(fmt.Sprintf("error executing template: %v", err))
	}

	s := output.String()
	m0.SQL = append(m0.SQL, s)
}

func (m0 *Migration0) column(field *proto.Field) *column {
	return &column{
		Name: field.Name,
		Type: PostgresFieldTypes[field.Type],
	}
}

type table struct {
	Name    string
	Columns []*column
}

type column struct {
	Name string
	Type string
}

// todo: need to add [] when field type is a list
const templateCreateTable string = `CREATE TABLE {{.Name}} (
{{range .Columns}}  {{.Name}} {{.Type}},
{{end}});`
