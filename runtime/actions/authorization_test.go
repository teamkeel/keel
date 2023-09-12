package actions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
)

type authorisationTestCase struct {
	// Name given to the test case
	name string
	// Valid keel schema for this test case
	keelSchema string
	// Action name to run test upon
	actionName string
	// Input map for operation
	input map[string]any
	// Expected SQL template generated (with ? placeholders for values)
	expectedTemplate string
	// OPTIONAL: Expected ordered argument slice
	expectedArgs []any
	// If resolved early, what was the authorisation result?
	// nil if early authorisation cannot be determined.
	earlyAuth *earlyAuthorisationResult
	identity  *auth.Identity
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

var identity = &auth.Identity{
	Id:    "identityId",
	Email: "keelson@keel.xyz",
}

var verifiedIdentity = &auth.Identity{
	Id:            "identityId",
	Email:         "keelson@keel.xyz",
	EmailVerified: true,
}

var authorisationTestCases = []authorisationTestCase{
	{
		name: "identity_check",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				actions {
					list listThings() {
						@permission(expression: thing.createdBy == ctx.identity)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			WHERE
				( "thing"."created_by_id" IS NOT DISTINCT FROM ? )`,
		expectedArgs: []any{identity.Id},
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
				actions {
					list listThings() {
						@permission(expression: thing.related.createdBy == ctx.identity)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			LEFT JOIN
				"related" AS "thing$related"
			ON
				"thing$related"."id" = "thing"."related_id"
			WHERE
				( "thing$related"."created_by_id" IS NOT DISTINCT FROM ? )`,
		expectedArgs: []any{identity.Id},
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
				actions {
					list listThings() {
						@permission(expression: thing.isActive == true)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			WHERE
				( "thing"."is_active" IS NOT DISTINCT FROM ? )`,
		expectedArgs: []any{true},
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
				actions {
					list listThings() {
						@permission(expression: thing.related.createdBy == thing.createdBy)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			LEFT JOIN
				"related" AS "thing$related"
			ON
				"thing$related"."id" = "thing"."related_id"
			WHERE
				( "thing$related"."created_by_id" IS NOT DISTINCT FROM "thing"."created_by_id" )`,
		earlyAuth: CouldNotAuthoriseEarly(),
	},
	{
		name: "multiple_conditions_and",
		keelSchema: `
			model Thing {
				fields {
					isActive Boolean
					createdBy Identity
				}
				actions {
					list listThings() {
						@permission(expression: thing.isActive == true and thing.createdBy == ctx.identity)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			WHERE
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? AND "thing"."created_by_id" IS NOT DISTINCT FROM ? ) )`,
		expectedArgs: []any{true, identity.Id},
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
				actions {
					list listThings() {
						@permission(expression: thing.isActive == true or thing.createdBy == ctx.identity)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			WHERE
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? OR "thing"."created_by_id" IS NOT DISTINCT FROM ? ) )`,
		expectedArgs: []any{true, identity.Id},
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
				actions {
					list listThings() {
						@permission(expression: thing.isActive == true)
						@permission(expression: thing.createdBy == ctx.identity)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			WHERE
				( "thing"."is_active" IS NOT DISTINCT FROM ?
					OR
				 "thing"."created_by_id" IS NOT DISTINCT FROM ? )`,
		expectedArgs: []any{true, identity.Id},
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
				actions {
					list listThings() {
						@permission(expression: thing.isActive == true and thing.createdBy == ctx.identity)
						@permission(expression: thing.createdBy == thing.related.createdBy)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			LEFT JOIN "related" AS
				"thing$related" ON "thing$related"."id" = "thing"."related_id"
			WHERE
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? AND "thing"."created_by_id" IS NOT DISTINCT FROM ? ) OR "thing"."created_by_id" IS NOT DISTINCT FROM "thing$related"."created_by_id" )`,
		expectedArgs: []any{true, identity.Id},
		earlyAuth:    CouldNotAuthoriseEarly(),
	},
	{
		name: "early_evaluate_create_op",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				actions {
					create createThing() {
						@set(thing.createdBy.id = ctx.identity.id)
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "createThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_get_op",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_update_op",
		keelSchema: `
			model Thing {
				actions {
					update updateThing(id) {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "updateThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_list_op",
		keelSchema: `
			model Thing {
				actions {
					list listThing() {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "listThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_delete_op",
		keelSchema: `
			model Thing {
				actions {
					delete deleteThing(id) {
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "deleteThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_isauth_lhs",
		keelSchema: `
			model Thing {
				actions {
					create createThing() {
						@permission(expression: ctx.isAuthenticated == false)
					}
				}
			}`,
		actionName: "createThing",
		earlyAuth:  AuthorisationDeniedEarly(),
	},
	{
		name: "early_evaluate_isauth_rhs",
		keelSchema: `
			model Thing {
				actions {
					create createThing() {
						@permission(expression: false == ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "createThing",
		earlyAuth:  AuthorisationDeniedEarly(),
	},
	{
		name: "cannot_early_evaluate_multiple_conditions_and_with_database",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated and thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  CouldNotAuthoriseEarly(),
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing"."id"
			FROM
				"thing"
			LEFT JOIN "identity" AS "thing$created_by" ON
				"thing$created_by"."id" = "thing"."created_by_id"
			WHERE 
				"thing"."id" IS NOT DISTINCT FROM ? AND 
				( ( ? IS NOT DISTINCT FROM ? AND "thing$created_by"."id" IS NOT DISTINCT FROM ? ) )`,
		expectedArgs: []any{"123", true, true, identity.Id},
	},
	{
		name: "early_evaluate_multiple_conditions_or_with_database",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated or thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_attributes_with_database",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_attributes_authorised",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated)
						@permission(expression: ctx.isAuthenticated == false)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_and_conditions_authorised",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated and ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_or_conditions_authorised",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated or ctx.isAuthenticated)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
	},
	{
		name: "early_evaluate_multiple_and_conditions_not_authorised",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id) {
						@permission(expression: ctx.isAuthenticated and false)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationDeniedEarly(),
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
				actions {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   verifiedIdentity,
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
				actions {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationDeniedEarly(),
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
				actions {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   verifiedIdentity,
	},
	{
		name: "early_evaluate_roles_email_net_verified_not_authorised",
		keelSchema: `
			role Admin {
				emails {
					"keelson@keel.xyz"
				}
			}
			model Thing {
				actions {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationDeniedEarly(),
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
				actions {
					get getThing(id) {
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationDeniedEarly(),
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
				actions {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   verifiedIdentity,
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
				actions {
					get getThing(id) {
						@permission(expression: true)
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
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
				actions {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  CouldNotAuthoriseEarly(),
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
				actions {
					get getThing(id) {
						@permission(expression: false)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  CouldNotAuthoriseEarly(),
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
				actions {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  CouldNotAuthoriseEarly(),
	},
	{
		name: "not_verified",
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
				actions {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  CouldNotAuthoriseEarly(),
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
				actions {
					get getThing(id) {
						@permission(expression: false)
						@permission(roles: [Admin])
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   verifiedIdentity,
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
				actions {
					get getThing(id) {
						@permission(expression: false)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
						@permission(roles: [Admin])
						
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   verifiedIdentity,
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
				actions {
					get getThing(id) {
						@permission(roles: [Admin])
						@permission(expression: false)
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   verifiedIdentity,
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
				actions {
					get getThing(id) {
						@permission(expression: thing.createdBy.id == ctx.identity.id)
						@permission(expression: false)
						@permission(roles: [Admin])
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   verifiedIdentity,
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
				actions {
					get getThing(id) {
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
				@permission(expression: true, actions: [get])
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  CouldNotAuthoriseEarly(),
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
				actions {
					get getThing(id) {
						@permission(expression: true)
					}
				}
				@permission(expression: thing.createdBy.id == ctx.identity.id, actions: [get])
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  AuthorisationGrantedEarly(),
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
				actions {
					get getThing(id) {
						@permission(expression: false)
					}
				}
				@permission(expression: thing.createdBy.id == ctx.identity.id, actions: [get])
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  AuthorisationDeniedEarly(),
	},
	{
		name: "multiple_model_level_permissions_ored",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
					updatedBy Identity
				}
				actions {
					get getThing(id)
				}
				@permission(expression: thing.createdBy.id == ctx.identity.id, actions: [get])
				@permission(expression: thing.updatedBy.id == ctx.identity.id, actions: [get])
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123"},
		earlyAuth:  CouldNotAuthoriseEarly(),
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing"."id" 
			FROM 
				"thing" 
			LEFT JOIN 
				"identity" AS "thing$created_by" ON "thing$created_by"."id" = "thing"."created_by_id" 
			LEFT JOIN 
				"identity" AS "thing$updated_by" ON "thing$updated_by"."id" = "thing"."updated_by_id" 
			WHERE 
				"thing"."id" IS NOT DISTINCT FROM ? AND 
				( "thing$created_by"."id" IS NOT DISTINCT FROM ? OR "thing$updated_by"."id" IS NOT DISTINCT FROM ? )`,
		expectedArgs: []any{"123", identity.Id, identity.Id},
	},
	{
		name: "filters_and_permissions_and_relationships",
		keelSchema: `
			model User {
				fields {
					identity Identity @unique
					organisations UserOrganisation[]
				}
			
				actions {
					list listUsersByOrganisation(organisations.organisation.id) {
						@permission(expression: ctx.identity in user.organisations.organisation.users.user.identity)
					}
				}
			}
			model Organisation {
				fields {
					users UserOrganisation[]
				}
			}
			model UserOrganisation {
				fields {
					user User
					organisation Organisation
				}
			}`,
		actionName: "listUsersByOrganisation",
		input: map[string]any{
			"organisations": map[string]any{
				"organisation": map[string]any{
					"id": map[string]any{
						"equals": "123"}}}},
		earlyAuth: CouldNotAuthoriseEarly(),
		expectedTemplate: `
			SELECT 
				DISTINCT ON("user"."id") "user"."id" 
			FROM 
				"user" 
			LEFT JOIN 
				"user_organisation" AS "user$organisations" ON "user$organisations"."user_id" = "user"."id" 
			LEFT JOIN 
				"organisation" AS "user$organisations$organisation" ON "user$organisations$organisation"."id" = "user$organisations"."organisation_id" 
			LEFT JOIN 
				"user_organisation" AS "user$organisations$organisation$users" ON "user$organisations$organisation$users"."organisation_id" = "user$organisations$organisation"."id" 
			LEFT JOIN 
				"user" AS "user$organisations$organisation$users$user" ON "user$organisations$organisation$users$user"."id" = "user$organisations$organisation$users"."user_id" 
			WHERE 
				"user$organisations$organisation"."id" IS NOT DISTINCT FROM ? AND ( ? IS NOT DISTINCT FROM "user$organisations$organisation$users$user"."identity_id" )
			`,
		expectedArgs: []any{"123", identity.Id},
	},
}

func TestPermissionQueryBuilder(t *testing.T) {
	for _, testCase := range authorisationTestCases {
		t.Run(testCase.name, func(t *testing.T) {

			activeIdentity := identity

			if testCase.identity != nil {
				activeIdentity = testCase.identity
			}

			ctx := context.Background()
			ctx = auth.WithIdentity(ctx, activeIdentity)

			scope, _, _, err := generateQueryScope(ctx, testCase.keelSchema, testCase.actionName)
			if err != nil {
				require.NoError(t, err)
			}

			permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

			canResolveEarly, authorised, err := actions.TryResolveAuthorisationEarly(scope, permissions)
			if err != nil {
				require.NoError(t, err)
			}

			if canResolveEarly {
				require.NotNil(t, testCase.earlyAuth, "earlyAuth is CouldNotAuthoriseEarly(), but authorised was determined early. Expected earlyAuthorisationResult.")
				require.Equal(t, testCase.earlyAuth.authorised, authorised, "earlyAuth.authorised not matching. Expected: %v, Actual: %v", testCase.earlyAuth.authorised, authorised)
			} else {
				require.Nil(t, testCase.earlyAuth, "earlyAuth should be CouldNotAuthoriseEarly() because authorised could not be determined given early. Expected nil.")
			}

			if !canResolveEarly {
				permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

				statement, err := actions.GeneratePermissionStatement(scope, permissions, testCase.input)
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
