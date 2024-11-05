package expressions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

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
		name: "get_where_filter",
		keelSchema: `
			model Thing {
				fields {
					name Text
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
				statement, err = generateStatement(query, scope, testCase.input)
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

func generateStatement(query *actions.QueryBuilder, scope *actions.Scope, input map[string]any) (*actions.Statement, error) {
	err := query.ApplyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	for _, where := range scope.Action.WhereExpressions {
		// Replace the and/or operators
		expr := strings.ReplaceAll(where.Source, " or ", " || ")
		expr = strings.ReplaceAll(expr, " and ", " && ")

		parser, err := NewParser(scope.Schema, scope.Model)
		if err != nil {
			return nil, err
		}

		validateErrs, err := parser.Validate(expr, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
		if err != nil {
			return nil, err
		}
		if validateErrs != nil {
			return nil, errors.New(strings.Join(validateErrs, "\n"))
		}

		err = parser.Build(query, expr, input)
		if err != nil {
			return nil, err
		}

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
