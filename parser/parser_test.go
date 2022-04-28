package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/parser"
)

func parse(t *testing.T, s string) *parser.Schema {
	schema, err := parser.Parse(s)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return schema
}

func TestEmptyModel(t *testing.T) {
	schema := parse(t, `model Person { }`)
	assert.Equal(t, "Person", schema.Declarations[0].Model.Name)
}

func TestModelWithFields(t *testing.T) {
	schema := parse(t, `
	  model Author {
		  fields {
			name Text
			books Book[]
		  }
		}`)
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)
}

func TestModelWithFunctions(t *testing.T) {
	schema := parse(t, `
	model Author {
		fields {
		  name Text
		  books Book[]
		}
		functions {
		  create createAuthor(name)
		  get author(id)
		}
	  }`)
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)

	assert.Equal(t, true, schema.Declarations[0].Model.Sections[1].Functions[0].Create)
	assert.Equal(t, "createAuthor", schema.Declarations[0].Model.Sections[1].Functions[0].Name)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Functions[0].Arguments, 1)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[1].Functions[0].Arguments[0].Name)

	assert.Equal(t, true, schema.Declarations[0].Model.Sections[1].Functions[1].Get)
	assert.Equal(t, "author", schema.Declarations[0].Model.Sections[1].Functions[1].Name)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Functions[1].Arguments, 1)
	assert.Equal(t, "id", schema.Declarations[0].Model.Sections[1].Functions[1].Arguments[0].Name)
}

func TestModelWithFieldAttributes(t *testing.T) {
	schema := parse(t, `
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
	  }`)
	assert.Len(t, schema.Declarations[0].Model.Sections[0].Fields[1].Attributes, 1)
	assert.Equal(t, "unique", schema.Declarations[0].Model.Sections[0].Fields[1].Attributes[0].Name)
}

func TestModelWithPermissionAttributes(t *testing.T) {
	schema := parse(t, `
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
		  [get],
		  true
		)
	  }`)
	assert.Equal(t, "permission", schema.Declarations[0].Model.Sections[2].Attribute.Name)
	assert.Equal(t, "get", schema.Declarations[0].Model.Sections[2].Attribute.Arguments[0].Value.Array[0].Ident[0])
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1].Value.True)
}

func TestModelWithExpressionAttribute(t *testing.T) {
	schema := parse(t, `
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
		  [create],
		  "admin" in ctx.identity.roles
		)

		@permission(
	      [get],
		  ctx.identity = author.identity
		)
	  }`)

	expr := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1].Expression
	assert.Equal(t, "admin", expr.LHS.String)
	assert.Equal(t, "in", expr.Op)
	assert.Equal(t, "ctx.identity.roles", strings.Join(expr.RHS.Ident, "."))

	expr = schema.Declarations[0].Model.Sections[3].Attribute.Arguments[1].Expression
	assert.Equal(t, "ctx.identity", strings.Join(expr.LHS.Ident, "."))
	assert.Equal(t, "=", expr.Op)
	assert.Equal(t, "author.identity", strings.Join(expr.RHS.Ident, "."))
}
