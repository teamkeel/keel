package actions_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/schema"
)

type testCase struct {
	// Name given to the test case
	name string
	// Valid keel schema for this test case
	keelSchema string
	// Operation name to run test upon
	operationName string
	// Input map for operation
	input map[string]any
	// Expected SQL template generated (with ? placeholders for values)
	expectedTemplate string
	// OPTIONAL: Expected ordered argument slice
	expectedArgs []any
}

var testCases = []testCase{
	{
		name: "get_op_by_id",
		keelSchema: `
			model Thing {
				operations {
					get getThing(id)
				}
				@permission(expression: true, actions: [get])
			}`,
		operationName: "getThing",
		input:         map[string]any{"id": "123"},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*
			FROM
				"thing"
			WHERE
				"thing"."id" IS NOT DISTINCT FROM ?`,
		expectedArgs: []any{"123"},
	},
	{
		name: "get_op_by_id_where",
		keelSchema: `
			model Thing {
				fields {
					isActive Boolean
				}
				operations {
					get getThing(id) {
						@where(thing.isActive == true)
					}
				}
				@permission(expression: true, actions: [get])
			}`,
		operationName: "getThing",
		input:         map[string]any{"id": "123"},
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
	{
		name: "list_op_no_filter",
		keelSchema: `
			model Thing {
				operations {
					list listThings() 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			ORDER BY 
				"thing"."id" LIMIT 50`,
	},
	{
		name: "list_op_implicit_input_text_contains",
		keelSchema: `
			model Thing {
				fields {
					name Text
				}
				operations {
					list listThings(name) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"contains": "bob"}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE
				"thing"."name" LIKE ?
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{"%%bob%%"},
	},
	{
		name: "list_op_implicit_input_text_startsWith",
		keelSchema: `
			model Thing {
				fields {
					name Text
				}
				operations {
					list listThings(name) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"startsWith": "bob"}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE
				"thing"."name" LIKE ?
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{"bob%%"},
	},
	{
		name: "list_op_implicit_input_text_endsWith",
		keelSchema: `
			model Thing {
				fields {
					name Text
				}
				operations {
					list listThings(name) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"endsWith": "bob"}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE
				"thing"."name" LIKE ?
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{"%%bob"},
	},
	{
		name: "list_op_implicit_input_text_oneof",
		keelSchema: `
            model Thing {
                fields {
                    name Text
                }
                operations {
                    list listThings(name) 
                }
                @permission(expression: true, actions: [list])
            }`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"oneOf": []any{"bob", "dave", "adam", "pete"}}}},
		expectedTemplate: `
            SELECT 
                DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
            FROM 
                "thing" 
            WHERE
                "thing"."name" IN (?, ?, ?, ?)
            ORDER BY 
                "thing"."id" LIMIT 50`,
		expectedArgs: []any{"bob", "dave", "adam", "pete"},
	},
	{
		name: "list_op_implicit_input_enum_oneof",
		keelSchema: `
            model Thing {
                fields {
                    category Category
                }
                operations {
                    list listThings(category) 
                }
                @permission(expression: true, actions: [list])
            }
			enum Category {
				Technical
				Food
				Lifestyle
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"category": map[string]any{
					"oneOf": []any{"Technical", "Food"}}}},
		expectedTemplate: `
            SELECT 
                DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
            FROM 
                "thing" 
            WHERE
                "thing"."category" IN (?, ?)
            ORDER BY 
                "thing"."id" LIMIT 50`,
		expectedArgs: []any{"Technical", "Food"},
	},
	{
		name: "list_op_implicit_input_timestamp_after",
		keelSchema: `
			model Thing {
				operations {
					list listThings(createdAt) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"after": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE
				"thing"."created_at" > ? 
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)},
	},
	{
		name: "list_op_implicit_input_timestamp_onorafter",
		keelSchema: `
			model Thing {
				operations {
					list listThings(createdAt) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"onOrAfter": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE
				"thing"."created_at" >= ? 
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)},
	},
	{
		name: "list_op_implicit_input_timestamp_after",
		keelSchema: `
			model Thing {
				operations {
					list listThings(createdAt) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"before": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE
				"thing"."created_at" < ? 
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)},
	},
	{
		name: "list_op_implicit_input_timestamp_onorbefore",
		keelSchema: `
			model Thing {
				operations {
					list listThings(createdAt) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"onOrBefore": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE
				"thing"."created_at" <= ? 
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)},
	},
	{
		name: "list_op_expression_text_in",
		keelSchema: `
			model Thing {
				fields {
                    title Text
                }
				operations {
					list listThings() {
						@where(thing.title in ["title1", "title2"])
					} 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input:         map[string]any{},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE 
				"thing"."title" IN (?, ?)
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{"title1", "title2"},
	},
	{
		name: "list_op_expression_text_notin",
		keelSchema: `
			model Thing {
				fields {
                    title Text
                }
				operations {
					list listThings() {
						@where(thing.title not in ["title1", "title2"])
					} 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input:         map[string]any{},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE 
				"thing"."title" NOT IN (?, ?)
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{"title1", "title2"},
	},
	{
		name: "list_op_expression_number_in",
		keelSchema: `
			model Thing {
				fields {
                    age Number
                }
				operations {
					list listThings() {
						@where(thing.age in [10, 20])
					} 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input:         map[string]any{},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE 
				"thing"."age" IN (?, ?)
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{int64(10), int64(20)},
	},
	{
		name: "list_op_expression_number_notin",
		keelSchema: `
			model Thing {
				fields {
                    age Number
                }
				operations {
					list listThings() {
						@where(thing.age not in [10, 20])
					} 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input:         map[string]any{},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			WHERE 
				"thing"."age" NOT IN (?, ?)
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{int64(10), int64(20)},
	},
	{
		name: "list_op_implicit_input_on_nested_model",
		keelSchema: `
			model Parent {
				fields {
					name Text
				}
			}	
			model Thing {
				fields {
					parent Parent
				}
				operations {
					list listThings(parent.name) 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"parentName": map[string]any{
					"equals": "bob"}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			INNER JOIN 
				"parent" AS "thing$parent" 
					ON "thing$parent"."id" = "thing"."parent_id" 
			WHERE 
				"thing$parent"."name" IS NOT DISTINCT FROM ?
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{"bob"},
	},
	{
		name: "list_op_where_expression_on_nested_model",
		keelSchema: `
			model Parent {
				fields {
					name Text
					isActive Boolean
				}
			}	
			model Thing {
				fields {
					parent Parent
				}
				operations {
					list listThings() {
						@where(thing.parent.isActive == false)
					} 
				}
				@permission(expression: true, actions: [list])
			}`,
		operationName: "listThings",
		input: map[string]any{
			"where": map[string]any{}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing".id) OVER (ORDER BY "thing".id) IS NOT NULL THEN true ELSE false END AS hasNext 
			FROM 
				"thing" 
			INNER JOIN 
				"parent" AS "thing$parent" 
					ON "thing$parent"."id" = "thing"."parent_id" 
			WHERE 
				"thing$parent"."is_active" IS NOT DISTINCT FROM ?
			ORDER BY 
				"thing"."id" LIMIT 50`,
		expectedArgs: []any{false},
	},
	{
		name: "create_op_nested_model",
		keelSchema: `
			model Parent {
				fields {
					name Text
				}
			}	
			model Thing {
				fields {
					name Text
					age Number
					isActive Boolean
					parent Parent
				}
				operations {
					create createThing() with (name, age, parent.id) {
						@set(thing.isActive = true)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		operationName: "createThing",
		input: map[string]any{
			"name":     "bob",
			"age":      21,
			"parentId": "123",
		},
		expectedTemplate: `
			INSERT INTO "thing" 
				(age, created_at, id, is_active, name, parent_id, updated_at)
			VALUES 
				(?, ?, ?, ?, ?, ?, ?) 
			RETURNING 
				"thing".*`,
	},
	{
		name: "update_op_nested_model",
		keelSchema: `
			model Parent {
				fields {
					name Text
				}
			}	
			model Thing {
				fields {
					name Text
					age Number
					isActive Boolean
					parent Parent
				}
				operations {
					update updateThing(id) with (name, age, parent.id) {
						@set(thing.isActive = true)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		operationName: "updateThing",
		input: map[string]any{
			"where": map[string]any{
				"id": "789",
			},
			"values": map[string]any{
				"name":     "bob",
				"age":      21,
				"parentId": "123",
			},
		},
		expectedTemplate: `
			UPDATE 
				"thing" 
			SET 
				age = ?, 
				is_active = ?,
				name = ?, 
				parent_id = ?
			WHERE 
				"thing"."id" IS NOT DISTINCT FROM ? 
			RETURNING 
				"thing".*`,
		expectedArgs: []any{21, true, "bob", "123", "789"},
	},
	{
		name: "delete_op_by_id",
		keelSchema: `
			model Thing {
				operations {
					delete deleteThing(id)
				}
				@permission(expression: true, actions: [delete])
			}`,
		operationName: "deleteThing",
		input:         map[string]any{"id": "123"},
		expectedTemplate: `
			DELETE FROM 
				"thing" 
			WHERE 
				"thing"."id" IS NOT DISTINCT FROM ?
			RETURNING "thing"."id"`,
		expectedArgs: []any{"123"},
	},
}

func TestQueryBuilder(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			scope, query, operation, err := generateQueryScope(testCase.keelSchema, testCase.operationName)
			if err != nil {
				require.NoError(t, err)
			}

			var statement *actions.Statement
			switch operation.Type {
			case proto.OperationType_OPERATION_TYPE_GET:
				statement, err = actions.GenerateGetStatement(query, scope, testCase.input)
			case proto.OperationType_OPERATION_TYPE_LIST:
				statement, err = actions.GenerateListStatement(query, scope, testCase.input)
			case proto.OperationType_OPERATION_TYPE_CREATE:
				statement, err = actions.GenerateCreateStatement(query, scope, testCase.input)
			case proto.OperationType_OPERATION_TYPE_UPDATE:
				statement, err = actions.GenerateUpdateStatement(query, scope, testCase.input)
			case proto.OperationType_OPERATION_TYPE_DELETE:
				statement, err = actions.GenerateDeleteStatement(query, scope, testCase.input)
			default:
				require.NoError(t, fmt.Errorf("unhandled operation type %s in sql generation", operation.Type.String()))
			}

			if err != nil {
				require.NoError(t, err)
			}

			require.Equal(t, clean(testCase.expectedTemplate), clean(statement.SqlTemplate()))

			if testCase.expectedArgs != nil {
				require.True(t, reflect.DeepEqual(testCase.expectedArgs, statement.SqlArgs()), fmt.Sprintf("SQL arguments not equal. Expected: %v, Actual: %v", testCase.expectedArgs, statement.SqlArgs()))
			}
		})
	}
}

// Generates a scope and query builder
func generateQueryScope(schemaText string, operationName string) (*actions.Scope, *actions.QueryBuilder, *proto.Operation, error) {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(schemaText)
	if err != nil {
		return nil, nil, nil, err
	}

	operation := proto.FindOperation(schema, operationName)
	if operation == nil {
		return nil, nil, nil, fmt.Errorf("operation not found in schema: %s", operationName)
	}

	model := proto.FindModel(schema.Models, operation.ModelName)
	query := actions.NewQuery(model)
	scope := actions.NewScope(context.Background(), operation, schema)

	return scope, query, operation, nil
}

// Trims and removes redundant spacing
func clean(sql string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
}
