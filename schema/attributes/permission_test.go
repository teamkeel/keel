package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/attributes"
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
	where := action.Attributes[0]

	expression := where.Arguments[0].Expression

	parser, err := attributes.NewPermissionRoleParser(schema)
	require.NoError(t, err)

	issues, err := parser.Validate(expression.String())
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
	where := action.Attributes[0]

	expression := where.Arguments[0].Expression

	parser, err := attributes.NewPermissionRoleParser(schema)
	require.NoError(t, err)

	issues, err := parser.Validate(expression.String())
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'list(_RoleDefinition)' but it is '_RoleDefinition'", issues[0])
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
	where := action.Attributes[0]

	expression := where.Arguments[0].Expression

	parser, err := attributes.NewPermissionRoleParser(schema)
	require.NoError(t, err)

	issues, err := parser.Validate(expression.String())
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to 'Unknown' (in container '')", issues[0])
}

func TestPermissionActions_Valid(t *testing.T) {
	expression := "[get, list, create, update, delete]"

	parser, err := attributes.NewPermissionActionsParser()
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestPermissionActions_NotArray(t *testing.T) {
	expression := "list"

	parser, err := attributes.NewPermissionActionsParser()
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'list(_ActionTypeDefinition)' but it is '_ActionTypeDefinition'", issues[0])
}

func TestPermissionActions_UnknownValue(t *testing.T) {
	expression := "[list,write]"

	parser, err := attributes.NewPermissionActionsParser()
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to 'write' (in container '')", issues[0])
}
