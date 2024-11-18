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
					@where(person.name == "Keel")
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	where := action.Attributes[0]

	expression := where.Arguments[0].Expression

	parser, err := attributes.NewWhereExpressionParser(schema, action)

	issues, err := parser.Validate(expression.String())
	require.NoError(t, err)
	require.Empty(t, issues)
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
	where := action.Attributes[0]

	expression := where.Arguments[0].Expression

	parser, err := attributes.NewWhereExpressionParser(schema, action)

	issues, err := parser.Validate(expression.String())
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
					@where(person.name == "Keel" and person.isActive)
				}
			}
		}`})

	action := query.Action(schema, "listPeople")
	where := action.Attributes[0]

	expression := where.Arguments[0].Expression

	parser, err := attributes.NewWhereExpressionParser(schema, action)

	issues, err := parser.Validate(expression.String())
	require.NoError(t, err)
	require.Empty(t, issues)
}
