package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
)

func parse(t *testing.T, s *reader.SchemaFile) *parser.AST {
	schema, err := parser.Parse(s)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return schema
}

func TestEmptyModel(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `model Person { }`})
	assert.Equal(t, "Person", schema.Declarations[0].Model.Name.Value)
}

func TestModelWithFields(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	  model Author {
		  fields {
			name Text
			books Book[]
			rating Number
		  }
		}`})
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name.Value)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name.Value)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name.Value)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)

	assert.Equal(t, "rating", schema.Declarations[0].Model.Sections[0].Fields[2].Name.Value)
	assert.Equal(t, "Number", schema.Declarations[0].Model.Sections[0].Fields[2].Type)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[2].Repeated)

}

func TestModelWithFunctions(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
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
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name.Value)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name.Value)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name.Value)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)

	assert.Equal(t, "create", schema.Declarations[0].Model.Sections[1].Functions[0].Type)
	assert.Equal(t, "createAuthor", schema.Declarations[0].Model.Sections[1].Functions[0].Name.Value)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Functions[0].Inputs, 1)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[1].Functions[0].Inputs[0].Type.Fragments[0].Fragment)

	assert.Equal(t, "get", schema.Declarations[0].Model.Sections[1].Functions[1].Type)
	assert.Equal(t, "author", schema.Declarations[0].Model.Sections[1].Functions[1].Name.Value)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Functions[1].Inputs, 1)
	assert.Equal(t, "id", schema.Declarations[0].Model.Sections[1].Functions[1].Inputs[0].Type.Fragments[0].Fragment)
}

func TestModelWithFieldAttributes(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
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
	assert.Equal(t, "unique", schema.Declarations[0].Model.Sections[0].Fields[1].Attributes[0].Name.Value)
}

func TestRole(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
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

	assert.Equal(t, "Admin", schema.Declarations[1].Role.Name.Value)

	assert.Equal(t, "\"adam@keel.xyz\"", schema.Declarations[1].Role.Sections[1].Emails[0].Email)
	assert.Equal(t, "\"adam@keel.zyx\"", schema.Declarations[1].Role.Sections[1].Emails[1].Email)

	assert.Equal(t, "\"keel.xyz\"", schema.Declarations[1].Role.Sections[0].Domains[0].Domain)
	assert.Equal(t, "\"keel.zyx\"", schema.Declarations[1].Role.Sections[0].Domains[1].Domain)
}

func TestModelWithPermissionAttributes(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
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
	assert.Equal(t, "permission", schema.Declarations[0].Model.Sections[2].Attribute.Name.Value)

	arg1 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[0]
	assert.Equal(t, true, expressions.IsValue(arg1.Expression))
	assert.Equal(t, "expression", arg1.Label.Value)

	arg2 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1]
	assert.Equal(t, true, expressions.IsValue(arg2.Expression))
	assert.Equal(t, "actions", arg2.Label.Value)

	v1, err := expressions.ToValue(arg1.Expression)
	assert.NoError(t, err)
	assert.Equal(t, true, v1.True)

	v2, err := expressions.ToValue(arg2.Expression)
	assert.NoError(t, err)
	assert.Equal(t, "get", v2.Array.Values[0].Ident.Fragments[0].Fragment)
}

func TestAttributeWithNamedArguments(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
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
	assert.Equal(t, "role", arg1.Label.Value)

	arg2 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1]
	assert.Equal(t, true, expressions.IsValue(arg2.Expression))
	assert.Equal(t, "actions", arg2.Label.Value)

	v1, err := expressions.ToValue(arg1.Expression)
	assert.NoError(t, err)
	assert.Equal(t, "Admin", v1.Ident.Fragments[0].Fragment)

	v2, err := expressions.ToValue(arg2.Expression)
	assert.NoError(t, err)
	assert.Equal(t, "create", v2.Array.Values[0].Ident.Fragments[0].Fragment)
}

func TestAPI(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	api Web {
		@graphql

		models {
			Author
			Book
		}
	}`})
	assert.Equal(t, "Web", schema.Declarations[0].API.Name.Value)

	assert.Equal(t, "graphql", schema.Declarations[0].API.Sections[0].Attribute.Name.Value)

	assert.Equal(t, "Author", schema.Declarations[0].API.Sections[1].Models[0].Name.Value)
	assert.Equal(t, "Book", schema.Declarations[0].API.Sections[1].Models[1].Name.Value)
}

func TestParserPos(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `model Author {
    fields {
        name TextyTexty
    }
}`})

	// The field defintion starts on line 3 character 9
	assert.Equal(t, 3, schema.Declarations[0].Model.Sections[0].Fields[0].Pos.Line)
	assert.Equal(t, 9, schema.Declarations[0].Model.Sections[0].Fields[0].Pos.Column)
}

func TestEnum(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	enum Planets {
		Mercury
		Venus
		Earth
		Mars
	}`})
	assert.Equal(t, "Planets", schema.Declarations[0].Enum.Name.Value)

	assert.Equal(t, "Mercury", schema.Declarations[0].Enum.Values[0].Name.Value)
	assert.Equal(t, "Venus", schema.Declarations[0].Enum.Values[1].Name.Value)
	assert.Equal(t, "Earth", schema.Declarations[0].Enum.Values[2].Name.Value)
	assert.Equal(t, "Mars", schema.Declarations[0].Enum.Values[3].Name.Value)
}
