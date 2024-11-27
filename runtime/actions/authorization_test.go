package actions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
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
	identity  auth.Identity
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

var unverifiedIdentity = auth.Identity{
	"id":    "identityId",
	"email": "weaveton@weave.xyz",
}

var verifiedIdentity = auth.Identity{
	"id":            "identityId",
	"email":         "keelson@keel.xyz",
	"emailVerified": true,
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			WHERE
				("thing"."created_by_id" IS NOT DISTINCT FROM ?)
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "identity_check_by_id",
		keelSchema: `
			model Thing {
				fields {
					createdBy Identity
				}
				actions {
					list listThings() {
						@permission(expression: thing.createdBy.id == ctx.identity.id)
					}
				}
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT 
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM 
				"thing" 
			LEFT JOIN "identity" AS "thing$created_by" ON 
				"thing$created_by"."id" = "thing"."created_by_id" 
			WHERE 
				("thing$created_by"."id" IS NOT DISTINCT FROM ?)
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "identity_check_not_authenticated",
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			WHERE
				( "thing"."created_by_id" IS NOT DISTINCT FROM NULL )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		earlyAuth:    CouldNotAuthoriseEarly(),
		expectedArgs: []any{"idToAuthorise"},
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			LEFT JOIN
				"related" AS "thing$related"
			ON
				"thing$related"."id" = "thing"."related_id"
			WHERE
				( "thing$related"."created_by_id" IS NOT DISTINCT FROM ? )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			WHERE
				( "thing"."is_active" IS NOT DISTINCT FROM ? )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{true, "idToAuthorise"},
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			LEFT JOIN
				"related" AS "thing$related"
			ON
				"thing$related"."id" = "thing"."related_id"
			WHERE
				( "thing$related"."created_by_id" IS NOT DISTINCT FROM "thing"."created_by_id" )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		earlyAuth:    CouldNotAuthoriseEarly(),
		expectedArgs: []any{"idToAuthorise"},
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			WHERE
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? AND "thing"."created_by_id" IS NOT DISTINCT FROM ? ) )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{true, unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			WHERE
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? OR "thing"."created_by_id" IS NOT DISTINCT FROM ? ) )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{true, unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			WHERE
				( "thing"."is_active" IS NOT DISTINCT FROM ?
					OR
				 "thing"."created_by_id" IS NOT DISTINCT FROM ? )
				 AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{true, unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			LEFT JOIN "related" AS
				"thing$related" ON "thing$related"."id" = "thing"."related_id"
			WHERE
				( ( "thing"."is_active" IS NOT DISTINCT FROM ? AND "thing"."created_by_id" IS NOT DISTINCT FROM ? ) OR "thing"."created_by_id" IS NOT DISTINCT FROM "thing$related"."created_by_id" )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{true, unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM
				"thing"
			LEFT JOIN "identity" AS "thing$created_by" ON
				"thing$created_by"."id" = "thing"."created_by_id"
			WHERE
				( ( ? IS NOT DISTINCT FROM ? AND "thing$created_by"."id" IS NOT DISTINCT FROM ? ) )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{true, true, unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		identity:     unverifiedIdentity,
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
		identity:   unverifiedIdentity,
	},
	{
		name: "early_evaluate_multiple_attributes_with_database_multiple_attributes",
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
	},
	{
		name: "early_evaluate_inputs_authorised",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id, bool: Boolean) {
						@permission(expression: bool == true)
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123", "bool": true},
		earlyAuth:  AuthorisationGrantedEarly(),
		identity:   unverifiedIdentity,
	},
	{
		name: "early_evaluate_inputs_not_authorised",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id, bool: Boolean) {
						@permission(expression: bool == true)
					}
				}
			}`,
		actionName: "getThing",
		input:      map[string]any{"id": "123", "bool": false},
		earlyAuth:  AuthorisationDeniedEarly(),
		identity:   unverifiedIdentity,
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
		identity:   unverifiedIdentity,
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
				COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM 
				"thing" 
			LEFT JOIN 
				"identity" AS "thing$created_by" ON "thing$created_by"."id" = "thing"."created_by_id" 
			LEFT JOIN 
				"identity" AS "thing$updated_by" ON "thing$updated_by"."id" = "thing"."updated_by_id" 
			WHERE 
				( "thing$created_by"."id" IS NOT DISTINCT FROM ? OR "thing$updated_by"."id" IS NOT DISTINCT FROM ? )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		identity:     unverifiedIdentity,
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
				COUNT(DISTINCT "user"."id") = 1 AS authorised
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
				( ? IS NOT DISTINCT FROM "user$organisations$organisation$users$user"."identity_id" )
				AND "user"."id" = ANY(ARRAY[?]::TEXT[])
			`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		identity:     unverifiedIdentity,
	},
	{
		name: "identity_email_field_from_model",
		keelSchema: `
			model Thing {
				fields {
					identity Identity
				}
				actions {
					list list() {
						@permission(expression: thing.identity.email != "weaveton@weave.xyz")
					}
				}
			}`,
		actionName: "list",
		expectedTemplate: `
			SELECT COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM "thing" 
			LEFT JOIN "identity" AS "thing$identity" ON "thing$identity"."id" = "thing"."identity_id" 
			WHERE 
				( "thing$identity"."email" IS DISTINCT FROM ? )
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{"weaveton@weave.xyz", "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
	},
	{
		name: "identity_email_field_from_ctx",
		keelSchema: `
			model Thing {
				fields {
					identity Identity
				}
				actions {
					list list() {
						@permission(expression: ctx.identity.email != "weaveton@weave.xyz")
					}
				}
			}`,
		actionName: "list",
		expectedTemplate: `
			SELECT COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM "thing"
			WHERE (
				(SELECT "identity"."email" 
				FROM "identity" 
				WHERE 
					"identity"."id" IS NOT DISTINCT FROM ? AND 
					"identity"."email" IS DISTINCT FROM NULL) IS DISTINCT FROM ?)
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "weaveton@weave.xyz", "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "identity_email_field_from_ctx_not_authenticated",
		keelSchema: `
			model Thing {
				fields {
					identity Identity
				}
				actions {
					list list() {
						@permission(expression: ctx.identity.email != "weaveton@weave.xyz")
					}
				}
			}`,
		actionName: "list",
		expectedTemplate: `
			SELECT COUNT(DISTINCT "thing"."id") = 1 AS authorised
			FROM "thing"
			WHERE (
				(SELECT "identity"."email" 
				FROM "identity" 
				WHERE 
					"identity"."id" IS NOT DISTINCT FROM ? AND 
					"identity"."email" IS DISTINCT FROM NULL) IS DISTINCT FROM ?)
				AND "thing"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{"", "weaveton@weave.xyz", "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
	},
	{
		name: "identity_backlink_from_model",
		keelSchema: `
			model User {
				fields {
					isAdult Boolean
					identity Identity @unique
				}
			}
			model AdultFilm {
			}
			model Admit {
				fields {
					film AdultFilm
					identity Identity
				}
				actions {
					create admit() with (film.id, identity.id) {
						@permission(expression: admit.identity.user.isAdult)
					}
				}
			}`,
		actionName: "admit",
		expectedTemplate: `
			SELECT COUNT(DISTINCT "admit"."id") = 1 AS authorised
			FROM "admit" 
			LEFT JOIN "identity" AS "admit$identity" ON "admit$identity"."id" = "admit"."identity_id" 
			LEFT JOIN "user" AS "admit$identity$user" ON "admit$identity$user"."identity_id" = "admit$identity"."id" 
			WHERE 
				( "admit$identity$user"."is_adult" IS NOT DISTINCT FROM ? )
				AND "admit"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{true, "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
	},
	{
		name: "identity_backlink_from_ctx",
		keelSchema: `
			model User {
				fields {
					isAdult Boolean
					identity Identity @unique
				}
			}
			model AdultFilm {
				actions {
					get getFilm(id) {
						@permission(expression: ctx.identity.user.isAdult)
					}
				}
			}`,
		actionName: "getFilm",
		input:      map[string]any{"id": "123"},
		expectedTemplate: `
			SELECT COUNT(DISTINCT "adult_film"."id") = 1 AS authorised
			FROM "adult_film" 
			WHERE 
				((SELECT "identity$user"."is_adult" 
					FROM "identity" 
					LEFT JOIN "user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$user"."is_adult" IS DISTINCT FROM NULL) IS NOT DISTINCT FROM ?)
				AND "adult_film"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), true, "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "identity_backlink_ctx_compare_with_model_id",
		keelSchema: `
			model User {
				fields {
					isAdult Boolean
					identity Identity @unique
				}
			}
			model AdultFilm {
				fields {
					identity Identity @unique
				}
				actions {
					get getFilm(id) {
						@permission(expression: ctx.identity.user.id == adultFilm.identity.user.id)
					}
				}
			}`,
		actionName: "getFilm",
		input:      map[string]any{"id": "123"},
		expectedTemplate: `
			SELECT COUNT(DISTINCT "adult_film"."id") = 1 AS authorised
			FROM "adult_film" 
			LEFT JOIN "identity" AS "adult_film$identity" ON "adult_film$identity"."id" = "adult_film"."identity_id" 
			LEFT JOIN "user" AS "adult_film$identity$user" ON "adult_film$identity$user"."identity_id" = "adult_film$identity"."id" 
			WHERE 
				((SELECT "identity$user"."id" 
					FROM "identity" 
					LEFT JOIN "user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ?  AND "identity$user"."id" IS DISTINCT FROM NULL) IS NOT DISTINCT FROM "adult_film$identity$user"."id")
				AND "adult_film"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "identity_backlink_ctx_compare_with_model",
		keelSchema: `
			model User {
				fields {
					isAdult Boolean
					identity Identity @unique
				}
			}
			model AdultFilm {
				fields {
					identity Identity @unique
				}
				actions {
					get getFilm(id) {
						@permission(expression: ctx.identity.user == adultFilm.identity.user)
					}
				}
			}`,
		actionName: "getFilm",
		input:      map[string]any{"id": "123"},
		expectedTemplate: `
			SELECT COUNT(DISTINCT "adult_film"."id") = 1 AS authorised
			FROM "adult_film" 
			LEFT JOIN "identity" AS "adult_film$identity" ON "adult_film$identity"."id" = "adult_film"."identity_id" 
			LEFT JOIN "user" AS "adult_film$identity$user" ON "adult_film$identity$user"."identity_id" = "adult_film$identity"."id" 
			WHERE
				((SELECT "identity$user"."id" 
					FROM "identity" 
					LEFT JOIN "user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ?  AND "identity$user"."id" IS DISTINCT FROM NULL) IS NOT DISTINCT FROM "adult_film$identity$user"."id")
				AND "adult_film"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "identity_model_comparison",
		keelSchema: `
			model User {
				fields {
					identity Identity @unique
				}
			}
			model BankAccount {
				fields {
					identity Identity @unique
				}
				actions {
					get getBankAccount(id) {
						@permission(expression: ctx.identity == bankAccount.identity)
					}
				}
			}`,
		actionName: "getBankAccount",
		input:      map[string]any{"id": "123"},
		expectedTemplate: `
			SELECT 
				COUNT(DISTINCT "bank_account"."id") = 1 AS authorised
			FROM 
				"bank_account" 
			WHERE
				(? IS NOT DISTINCT FROM "bank_account"."identity_id")
				AND "bank_account"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "model_in_many_models",
		keelSchema: `
			model Entity {
				fields {
					name Text
					users EntityUser[]
				}
			}	
			model EntityUser {
				fields {
					identity Identity @unique @relation(user)
					entity Entity
				}
				actions {
					list entityUsers(entity.id) {
						@permission(expression: entityUser in ctx.identity.administrator.access.entity.users)
					}
				}
			}
			model EntityAccess {
				fields {
					entity Entity
					admin Administrator
				}
			}
			model Administrator {
				fields {
					access EntityAccess[]
					identity Identity @unique
				}
			}`,
		actionName: "entityUsers",
		input:      map[string]any{"entity": map[string]any{"id": map[string]any{"equals": "123"}}},
		expectedTemplate: `
			SELECT 
				COUNT(DISTINCT "entity_user"."id") = 1 AS authorised
			FROM 
				"entity_user" 
			WHERE
				("entity_user"."id" IN 
					(SELECT "identity$administrator$access$entity$users"."id" 
					FROM "identity" 
					LEFT JOIN "administrator" AS "identity$administrator" ON "identity$administrator"."identity_id" = "identity"."id" 
					LEFT JOIN "entity_access" AS "identity$administrator$access" ON "identity$administrator$access"."admin_id" = "identity$administrator"."id" 
					LEFT JOIN "entity" AS "identity$administrator$access$entity" ON "identity$administrator$access$entity"."id" = "identity$administrator$access"."entity_id" 
					LEFT JOIN "entity_user" AS "identity$administrator$access$entity$users" ON "identity$administrator$access$entity$users"."entity_id" = "identity$administrator$access$entity"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$administrator$access$entity$users"."id" IS DISTINCT FROM NULL))
				AND "entity_user"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
	{
		name: "model_id_in_many_models_ids",
		keelSchema: `
			model Entity {
				fields {
					name Text
					users EntityUser[]
				}
			}	
			model EntityUser {
				fields {
					identity Identity @unique @relation(user)
					entity Entity
				}
				actions {
					list entityUsers(entity.id) {
						@permission(expression: entityUser.id in ctx.identity.administrator.access.entity.users.id)
					}
				}
			}
			model EntityAccess {
				fields {
					entity Entity
					admin Administrator
				}
			}
			model Administrator {
				fields {
					access EntityAccess[]
					identity Identity @unique
				}
			}`,
		actionName: "entityUsers",
		input:      map[string]any{"entity": map[string]any{"id": map[string]any{"equals": "123"}}},
		expectedTemplate: `
			SELECT 
				COUNT(DISTINCT "entity_user"."id") = 1 AS authorised
			FROM 
				"entity_user"
			WHERE
				("entity_user"."id" IN 
					(SELECT "identity$administrator$access$entity$users"."id" 
					FROM "identity" 
					LEFT JOIN "administrator" AS "identity$administrator" ON "identity$administrator"."identity_id" = "identity"."id" 
					LEFT JOIN "entity_access" AS "identity$administrator$access" ON "identity$administrator$access"."admin_id" = "identity$administrator"."id" 
					LEFT JOIN "entity" AS "identity$administrator$access$entity" ON "identity$administrator$access$entity"."id" = "identity$administrator$access"."entity_id" 
					LEFT JOIN "entity_user" AS "identity$administrator$access$entity$users" ON "identity$administrator$access$entity$users"."entity_id" = "identity$administrator$access$entity"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$administrator$access$entity$users"."id" IS DISTINCT FROM NULL))
				AND "entity_user"."id" = ANY(ARRAY[?]::TEXT[])`,
		expectedArgs: []any{unverifiedIdentity[parser.FieldNameId].(string), "idToAuthorise"},
		earlyAuth:    CouldNotAuthoriseEarly(),
		identity:     unverifiedIdentity,
	},
}

func TestPermissionQueryBuilder(t *testing.T) {
	t.Parallel()
	for _, tc := range authorisationTestCases {
		testCase := tc

		// if testCase.name != "early_evaluate_list_op" {
		// 	continue
		// }

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			if testCase.identity != nil {
				ctx = auth.WithIdentity(ctx, testCase.identity)
			}

			ctx = runtimectx.WithSecrets(ctx, map[string]string{"MY_SECRET": "1234"})

			scope, _, _, err := generateQueryScope(ctx, testCase.keelSchema, testCase.actionName)
			if err != nil {
				require.NoError(t, err)
			}

			permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

			canResolveEarly, authorised, err := actions.TryResolveAuthorisationEarly(scope, testCase.input, permissions)
			if err != nil {
				require.NoError(t, err)
			}

			if canResolveEarly {
				require.NotNil(t, testCase.earlyAuth, "earlyAuth is CouldNotAuthoriseEarly(), but authorisation was determined early.")
				if testCase.earlyAuth != nil {
					if authorised {
						require.Equal(t, testCase.earlyAuth.authorised, true, "earlyAuth is AuthorisationDeniedEarly(). Expected AuthorisationGrantedEarly().")
					} else {
						require.Equal(t, testCase.earlyAuth.authorised, false, "earlyAuth is AuthorisationGrantedEarly(). Expected AuthorisationDeniedEarly().")
					}
				}
			} else {
				require.Nil(t, testCase.earlyAuth, "earlyAuth should be CouldNotAuthoriseEarly() because authorised could not be determined early.")
			}

			if !canResolveEarly {
				permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

				statement, err := actions.GeneratePermissionStatement(scope, permissions, testCase.input, []string{"idToAuthorise"})
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
