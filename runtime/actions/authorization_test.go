package actions_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

type authorisationTestCase struct {
	// Name given to the test case
	name string
	// Valid keel schema for this test case
	keelSchema string
	// Operation name to run test upon
	operationName string
	// Expected SQL template generated (with ? placeholders for values)
	expectedTemplate string
	// OPTIONAL: Expected ordered argument slice
	expectedArgs []any
}

var identity = &runtimectx.Identity{
	Id:    "identityId",
	Email: "keelson@keel.xyz",
}

var rowsToAuthorise = []map[string]any{
	{"id": "id1"},
	{"id": "id2"},
	{"id": "id3"},
}

var idsToAuthorise = lo.Map(rowsToAuthorise, func(row map[string]any, _ int) any { return row["id"] })

var authorisationTestCases = []authorisationTestCase{
	{
		name: "identity_check",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					list listThings() {
						@permission(expression: thing.createdBy == ctx.identity)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing"  
			WHERE 
				( "thing"."created_by_id" IS NOT DISTINCT FROM ? ) 
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{identity.Id}, idsToAuthorise...),
	},
	{
		name: "identity_on_related_model",
		keelSchema: `
			model Related {
				fields {
					createdBy Identity
				}
			}
			model Thing {
				fields {
					related Related
				}
				operations {
					list listThings() {
						@permission(expression: thing.related.createdBy == ctx.identity)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing"  
			INNER JOIN 
				"related" AS "thing$related" 
			ON 
				"thing$related"."id" = "thing"."related_id"
			WHERE 
				( "thing$related"."created_by_id" IS NOT DISTINCT FROM ? ) 
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{identity.Id}, idsToAuthorise...),
	},
	{
		name: "field_with_literal",
		keelSchema: `
			model Thing {
				fields {
					isActive Boolean
					createdBy Identity
				}
				operations {
					list listThings() {
						@permission(expression: thing.isActive == true)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing"  
			WHERE 
				( "thing"."is_active" IS NOT DISTINCT FROM ? ) 
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{true}, idsToAuthorise...),
	},
	{
		name: "field_with_related_field",
		keelSchema: `
			model Related {
				fields {
					createdBy Identity
				}
			}
			model Thing {
				fields {
					createdBy Identity
					related Related
				}
				operations {
					list listThings() {
						@permission(expression: thing.related.createdBy == thing.createdBy)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing"  
			INNER JOIN 
				"related" AS "thing$related" 
			ON 
				"thing$related"."id" = "thing"."related_id"
			WHERE 
				( "thing$related"."created_by_id" IS NOT DISTINCT FROM "thing"."created_by_id" )
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: idsToAuthorise,
	},
	{
		name: "multiple_conditions_and",
		keelSchema: `
			model Thing {
				fields {
					isActive Boolean
					createdBy Identity
				}
				operations {
					list listThings() {
						@permission(expression: thing.isActive == true and thing.createdBy == ctx.identity)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing"  
			WHERE 
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? AND "thing"."created_by_id" IS NOT DISTINCT FROM ? ) ) 
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{true, identity.Id}, idsToAuthorise...),
	},
	{
		name: "multiple_conditions_or",
		keelSchema: `
			model Thing {
				fields {
					isActive Boolean
					createdBy Identity
				}
				operations {
					list listThings() {
						@permission(expression: thing.isActive == true or thing.createdBy == ctx.identity)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing"  
			WHERE 
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? OR "thing"."created_by_id" IS NOT DISTINCT FROM ? ) ) 
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{true, identity.Id}, idsToAuthorise...),
	},
	{
		name: "multiple_permission_attributes",
		keelSchema: `
			model Thing {
				fields {
					isActive Boolean
					createdBy Identity
				}
				operations {
					list listThings() {
						@permission(expression: thing.isActive == true)
						@permission(expression: thing.createdBy == ctx.identity)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing"  
			WHERE 
				( "thing"."is_active" IS NOT DISTINCT FROM ?
					OR
				 "thing"."created_by_id" IS NOT DISTINCT FROM ? ) 
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{true, identity.Id}, idsToAuthorise...),
	},
	{
		name: "multiple_permission_attributes_with_multiple_conditions",
		keelSchema: `
			model Related {
				fields {
					createdBy Identity
				}
			}
			model Thing {
				fields {
					isActive Boolean
					related Related
					createdBy Identity
				}
				operations {
					list listThings() {
						@permission(expression: thing.isActive == true and thing.createdBy == ctx.identity)
						@permission(expression: thing.createdBy == thing.related.createdBy)
					}
				}
			}`,
		operationName: "listThings",
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing" 
			INNER JOIN "related" AS 
				"thing$related" ON "thing$related"."id" = "thing"."related_id" 
			WHERE 
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? AND "thing"."created_by_id" IS NOT DISTINCT FROM ? ) OR "thing"."created_by_id" IS NOT DISTINCT FROM "thing$related"."created_by_id" ) 
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{true, identity.Id}, idsToAuthorise...),
	},
}

func TestPermissionQueryBuilder(t *testing.T) {
	for _, testCase := range authorisationTestCases {
		t.Run(testCase.name, func(t *testing.T) {

			ctx := context.Background()
			ctx = runtimectx.WithIdentity(ctx, identity)

			scope, _, _, err := generateQueryScope(ctx, testCase.keelSchema, testCase.operationName)
			if err != nil {
				require.NoError(t, err)
			}

			statement, err := actions.GeneratePermissionStatement(scope, rowsToAuthorise)
			if err != nil {
				require.NoError(t, err)
			}

			require.Equal(t, clean(testCase.expectedTemplate), clean(statement.SqlTemplate()))

			if testCase.expectedArgs != nil {
				for i := 1; i < len(testCase.expectedArgs); i++ {
					if testCase.expectedArgs[i] != ignore && testCase.expectedArgs[i] != statement.SqlArgs()[i] {
						assert.Failf(t, "Arguments not matching", "SQL argument at index %d not matching. Expected: %v, Actual: %v", i, testCase.expectedArgs[i], statement.SqlArgs()[i])
						break
					}
				}

				if len(testCase.expectedArgs) != len(statement.SqlArgs()) {
					assert.Failf(t, "Argument count not matching", "Expected: %v, Actual: %v", len(testCase.expectedArgs), len(statement.SqlArgs()))
				}
			}

		})
	}
}
