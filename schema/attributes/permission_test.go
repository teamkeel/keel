package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

func TestPermissionRole_Valid(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			actions {
				list listPeople() {
					@permission(roles: [Admin])
				}
			}
		}
		role Admin {
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidatePermissionRoles(schema, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestPermissionRole_InvalidNotArray(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			actions {
				list listPeople() {
					@permission(roles: Admin)
				}
			}
		}
		role Admin {
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidatePermissionRoles(schema, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Role[] but it is Role", issues[0].Message)

	require.Equal(t, 5, issues[0].Pos.Line)
	require.Equal(t, 25, issues[0].Pos.Column)
	require.Equal(t, 79, issues[0].Pos.Offset)
	require.Equal(t, 5, issues[0].EndPos.Line)
	require.Equal(t, 30, issues[0].EndPos.Column)
	require.Equal(t, 84, issues[0].EndPos.Offset)
}

func TestPermissionRole_UnknownRole(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			actions {
				list listPeople() {
					@permission(roles: Unknown)
				}
			}
		}
		role Admin {
		}`})

	action := query.Action(schema, "listPeople")
	expression := action.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidatePermissionRoles(schema, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "unknown identifier 'Unknown'", issues[0].Message)
}

func TestPermissionActions_Valid(t *testing.T) {
	expression, err := parser.ParseExpression("[get, list, create, update, delete]")
	require.NoError(t, err)

	issues, err := attributes.ValidatePermissionActions(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestPermissionActions_NotArray(t *testing.T) {
	expression, err := parser.ParseExpression("list")
	require.NoError(t, err)

	issues, err := attributes.ValidatePermissionActions(expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type ActionType[] but it is ActionType", issues[0].Message)
}

func TestPermissionActions_UnknownValue(t *testing.T) {
	expression, err := parser.ParseExpression("[list,write]")
	require.NoError(t, err)

	issues, err := attributes.ValidatePermissionActions(expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "unknown identifier 'write'", issues[0].Message)

	require.Equal(t, 1, issues[0].Pos.Line)
	require.Equal(t, 6, issues[0].Pos.Column)
	require.Equal(t, 6, issues[0].Pos.Offset)
	require.Equal(t, 1, issues[0].EndPos.Line)
	require.Equal(t, 11, issues[0].EndPos.Column)
	require.Equal(t, 11, issues[0].EndPos.Offset)
}
