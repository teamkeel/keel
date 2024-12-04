package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/expressions/resolve"
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

	target, expression, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.NoError(t, err)

	lhs, err := resolve.AsIdent(target.String())
	require.NoError(t, err)

	require.Equal(t, "person", lhs[0])
	require.Equal(t, "isActive", lhs[1])

	issues, err := attributes.ValidateSetExpression(schema, action, target, expression)
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

	target, expression, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.NoError(t, err)

	lhs, err := resolve.AsIdent(target.String())
	require.NoError(t, err)

	require.Equal(t, "person", lhs[0])
	require.Equal(t, "company", lhs[1])
	require.Equal(t, "isActive", lhs[2])

	issues, err := attributes.ValidateSetExpression(schema, action, target, expression)
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

	target, expression, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.NoError(t, err)

	lhs, err := resolve.AsIdent(target.String())
	require.NoError(t, err)

	require.Equal(t, "person", lhs[0])
	require.Equal(t, "isActive", lhs[1])

	issues, err := attributes.ValidateSetExpression(schema, action, target, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Boolean but it is Text", issues[0].Message)
}

func parse(t *testing.T, s *reader.SchemaFile) []*parser.AST {
	schema, err := parser.Parse(s)
	if err != nil {
		require.Fail(t, err.Error())
	}

	return []*parser.AST{schema}
}
