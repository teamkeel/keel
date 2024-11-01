package expressions

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/schema"
)

type testCase struct {
	// Name given to the test case
	name string
	// Valid keel schema for this test case
	keelSchema string
	// action name to run test upon
	actionName string
	// Input map for action
	input map[string]any
	// OPTIONAL: Authenticated identity for the query
	identity auth.Identity
	// Expected SQL template generated (with ? placeholders for values)
	expectedTemplate string
	// OPTIONAL: Expected ordered argument slice
	expectedArgs []any
}

var testCases = []testCase{
	{
		name: "get_op_by_id_where",
		keelSchema: `
			model Thing {
				fields {
					isActive Boolean
				}
				actions {
					get getThing(id) {
						@where(thing.isActive == true)
					}
				}
				@permission(expression: true, actions: [get])
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*
			FROM
				"thing"
			WHERE
				"thing"."id" IS NOT DISTINCT FROM ?
				AND "thing"."is_active" IS NOT DISTINCT FROM ?`,
		expectedArgs: []any{"123", true},
	},
}

func TestQueryBuilder(t *testing.T) {
	t.Parallel()
	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			if testCase.identity != nil {
				ctx = auth.WithIdentity(ctx, testCase.identity)
			}

			scope, query, action, err := generateQueryScope(ctx, testCase.keelSchema, testCase.actionName)
			if err != nil {
				require.NoError(t, err)
			}

			var statement *actions.Statement
			switch action.Type {
			case proto.ActionType_ACTION_TYPE_GET:
				statement, err = generateGetStatement(query, scope, testCase.input)
			case proto.ActionType_ACTION_TYPE_LIST:
				statement, _, err = actions.GenerateListStatement(query, scope, testCase.input)
			case proto.ActionType_ACTION_TYPE_CREATE:
				statement, err = actions.GenerateCreateStatement(query, scope, testCase.input)
			case proto.ActionType_ACTION_TYPE_UPDATE:
				statement, err = actions.GenerateUpdateStatement(query, scope, testCase.input)
			case proto.ActionType_ACTION_TYPE_DELETE:
				statement, err = actions.GenerateDeleteStatement(query, scope, testCase.input)
			default:
				require.NoError(t, fmt.Errorf("unhandled action type %s in sql generation", action.Type.String()))
			}

			if err != nil {
				require.NoError(t, err)
			}

			require.Equal(t, clean(testCase.expectedTemplate), clean(statement.SqlTemplate()))

			if testCase.expectedArgs != nil {
				if len(testCase.expectedArgs) != len(statement.SqlArgs()) {
					assert.Failf(t, "Argument count not matching", "Expected: %v, Actual: %v", len(testCase.expectedArgs), len(statement.SqlArgs()))
				} else {
					for i := 0; i < len(testCase.expectedArgs); i++ {
						if testCase.expectedArgs[i] != statement.SqlArgs()[i] {
							assert.Failf(t, "Arguments not matching", "SQL argument at index %d not matching. Expected: %v, Actual: %v", i, testCase.expectedArgs[i], statement.SqlArgs()[i])
							break
						}
					}
				}
			}
		})
	}
}

func generateGetStatement(query *actions.QueryBuilder, scope *actions.Scope, input map[string]any) (*actions.Statement, error) {

	typeProvider := NewTypeProvider(scope.Schema)

	env, err := cel.NewCustomEnv(
		KeelLib(),
		cel.CustomTypeProvider(typeProvider),
		cel.Declarations(
			decls.NewVar(strcase.ToLowerCamel(scope.Model.Name), decls.NewObjectType(scope.Model.Name)),
			decls.NewVar("ctx", decls.NewObjectType("Context")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("program setup err: %s", err)
	}

	for _, where := range scope.Action.WhereExpressions {

		ast, issues := env.Compile(where.Source)
		if issues != nil && issues.Err() != nil {
			return nil, issues.Err()
		}

		Build(ast, query, input)

		// expression, err := parser.ParseExpression(where.Source)
		// if err != nil {
		// 	return err
		// }

		// // Resolve the database statement for this expression
		// err = query.whereByExpression(scope, expression, args)
		// if err != nil {
		// 	return err
		// }

		// Where attributes are ANDed together
		query.And()
	}
	// Select all columns and distinct on id
	query.Select(actions.AllFields())
	query.DistinctOn(actions.IdField())

	return query.SelectStatement(), nil
}

// Generates a scope and query builder
func generateQueryScope(ctx context.Context, schemaString string, actionName string) (*actions.Scope, *actions.QueryBuilder, *proto.Action, error) {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(schemaString, config.Empty)
	if err != nil {
		return nil, nil, nil, err
	}

	action := schema.FindAction(actionName)
	if action == nil {
		return nil, nil, nil, fmt.Errorf("action not found in schema: %s", actionName)
	}

	model := schema.FindModel(action.ModelName)
	query := actions.NewQuery(model)
	scope := actions.NewScope(ctx, action, schema)

	return scope, query, action, nil
}

// Trims and removes redundant spacing and other characters
func clean(sql string) string {
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")
	sql = strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
	sql = strings.ReplaceAll(sql, "( ", "(")
	sql = strings.ReplaceAll(sql, " )", ")")
	return sql
}
