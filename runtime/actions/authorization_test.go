package actions_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
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
	// If resolved early, what was the authorisation result?
	// nil if early authorisation cannot be determined.
	earlyAuth *earlyAuthorisationResult
}

type earlyAuthorisationResult struct {
	authorised bool
}

func CouldNotAuthoriseEarly() *earlyAuthorisationResult {
	return nil
}

func AuthorisationGrantedEarly() *earlyAuthorisationResult {
	return &earlyAuthorisationResult{authorised: true}
}

func AuthorisationDeniedEarly() *earlyAuthorisationResult {
	return &earlyAuthorisationResult{authorised: false}
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
		earlyAuth:    CouldNotAuthoriseEarly(),
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
		earlyAuth:    CouldNotAuthoriseEarly(),
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
		earlyAuth:    CouldNotAuthoriseEarly(),
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
		earlyAuth:    CouldNotAuthoriseEarly(),
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
		earlyAuth:    CouldNotAuthoriseEarly(),
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
		earlyAuth:    CouldNotAuthoriseEarly(),
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
		earlyAuth:    CouldNotAuthoriseEarly(),
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
		earlyAuth:    CouldNotAuthoriseEarly(),
	},
	{
		name: "early_evaluate_create_op",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					create createThing() {
						@set(thing.createdBy.id = ctx.identity.id)
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "createThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_get_op",
		keelSchema: `
			model Thing {
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_update_op",
		keelSchema: `
			model Thing {
				operations {
					update updateThing(id) {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "updateThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_list_op",
		keelSchema: `
			model Thing {
				operations {
					list listThing() {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "listThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_delete_op",
		keelSchema: `
			model Thing {
				operations {
					delete deleteThing(id) {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "deleteThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_isauth_lhs",
		keelSchema: `
			model Thing {
				operations {
					create createThing() {
						@permission(expression: ctx.isAuthenticated == false)
					}
				}
			}`,
		operationName: "createThing",
		earlyAuth:     AuthorisationDeniedEarly(),
	},
	{
		name: "early_evaluate_isauth_rhs",
		keelSchema: `
			model Thing {
				operations {
					create createThing() {
						@permission(expression: false == ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "createThing",
		earlyAuth:     AuthorisationDeniedEarly(),
	},
	{
		name: "cannot_early_evaluate_multiple_conditions_and_with_database",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated and thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     CouldNotAuthoriseEarly(),
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			INNER JOIN "identity" AS "thing$created_by" ON
				"thing$created_by"."id" = "thing"."created_by_id"
			WHERE
				( ( ? IS NOT DISTINCT FROM ? AND "thing$created_by"."id" IS NOT DISTINCT FROM ? ) )
				AND "thing"."id" IN (?, ?, ?)`,
		expectedArgs: append([]any{true, true, identity.Id}, idsToAuthorise...),
	},
	{
		name: "early_evaluate_multiple_conditions_or_with_database",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated or thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_attributes_with_database",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_attributes_authorised",
		keelSchema: `
			model Thing {
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated)
						@permission(expression: ctx.isAuthenticated == false)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_and_conditions_authorised",
		keelSchema: `
			model Thing {
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated and ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_or_conditions_authorised",
		keelSchema: `
			model Thing {
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated or ctx.isAuthenticated)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_and_conditions_not_authorised",
		keelSchema: `
			model Thing {
				operations {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated and false)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationDeniedEarly(),
	},
	{
		name: "early_evaluate_roles_domain_authorised",
		keelSchema: `
			role Admin {
				domains {
					"keel.xyz"
				}
			}
			model Thing {
				operations {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_roles_domain_not_authorised",
		keelSchema: `
			role Admin {
				domains {
					"gmail.com"
				}
			}
			model Thing {
				operations {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationDeniedEarly(),
	},
	{
		name: "early_evaluate_roles_email_authorised",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				operations {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_roles_email_not_authorised",
		keelSchema: `
			role Admin {
				emails {
					"keelson@gmail.com"
				}
			}
			model Thing {
				operations {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationDeniedEarly(),
	},
	{
		name: "early_evaluate_passed_role_and_failed_permissions_authorised",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				operations {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_failed_role_and_passed_permissions_authorised",
		keelSchema: `
			role Admin {
				emails {
					"keelson@gmail.com"
				}
			}
			model Thing {
				operations {
					get getThing(id) {
						@permission(expression: true)
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "cannot_early_evaluate_failed_role_and_failed_permissions_and_database",
		keelSchema: `
			role Admin {
				emails {
					"keelson@gmail.com"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     CouldNotAuthoriseEarly(),
	},
	{
		name: "cannot_early_evaluate_failed_role_and_failed_permissions_and_database_2",
		keelSchema: `
			role Admin {
				emails {
					"keelson@gmail.com"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: false)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     CouldNotAuthoriseEarly(),
	},
	{
		name: "cannot_early_evaluate_failed_role_and_failed_permissions_and_database_3",
		keelSchema: `
			role Admin {
				emails {
					"keelson@gmail.com"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     CouldNotAuthoriseEarly(),
	},
	{
		name: "can_early_evaluate_mixed_permissions_authorised",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "can_early_evaluate_mixed_permissions_authorised_2",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: false)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
						@permission(roles: [Admin])
						
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "can_early_evaluate_mixed_permissions_authorised_3",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(roles: [Admin])
						@permission(expression: false)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "can_early_evaluate_mixed_permissions_authorised_4",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: thing.createdBy.id == ctx.identity.id)
						@permission(expression: false)
						@permission(roles: [Admin])
					}
				}
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "cannot_early_evaluate_op_level_permissions",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
				@permission(expression: true, actions: [get])
			}`,
		operationName: "getThing",
		earlyAuth:     CouldNotAuthoriseEarly(),
	},
	{
		name: "can_early_evaluate_op_level_permissions_granted",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: true)
					}
				}
				@permission(expression: thing.createdBy.id == ctx.identity.id, actions: [get])
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationGrantedEarly(),
	},
	{
		name: "can_early_evaluate_op_level_permissions_denied",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				fields {
					createdBy Identity
				}
				operations {
					get getThing(id) {
						@permission(expression: false)
					}
				}
				@permission(expression: thing.createdBy.id == ctx.identity.id, actions: [get])
			}`,
		operationName: "getThing",
		earlyAuth:     AuthorisationDeniedEarly(),
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

			permissions := proto.PermissionsForAction(scope.Schema, scope.Operation)

			canResolveEarly, authorised, err := actions.TryResolveAuthorisationEarly(scope, permissions)
			if err != nil {
				require.NoError(t, err)
			}

			if canResolveEarly {
				require.NotNil(t, testCase.earlyAuth, "earlyAuthorisationResult is nil, but authorised was determined early. Expected earlyAuthorisationResult.")
				require.Equal(t, testCase.earlyAuth.authorised, authorised, "earlyAuthorisationResult.authorised not matching. Expected: %v, Actual: %v", testCase.earlyAuth.authorised, authorised)
			} else {
				require.Nil(t, testCase.earlyAuth, "earlyAuthorisationResult should be nil because authorised could not be determined given early. Expected nil.")
			}

			if !canResolveEarly {
				permissions := proto.PermissionsForAction(scope.Schema, scope.Operation)

				statement, err := actions.GeneratePermissionStatement(scope, permissions, rowsToAuthorise)
				if err != nil {
					require.NoError(t, err)
				}

				if testCase.expectedTemplate != "" {
					require.Equal(t, clean(testCase.expectedTemplate), clean(statement.SqlTemplate()))
				}

				if testCase.expectedArgs != nil {
					for i := 0; i < len(testCase.expectedArgs); i++ {
						if testCase.expectedArgs[i] != statement.SqlArgs()[i] {
							assert.Failf(t, "Arguments not matching", "SQL argument at index %d not matching. Expected: %v, Actual: %v", i, testCase.expectedArgs[i], statement.SqlArgs()[i])
							break
						}
					}

					if len(testCase.expectedArgs) != len(statement.SqlArgs()) {
						assert.Failf(t, "Argument count not matching", "Expected: %v, Actual: %v", len(testCase.expectedArgs), len(statement.SqlArgs()))
					}
				}
			}
		})
	}
}
