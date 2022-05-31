package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/inputs"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

func parse(t *testing.T, s *inputs.SchemaFile) *parser.Schema {
	schema, err := parser.Parse(s)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return schema
}

func TestEmptyModel(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `model Person { }`})
	assert.Equal(t, "Person", schema.Declarations[0].Model.Name)
}

func TestModelWithFields(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `
	  model Author {
		  fields {
			name Text
			books Book[]
		  }
		}`})
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)
}

func TestModelWithFunctions(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `
	model Author {
		fields {
		  name Text
		  books Book[]
		}
		functions {
		  create createAuthor(name)
		  get author(id)
		}
	  }`})
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)

	assert.Equal(t, "create", schema.Declarations[0].Model.Sections[1].Functions[0].Type)
	assert.Equal(t, "createAuthor", schema.Declarations[0].Model.Sections[1].Functions[0].Name)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Functions[0].Arguments, 1)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[1].Functions[0].Arguments[0].Name)

	assert.Equal(t, "get", schema.Declarations[0].Model.Sections[1].Functions[1].Type)
	assert.Equal(t, "author", schema.Declarations[0].Model.Sections[1].Functions[1].Name)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Functions[1].Arguments, 1)
	assert.Equal(t, "id", schema.Declarations[0].Model.Sections[1].Functions[1].Arguments[0].Name)
}

func TestModelWithFieldAttributes(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `
	model Book {
		fields {
		  title Text
		  isbn Text {
			@unique
		  }
		  authors Author[]
		}
		functions {
		  create createBook(title, authors)
		  get book(id)
		  get bookByIsbn(isbn)
		}
	  }`})
	assert.Len(t, schema.Declarations[0].Model.Sections[0].Fields[1].Attributes, 1)
	assert.Equal(t, "unique", schema.Declarations[0].Model.Sections[0].Fields[1].Attributes[0].Name)
}

func TestRole(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `
	model Post {
			fields {
				title Text
			}
			@permission(
				actions: [get],
				role: Admin
			)
		}
	  role Admin {
			domains {
				"keel.xyz"
				"keel.zyx"
			}
			emails {
				"adam@keel.xyz"
				"adam@keel.zyx"
			}
		}
	`})

	assert.Equal(t, "Admin", schema.Declarations[1].Role.Name)

	assert.Equal(t, "\"adam@keel.xyz\"", schema.Declarations[1].Role.Sections[1].Emails[0].Email)
	assert.Equal(t, "\"adam@keel.zyx\"", schema.Declarations[1].Role.Sections[1].Emails[1].Email)

	assert.Equal(t, "\"keel.xyz\"", schema.Declarations[1].Role.Sections[0].Domains[0].Domain)
	assert.Equal(t, "\"keel.zyx\"", schema.Declarations[1].Role.Sections[0].Domains[1].Domain)
}

func TestModelWithPermissionAttributes(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `
	model Author {
		fields {
		  name Text
		  books Book[]
		}
	  
		functions {
		  create createAuthor(name)
		  get author(id)
		}
	  
		@permission(
		  expression: true,
		  actions: [get],
		  role: Admin
		)
	}
	role Admin {
		emails {
			"adam@keel.xyz"
		}
	}`})
	assert.Equal(t, "permission", schema.Declarations[0].Model.Sections[2].Attribute.Name)

	arg1 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[0]
	assert.Equal(t, true, expressions.IsValue(arg1.Expression))
	assert.Equal(t, "expression", arg1.Name)

	arg2 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1]
	assert.Equal(t, true, expressions.IsValue(arg2.Expression))
	assert.Equal(t, "actions", arg2.Name)

	v1, err := expressions.ToValue(arg1.Expression)
	assert.NoError(t, err)
	assert.Equal(t, true, v1.True)

	v2, err := expressions.ToValue(arg2.Expression)
	assert.NoError(t, err)
	assert.Equal(t, "get", v2.Array.Values[0].Ident[0])
}

func TestAttributeWithNamedArguments(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `
	model Author {
		fields {
		  identity Identity
		  name Text
		  books Book[]
		}
	  
		functions {
		  create createAuthor(name)
		  get author(id)
		}

		@permission(
			role: Admin,
			actions: [create]
		)
	  }`})

	arg1 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[0]
	assert.Equal(t, true, expressions.IsValue(arg1.Expression))
	assert.Equal(t, "role", arg1.Name)

	arg2 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1]
	assert.Equal(t, true, expressions.IsValue(arg2.Expression))
	assert.Equal(t, "actions", arg2.Name)

	v1, err := expressions.ToValue(arg1.Expression)
	assert.NoError(t, err)
	assert.Equal(t, "Admin", v1.Ident[0])

	v2, err := expressions.ToValue(arg2.Expression)
	assert.NoError(t, err)
	assert.Equal(t, "create", v2.Array.Values[0].Ident[0])
}

func TestAPI(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `
	api Web {
		@graphql

		models {
			Author
			Book
		}
	}`})
	assert.Equal(t, "Web", schema.Declarations[0].API.Name)

	assert.Equal(t, "graphql", schema.Declarations[0].API.Sections[0].Attribute.Name)

	assert.Equal(t, "Author", schema.Declarations[0].API.Sections[1].Models[0].ModelName)
	assert.Equal(t, "Book", schema.Declarations[0].API.Sections[1].Models[1].ModelName)
}

func TestParserPos(t *testing.T) {
	schema := parse(t, &inputs.SchemaFile{FileName: "test.keel", Contents: `model Author {
    fields {
        name TextyTexty
    }
}`})

	// The field defintion starts on line 3 character 9
	assert.Equal(t, 3, schema.Declarations[0].Model.Sections[0].Fields[0].Pos.Line)
	assert.Equal(t, 9, schema.Declarations[0].Model.Sections[0].Fields[0].Pos.Column)
}
