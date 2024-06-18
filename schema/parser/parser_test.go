package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			netWorth Decimal
		  }
		}`})
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name.Value)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name.Value)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type.Value)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name.Value)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type.Value)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)

	assert.Equal(t, "rating", schema.Declarations[0].Model.Sections[0].Fields[2].Name.Value)
	assert.Equal(t, "Number", schema.Declarations[0].Model.Sections[0].Fields[2].Type.Value)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[2].Repeated)

	assert.Equal(t, "netWorth", schema.Declarations[0].Model.Sections[0].Fields[3].Name.Value)
	assert.Equal(t, "Decimal", schema.Declarations[0].Model.Sections[0].Fields[3].Type.Value)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[3].Repeated)
}

func TestModelWithFunctions(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Author {
		fields {
		  name Text
		  books Book[]
		}
		actions {
		  create createAuthor(name) {
				@function
			}
		  get author(id) {
				@function
			}
		}
	  }`})
	assert.Equal(t, "Author", schema.Declarations[0].Model.Name.Value)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[0].Fields[0].Name.Value)
	assert.Equal(t, "Text", schema.Declarations[0].Model.Sections[0].Fields[0].Type.Value)
	assert.Equal(t, false, schema.Declarations[0].Model.Sections[0].Fields[0].Repeated)

	assert.Equal(t, "books", schema.Declarations[0].Model.Sections[0].Fields[1].Name.Value)
	assert.Equal(t, "Book", schema.Declarations[0].Model.Sections[0].Fields[1].Type.Value)
	assert.Equal(t, true, schema.Declarations[0].Model.Sections[0].Fields[1].Repeated)

	assert.Equal(t, "create", schema.Declarations[0].Model.Sections[1].Actions[0].Type.Value)
	assert.Equal(t, "createAuthor", schema.Declarations[0].Model.Sections[1].Actions[0].Name.Value)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Actions[0].Inputs, 1)
	assert.Equal(t, "name", schema.Declarations[0].Model.Sections[1].Actions[0].Inputs[0].Type.Fragments[0].Fragment)

	assert.Equal(t, "get", schema.Declarations[0].Model.Sections[1].Actions[1].Type.Value)
	assert.Equal(t, "author", schema.Declarations[0].Model.Sections[1].Actions[1].Name.Value)
	assert.Len(t, schema.Declarations[0].Model.Sections[1].Actions[1].Inputs, 1)
	assert.Equal(t, "id", schema.Declarations[0].Model.Sections[1].Actions[1].Inputs[0].Type.Fragments[0].Fragment)
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
		actions {
		  create createBook(title, authors) {
				@function
			}
		  get book(id) {
				@function
			}
		  get bookByIsbn(isbn) {
				@function
			}
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

		actions {
		  create createAuthor(name) {
				@function
			}
		  get author(id) {
				@function
			}
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
	assert.Equal(t, true, arg1.Expression.IsValue())
	assert.Equal(t, "expression", arg1.Label.Value)

	arg2 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1]
	assert.Equal(t, true, arg2.Expression.IsValue())
	assert.Equal(t, "actions", arg2.Label.Value)

	v1, err := arg1.Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, true, v1.True)

	v2, err := arg2.Expression.ToValue()
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

		actions {
		  create createAuthor(name) {
				@function
			}
		  get author(id) {
				@function
			}
		}

		@permission(
			role: Admin,
			actions: [create]
		)
	  }`})

	arg1 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[0]
	assert.Equal(t, true, arg1.Expression.IsValue())
	assert.Equal(t, "role", arg1.Label.Value)

	arg2 := schema.Declarations[0].Model.Sections[2].Attribute.Arguments[1]
	assert.Equal(t, true, arg2.Expression.IsValue())
	assert.Equal(t, "actions", arg2.Label.Value)

	v1, err := arg1.Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "Admin", v1.Ident.Fragments[0].Fragment)

	v2, err := arg2.Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "create", v2.Array.Values[0].Ident.Fragments[0].Fragment)
}

func TestOperationWithOrderByAttribute(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
model Author {
    fields {
        firstName Text
        surname Text
    }

    actions {
        list listAuthors() {
            @orderBy(firstName: asc, surname: desc)
        }
    }
}`})

	attribute := schema.Declarations[0].Model.Sections[1].Actions[0].Attributes[0]

	assert.Equal(t, "orderBy", attribute.Name.Value)

	arg1 := attribute.Arguments[0]
	assert.Equal(t, true, arg1.Expression.IsValue())
	assert.Equal(t, "firstName", arg1.Label.Value)

	arg2 := attribute.Arguments[1]
	assert.Equal(t, true, arg2.Expression.IsValue())
	assert.Equal(t, "surname", arg2.Label.Value)

	v1, err := arg1.Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "asc", v1.Ident.Fragments[0].Fragment)

	v2, err := arg2.Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "desc", v2.Ident.Fragments[0].Fragment)
}

