package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

func TestWhere_Valid(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				isActive Boolean
			}
			actions {
				list listPeople(name) {
					@where(_.name == "Keel")
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestWhere_InvalidType(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				isActive Boolean
			}
			actions {
				list listPeople(name) {
					@where(person.name == 123)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator '==' with types Text and Number", issues[0].Message)
}

func TestWhere_UnknownVariable(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				isActive Boolean
			}
			actions {
				list listPeople(name) {
					@where(person.name == something)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "unknown identifier 'something'", issues[0].Message)
}

func TestWhere_ValidField(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				secondName Text
			}
			actions {
				list listPeople(name) {
					@where(person.name == person.secondName)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestWhere_UnknownField(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				secondName Text
			}
			actions {
				list listPeople(name) {
					@where(person.name == person.something)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "field 'something' does not exist", issues[0].Message)
}

func TestWhere_NamedInput(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				isActive Boolean
			}
			actions {
				list listPeople(n: Text) {
					@where(person.name == n)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestWhere_FieldInput(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				isActive Boolean
			}
			actions {
				list listPeople(name) {
					@where(person.name == name)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestWhere_MultiConditions(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				isActive Boolean
			}
			actions {
				list listPeople() {
					@where(person.name == "Keel" && person.isActive)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateWhereExpression(schema, action, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}
