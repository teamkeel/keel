package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

func TestValid(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			name Text
			isActive Boolean
		}
		actions {
			create createPerson(name) {
				@set(person.isActive = true)
			}
		}
	}`})

	action := query.Action(schema, "createPerson")
	set := action.Attributes[0]

	operand, expression, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.NoError(t, err)

	require.Equal(t, "person", operand.Ident.Fragments[0].Fragment)
	require.Equal(t, "isActive", operand.Ident.Fragments[1].Fragment)

	parser, err := attributes.NewSetExpressionParser(schema, operand.Ident, action)

	issues, err := parser.Validate(expression.String())
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestValidWithRelationship(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			name Text
			isActive Boolean @default(true)
			company Organisation
		}
		actions {
			create createPerson(name, company.name) {
				@set(person.company.isActive = true)
			}
		}
	}
	model Organisation {
		fields {
			name Text
			isActive Boolean
		}
	}	
	`})

	action := query.Action(schema, "createPerson")
	set := action.Attributes[0]

	operand, expression, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.NoError(t, err)

	require.Equal(t, "person", operand.Ident.Fragments[0].Fragment)
	require.Equal(t, "company", operand.Ident.Fragments[1].Fragment)
	require.Equal(t, "isActive", operand.Ident.Fragments[2].Fragment)

	parser, err := attributes.NewSetExpressionParser(schema, operand.Ident, action)

	issues, err := parser.Validate(expression.String())
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestInvalidTypes(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			name Text
			isActive Boolean
		}
		actions {
			create createPerson(name) {
				@set(person.isActive = "Hello")
			}
		}
	}`})

	action := query.Action(schema, "createPerson")
	set := action.Attributes[0]

	operand, expression, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.NoError(t, err)

	require.Equal(t, "person", operand.Ident.Fragments[0].Fragment)
	require.Equal(t, "isActive", operand.Ident.Fragments[1].Fragment)

	parser, err := attributes.NewSetExpressionParser(schema, operand.Ident, action)

	issues, err := parser.Validate(expression.String())
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'bool' but it is 'string'", issues[0])
}

func parse(t *testing.T, s *reader.SchemaFile) []*parser.AST {
	schema, err := parser.Parse(s)
	if err != nil {
		require.Fail(t, err.Error())
	}

	return []*parser.AST{schema}
}