func TestOperationWithSortableAttribute(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
model Author {
    fields {
        firstName Text
        surname Text
    }

    actions {
        list listAuthors() {
            @sortable(firstName, surname)
        }
    }
}`})

	attribute := schema.Declarations[0].Model.Sections[1].Actions[0].Attributes[0]

	assert.Equal(t, "sortable", attribute.Name.Value)

	arg1 := attribute.Arguments[0]
	assert.Equal(t, true, arg1.Expression.IsValue())
	assert.Nil(t, arg1.Label)

	arg2 := attribute.Arguments[1]
	assert.Equal(t, true, arg2.Expression.IsValue())
	assert.Nil(t, arg2.Label)

	v1, err := arg1.Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "firstName", v1.Ident.Fragments[0].Fragment)

	v2, err := arg2.Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "surname", v2.Ident.Fragments[0].Fragment)
}

func TestOperationWithEmbedAttribute(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
model Author {
    fields {
        firstName Text
        surname Text
    }
}
model Book { 
	fields {
		title Text
		author Author
	}

	actions {
		get getBook(id) {
			@embed(author.firstName, author.surname)
			@embed(author.id)
		}
	}
}
`})

	attribute := schema.Declarations[1].Model.Sections[1].Actions[0].Attributes[0]
	attribute2 := schema.Declarations[1].Model.Sections[1].Actions[0].Attributes[1]

	assert.Equal(t, "embed", attribute.Name.Value)
	assert.Equal(t, "embed", attribute2.Name.Value)

	arg1 := attribute.Arguments[0]
	assert.Equal(t, true, arg1.Expression.IsValue())
	assert.Nil(t, arg1.Label)

	arg2 := attribute.Arguments[1]
	assert.Equal(t, true, arg2.Expression.IsValue())
	assert.Nil(t, arg2.Label)

	v1, err := arg1.Expression.ToString()
	assert.NoError(t, err)
	assert.Equal(t, "author.firstName", v1)

	v2, err := arg2.Expression.ToString()
	assert.NoError(t, err)
	assert.Equal(t, "author.surname", v2)

	arg := attribute2.Arguments[0]
	assert.Equal(t, true, arg.Expression.IsValue())
	assert.Nil(t, arg.Label)

	v, err := arg.Expression.ToString()
	assert.NoError(t, err)
	assert.Equal(t, "author.id", v)

}

func TestAPI(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	api Web {
		models {
			Author
			Book
		}
	}
	api Api {
		models {
			Book
		}
	}`})
	assert.Equal(t, "Web", schema.Declarations[0].API.Name.Value)
	assert.Equal(t, "Author", schema.Declarations[0].API.Sections[0].Models[0].Name.Value)
	assert.Equal(t, "Book", schema.Declarations[0].API.Sections[0].Models[1].Name.Value)

	assert.Equal(t, "Api", schema.Declarations[1].API.Name.Value)
	assert.Equal(t, "Book", schema.Declarations[1].API.Sections[0].Models[0].Name.Value)
}

func TestAPIWithActions(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	api Admin {
		models {
			Author {
				actions {
					deleteAuthor
					deleteAll
				}
			}
			Book {
				actions {
					deleteBook
					deleteAll
				}
			}
		}
	}
	api Api {
		models {
			Author {
				actions {
					getAuthor
				}
			}
			Book {
				actions {
					searchBooks
				}
			}
		}
	}`})

	assert.Equal(t, "Admin", schema.Declarations[0].API.Name.Value)

	assert.Equal(t, "Author", schema.Declarations[0].API.Sections[0].Models[0].Name.Value)
	assert.Equal(t, "deleteAuthor", schema.Declarations[0].API.Sections[0].Models[0].Sections[0].Actions[0].Name.Value)
	assert.Equal(t, "deleteAll", schema.Declarations[0].API.Sections[0].Models[0].Sections[0].Actions[1].Name.Value)

	assert.Equal(t, "Book", schema.Declarations[0].API.Sections[0].Models[1].Name.Value)
	assert.Equal(t, "deleteBook", schema.Declarations[0].API.Sections[0].Models[1].Sections[0].Actions[0].Name.Value)
	assert.Equal(t, "deleteAll", schema.Declarations[0].API.Sections[0].Models[1].Sections[0].Actions[1].Name.Value)

	assert.Equal(t, "Api", schema.Declarations[1].API.Name.Value)

	assert.Equal(t, "Author", schema.Declarations[1].API.Sections[0].Models[0].Name.Value)
	assert.Equal(t, "getAuthor", schema.Declarations[1].API.Sections[0].Models[0].Sections[0].Actions[0].Name.Value)

	assert.Equal(t, "Book", schema.Declarations[1].API.Sections[0].Models[1].Name.Value)
	assert.Equal(t, "searchBooks", schema.Declarations[1].API.Sections[0].Models[1].Sections[0].Actions[0].Name.Value)
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

