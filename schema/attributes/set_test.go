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

func TestSet_Valid(t *testing.T) {
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

	lhs, err := resolve.AsIdent(target)
	require.NoError(t, err)

	require.Equal(t, "person", lhs.Fragments[0])
	require.Equal(t, "isActive", lhs.Fragments[1])

	issues, err := attributes.ValidateSetExpression(schema, action, target, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestSet_ValidWithRelationship(t *testing.T) {
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

	lhs, err := resolve.AsIdent(target)
	require.NoError(t, err)

	require.Equal(t, "person", lhs.Fragments[0])
	require.Equal(t, "company", lhs.Fragments[1])
	require.Equal(t, "isActive", lhs.Fragments[2])

	issues, err := attributes.ValidateSetExpression(schema, action, target, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestSet_InvalidTypes(t *testing.T) {
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

	lhs, err := resolve.AsIdent(target)
	require.NoError(t, err)

	require.Equal(t, "person", lhs.Fragments[0])
	require.Equal(t, "isActive", lhs.Fragments[1])

	issues, err := attributes.ValidateSetExpression(schema, action, target, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Boolean but it is Text", issues[0].Message)

	require.Equal(t, 9, issues[0].Pos.Line)
	require.Equal(t, 10, issues[0].Pos.Column)
	require.Equal(t, 134, issues[0].Pos.Offset)
	require.Equal(t, 1, issues[0].EndPos.Line)
	require.Equal(t, 30, issues[0].EndPos.Column)
	require.Equal(t, 84, issues[0].EndPos.Offset)
}

func TestSet_InvalidAssignmentExpression(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			isActive Boolean
		}
		actions {
			create createPerson(name) {
				@set(person.isActive)
			}
		}
	}`})

	action := query.Action(schema, "createPerson")
	set := action.Attributes[0]

	_, _, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.ErrorIs(t, err, parser.ErrInvalidAssignmentExpression)
}

func TestSet_InvalidAssignmentExpression2(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			isActive Boolean
		}
		actions {
			create createPerson(name) {
				@set(123)
			}
		}
	}`})

	action := query.Action(schema, "createPerson")
	set := action.Attributes[0]

	_, _, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.ErrorIs(t, err, parser.ErrInvalidAssignmentExpression)
}

func TestSet_InvalidAssignmentExpression3(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			isActive Boolean
		}
		actions {
			create createPerson(name) {
				@set(post.isActive =)
			}
		}
	}`})

	action := query.Action(schema, "createPerson")
	set := action.Attributes[0]

	_, _, err := set.Arguments[0].Expression.ToAssignmentExpression()
	require.ErrorIs(t, err, parser.ErrInvalidAssignmentExpression)
}

func parse(t *testing.T, s *reader.SchemaFile) []*parser.AST {
	schema, err := parser.Parse(s)
	if err != nil {
		require.Fail(t, err.Error())
	}

	return []*parser.AST{schema}
}