func TestAttributeArgsParsing(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		@permission
		@permission()
		@permission(expression: true, actions: [get])
	}`})

	model := schema.Declarations[0].Model
	assert.Equal(t, "permission", model.Sections[0].Attribute.Name.Value)
	assert.Equal(t, "permission", model.Sections[1].Attribute.Name.Value)
	assert.Equal(t, "permission", model.Sections[2].Attribute.Name.Value)
	assert.Len(t, model.Sections[0].Attribute.Arguments, 0)
	assert.Len(t, model.Sections[1].Attribute.Arguments, 0)
	assert.Len(t, model.Sections[2].Attribute.Arguments, 2)
}

func TestArbitraryFunctions(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	message MyInput {
		id ID
	}
	message MyOutput {
		name Text
	}
	model Person {
		actions {
			read myThing(MyInput) returns (MyOutput) {
				@function
			}
		}
	}`})

	myInput := schema.Declarations[0].Message.Name.Value
	assert.Equal(t, "MyInput", myInput)
	myOutput := schema.Declarations[1].Message.Name.Value
	assert.Equal(t, "MyOutput", myOutput)
	model := schema.Declarations[2].Model
	assert.Equal(t, "Person", model.Name.Value)
	assert.Equal(t, "MyInput", model.Sections[0].Actions[0].Inputs[0].Type.Fragments[0].Fragment)
	assert.Equal(t, "MyOutput", model.Sections[0].Actions[0].Returns[0].Type.Fragments[0].Fragment)
}

func TestArbitraryFunctionsNestedMessage(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	message MyInput {
		nested NestedInput
	}
	message NestedInput {
		title Text
	}
	message MyOutput {
		name Text
	}
	model Person {
		actions {
			read myThing(MyInput) returns (MyOutput) {
				@function
			}
		}
	}`})

	myInput := schema.Declarations[0].Message.Name.Value
	assert.Equal(t, "MyInput", myInput)
	nestedFieldType := schema.Declarations[0].Message.Fields[0].Type.Value
	assert.Equal(t, "NestedInput", nestedFieldType)
	nestedInput := schema.Declarations[1].Message.Name.Value
	assert.Equal(t, "NestedInput", nestedInput)
}

func TestJobInputs(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	job MyJob {
  inputs {
    veryImportantValue Text
    someFlag Boolean?
  }

  @permission(roles: [Developer])
  @schedule("...")
}`})
	assert.Equal(t, "MyJob", schema.Declarations[0].Job.Name.Value)
	assert.Equal(t, "schedule", schema.Declarations[0].Job.Sections[2].Attribute.Name.Value)
	assert.Equal(t, "veryImportantValue", schema.Declarations[0].Job.Sections[0].Inputs[0].Name.Value)
	assert.Equal(t, "someFlag", schema.Declarations[0].Job.Sections[0].Inputs[1].Name.Value)
	assert.Equal(t, true, schema.Declarations[0].Job.Sections[0].Inputs[1].Optional)
	assert.Equal(t, "Text", schema.Declarations[0].Job.Sections[0].Inputs[0].Type.Value)
}

func TestOnAttributeArgsParsing(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		@on([create], sendWelcomeMail)
		@on([create, update], verifyEmail)
	}`})

	model := schema.Declarations[0].Model
	assert.Equal(t, "on", model.Sections[0].Attribute.Name.Value)
	assert.Equal(t, "on", model.Sections[1].Attribute.Name.Value)
	assert.Len(t, model.Sections[0].Attribute.Arguments, 2)
	assert.Len(t, model.Sections[1].Attribute.Arguments, 2)

	on1actiontypes, err := schema.Declarations[0].Model.Sections[0].Attribute.Arguments[0].Expression.ToValue()
	assert.NoError(t, err)
	assert.Len(t, on1actiontypes.Array.Values, 1)
	assert.Equal(t, "create", on1actiontypes.Array.Values[0].Ident.Fragments[0].Fragment)

	on1subscriber, err := schema.Declarations[0].Model.Sections[0].Attribute.Arguments[1].Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "sendWelcomeMail", on1subscriber.Ident.Fragments[0].Fragment)

	on2actiontypes, err := schema.Declarations[0].Model.Sections[1].Attribute.Arguments[0].Expression.ToValue()
	assert.NoError(t, err)
	assert.Len(t, on2actiontypes.Array.Values, 2)
	assert.Equal(t, "create", on2actiontypes.Array.Values[0].Ident.Fragments[0].Fragment)
	assert.Equal(t, "update", on2actiontypes.Array.Values[1].Ident.Fragments[0].Fragment)

	on2subscriber, err := schema.Declarations[0].Model.Sections[1].Attribute.Arguments[1].Expression.ToValue()
	assert.NoError(t, err)
	assert.Equal(t, "verifyEmail", on2subscriber.Ident.Fragments[0].Fragment)
}
