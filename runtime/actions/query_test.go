package actions_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/trace"
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

var identity = auth.Identity{
	"id":    "identityId",
	"email": "keelson@keel.xyz",
}

var testCases = []testCase{
	{
		name: "get_op_by_id",
		keelSchema: `
			model Thing {
				actions {
					get getThing(id)
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
	{
		name: "create_op_default_attribute",
		keelSchema: `
			model Person {
				fields {
					name Text @default("Bob")
					age Number @default(100)
					isActive Boolean @default(true)
				}
				actions {
					create createPerson()
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createPerson",
		input:      map[string]any{},
		expectedTemplate: `
			WITH
				new_1_person AS
					(INSERT INTO "person"
					DEFAULT VALUES
					RETURNING *)
			SELECT * FROM new_1_person`,
		expectedArgs: []any{},
	},
	{
		name: "create_op_set_attribute_literals",
		keelSchema: `
			model Person {
				fields {
					name Text
					age Number
					isActive Boolean
					bio Markdown
				}
				actions {
					create createPerson() {
						@set(person.name = "Bob")
						@set(person.age = 100)
						@set(person.isActive = true)
						@set(person.bio = "# Biography")
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createPerson",
		input:      map[string]any{},
		expectedTemplate: `
			WITH
				new_1_person AS
					(INSERT INTO "person"
						(age, bio, is_active, name)
					VALUES
						(?, ?, ?, ?)
					RETURNING *)
			SELECT * FROM new_1_person`,
		expectedArgs: []any{int64(100), "# Biography", true, "Bob"},
	},
	{
		name: "create_op_set_attribute_context_identity_id",
		keelSchema: `
			model Person {
				fields {
					mainIdentity Identity
				}
				actions {
					create createPerson() {
						@set(person.mainIdentity.id = ctx.identity.id)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createPerson",
		input:      map[string]any{},
		expectedTemplate: `
			WITH
				new_1_person AS
					(INSERT INTO "person"
						(main_identity_id)
					VALUES
						(?)
					RETURNING *)
			SELECT *, set_identity_id(?) AS __keel_identity_id FROM new_1_person`,
		identity:     identity,
		expectedArgs: []any{"identityId", identity[parser.FieldNameId].(string)},
	},
	{
		name: "create_op_set_attribute_context_identity",
		keelSchema: `
			model Person {
				fields {
					mainIdentity Identity
				}
				actions {
					create createPerson() {
						@set(person.mainIdentity = ctx.identity)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createPerson",
		input:      map[string]any{},
		expectedTemplate: `
			WITH
				new_1_person AS
					(INSERT INTO "person"
						(main_identity_id)
					VALUES
						(?)
					RETURNING *)
			SELECT *, set_identity_id(?) AS __keel_identity_id FROM new_1_person`,
		identity:     identity,
		expectedArgs: []any{"identityId", identity[parser.FieldNameId].(string)},
	},
	{
		name: "create_op_set_attribute_input",
		keelSchema: `
			model Person {
				fields {
					name Text
					nickName Text
				}
				actions {
					create createPerson() with (name) {
						@set(person.nickName = name)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createPerson",
		input:      map[string]any{"name": "Dave"},
		expectedTemplate: `
			WITH
				new_1_person AS
					(INSERT INTO "person"
						(name, nick_name)
					VALUES
						(?, ?)
					RETURNING *)
			SELECT *, set_identity_id(?) AS __keel_identity_id FROM new_1_person`,
		identity:     identity,
		expectedArgs: []any{"Dave", "Dave", identity[parser.FieldNameId].(string)},
	},
	{
		name: "create_op_set_attribute_identity_user_backlink",
		keelSchema: `
			model CompanyUser {
				fields {
					identity Identity @unique @relation(user)
				}
			}
			model Record {
				fields {
					name Text
					user CompanyUser
				}
				actions {
					create createRecord() with (name) {
						@set(record.user = ctx.identity.user)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createRecord",
		input:      map[string]any{"name": "Dave"},
		expectedTemplate: `
			WITH
				select_identity (column_0) AS (
					SELECT "identity$user"."id"
					FROM "identity"
					LEFT JOIN "company_user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id"
					WHERE "identity"."id" IS NOT DISTINCT FROM ?),
				new_1_record AS (
					INSERT INTO "record" (name, user_id)
					VALUES (
						?,
						(SELECT column_0 FROM select_identity))
					RETURNING *)
			SELECT *, set_identity_id(?) AS __keel_identity_id FROM new_1_record`,
		identity:     identity,
		expectedArgs: []any{identity[parser.FieldNameId].(string), "Dave", identity[parser.FieldNameId].(string)},
	},
	{
		name: "create_op_set_attribute_identity_user_backlink_field",
		keelSchema: `
			model CompanyUser {
				fields {
					identity Identity @unique @relation(user)
					isActive Boolean
				}
			}
			model Record {
				fields {
					name Text
					user CompanyUser
					isActive Boolean
				}
				actions {
					create createRecord() with (name) {
						@set(record.user = ctx.identity.user)
						@set(record.isActive = ctx.identity.user.isActive)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createRecord",
		input:      map[string]any{"name": "Dave"},
		expectedTemplate: `
			WITH
				select_identity (column_0, column_1) AS (
					SELECT "identity$user"."id", "identity$user"."is_active"
					FROM "identity"
					LEFT JOIN "company_user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id"
					WHERE "identity"."id" IS NOT DISTINCT FROM ?),
				new_1_record AS (
					INSERT INTO "record" (is_active, name, user_id)
					VALUES (
						(SELECT column_1 FROM select_identity),
						?,
						(SELECT column_0 FROM select_identity))
					RETURNING *)
			SELECT *, set_identity_id(?) AS __keel_identity_id FROM new_1_record`,
		identity:     identity,
		expectedArgs: []any{identity[parser.FieldNameId].(string), "Dave", identity[parser.FieldNameId].(string)},
	},

	{
		name: "update_op_set_attribute_context_identity_id",
		keelSchema: `
			model Person {
				fields {
					mainIdentity Identity
				}
				actions {
					update updatePerson(id) {
						@set(person.mainIdentity.id = ctx.identity.id)
					}
				}
				@permission(expression: true, actions: [update])
			}`,
		actionName: "updatePerson",
		input: map[string]any{
			"where": map[string]any{
				"id": "xyz",
			},
		},
		expectedTemplate: `
			UPDATE "person"
			SET main_identity_id = ?
			WHERE "person"."id" IS NOT DISTINCT FROM ?
			RETURNING "person".*, set_identity_id(?) AS __keel_identity_id`,
		identity:     identity,
		expectedArgs: []any{identity[parser.FieldNameId].(string), "xyz", identity[parser.FieldNameId].(string)},
	},
	{
		name: "update_op_set_attribute_context_identity",
		keelSchema: `
			model Person {
				fields {
					mainIdentity Identity
				}
				actions {
					update updatePerson(id) {
						@set(person.mainIdentity = ctx.identity)
					}
				}
				@permission(expression: true, actions: [update])
			}`,
		actionName: "updatePerson",
		input: map[string]any{
			"where": map[string]any{
				"id": "xyz",
			},
		},
		expectedTemplate: `
			UPDATE "person"
			SET main_identity_id = ?
			WHERE "person"."id" IS NOT DISTINCT FROM ?
			RETURNING "person".*, set_identity_id(?) AS __keel_identity_id`,
		identity:     identity,
		expectedArgs: []any{identity[parser.FieldNameId].(string), "xyz", identity[parser.FieldNameId].(string)},
	},
	{
		name: "update_op_set_attribute_input",
		keelSchema: `
			model Person {
				fields {
					name Text
					nickName Text
				}
				actions {
					update updatePerson(id) with (name) {
						@set(person.nickName = name)
					}
				}
				@permission(expression: true, actions: [update])
			}`,
		actionName: "updatePerson",
		input: map[string]any{
			"where": map[string]any{
				"id": "xyz",
			},
			"values": map[string]any{
				"name": "Dave",
			},
		},
		expectedTemplate: `
			UPDATE "person"
			SET
				name = ?,
				nick_name = ?
			WHERE "person"."id" IS NOT DISTINCT FROM ?
			RETURNING "person".*`,
		expectedArgs: []any{"Dave", "Dave", "xyz"},
	},
	{
		name: "update_op_set_attribute_identity_user_backlink_field",
		keelSchema: `
			model CompanyUser {
				fields {
					identity Identity @unique @relation(user)
					isActive Boolean
				}
			}
			model Record {
				fields {
					name Text
					user CompanyUser
					isActive Boolean
				}
				actions {
					update updateRecordOwner(id) {
						@set(record.user = ctx.identity.user)
						@set(record.isActive = ctx.identity.user.isActive)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "updateRecordOwner",
		input: map[string]any{
			"where": map[string]any{
				"id": "xyz",
			},
		},
		expectedTemplate: `
			WITH
				select_identity (column_0, column_1) AS (
					SELECT "identity$user"."id", "identity$user"."is_active"
					FROM "identity"
					LEFT JOIN "company_user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id"
					WHERE "identity"."id" IS NOT DISTINCT FROM ?)
			UPDATE "record"
			SET
				is_active = (SELECT column_1 FROM select_identity),
				user_id = (SELECT column_0 FROM select_identity)
			WHERE
				"record"."id" IS NOT DISTINCT FROM ?
			RETURNING "record".*, set_identity_id(?) AS __keel_identity_id`,
		identity:     identity,
		expectedArgs: []any{identity[parser.FieldNameId].(string), "xyz", identity[parser.FieldNameId].(string)},
	},
	{
		name: "create_op_optional_inputs",
		keelSchema: `
			model Person {
				fields {
					name Text?
					age Number?
					isActive Boolean?
				}
				actions {
					create createPerson() with (name?, age?, isActive?)
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createPerson",
		input:      map[string]any{},
		expectedTemplate: `
			WITH
				new_1_person AS
					(INSERT INTO "person"
					DEFAULT VALUES
					RETURNING *)
			SELECT * FROM new_1_person`,
		expectedArgs: []any{},
	},
	{
		name: "create_op_optional_inputs_on_M_to_1_relationship",
		keelSchema: `
			model Person {
				fields {
					name Text
					company Company?
				}
				actions {
					create createPerson() with (name, company.id?)
				}
				@permission(expression: true, actions: [create])
			}
			model Company {

			}`,
		actionName: "createPerson",
		input: map[string]any{
			"name": "Bob",
		},
		expectedTemplate: `
			WITH
				new_1_person AS
					(INSERT INTO "person"
						(name)
					VALUES
						(?)
					RETURNING *)
			SELECT * FROM new_1_person`,
		expectedArgs: []any{"Bob"},
	},
	{
		name: "update_op_set_attribute",
		keelSchema: `
			model Person {
				fields {
					name Text
					age Number
					isActive Boolean
					bio Markdown
				}
				actions {
					update updatePerson(id) {
						@set(person.name = "Bob")
						@set(person.age = 100)
						@set(person.isActive = true)
						@set(person.bio = "## Biography")
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "updatePerson",
		input: map[string]any{
			"where": map[string]any{
				"id": "xyz",
			},
		},
		expectedTemplate: `
			UPDATE
				"person"
			SET
			    age = ?, bio = ?, is_active = ?, name = ?
			WHERE
				"person"."id" IS NOT DISTINCT FROM ?
			RETURNING
				"person".*`,
		expectedArgs: []any{int64(100), "## Biography", true, "Bob", "xyz"},
	},
	{
		name: "update_op_optional_inputs",
		keelSchema: `
			model Person {
				fields {
					name Text?
					age Number?
					isActive Boolean?
				}
				actions {
					update updatePerson(id) with (name?, age?, isActive?)
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "updatePerson",
		input: map[string]any{
			"where": map[string]any{
				"id": "xyz",
			},
			"values": map[string]any{
				"name": "Bob",
			},
		},
		expectedTemplate: `
			UPDATE
				"person"
			SET
			    name = ?
			WHERE
				"person"."id" IS NOT DISTINCT FROM ?
			RETURNING
				"person".*`,
		expectedArgs: []any{"Bob", "xyz"},
	},
	{
		name: "list_op_no_filter",
		keelSchema: `
			model Thing {
				actions {
					list listThings()
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" ) AS totalCount
			FROM
				"thing"
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{50},
	},
	{
		name: "list_op_implicit_input_text_contains",
		keelSchema: `
			model Thing {
				fields {
					name Text
				}
				actions {
					list listThings(name)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"contains": "bob"}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."name" LIKE ?) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."name" LIKE ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"%%bob%%", "%%bob%%", 50},
	},
	{
		name: "list_op_implicit_input_text_startsWith",
		keelSchema: `
			model Thing {
				fields {
					name Text
				}
				actions {
					list listThings(name)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"startsWith": "bob"}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."name" LIKE ?) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."name" LIKE ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"bob%%", "bob%%", 50},
	},
	{
		name: "list_op_implicit_input_text_endsWith",
		keelSchema: `
			model Thing {
				fields {
					name Text
				}
				actions {
					list listThings(name)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"endsWith": "bob"}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."name" LIKE ?) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."name" LIKE ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"%%bob", "%%bob", 50},
	},
	{
		name: "list_op_implicit_input_text_oneof",
		keelSchema: `
            model Thing {
                fields {
                    name Text
                }
                actions {
                    list listThings(name)
                }
                @permission(expression: true, actions: [list])
            }`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"name": map[string]any{
					"oneOf": []any{"bob", "dave", "adam", "pete"}}}},
		expectedTemplate: `
            SELECT
                DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
								(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."name" = ANY(ARRAY[?, ?, ?, ?]::TEXT[])) AS totalCount
            FROM 
                "thing" 
            WHERE
                "thing"."name" = ANY(ARRAY[?, ?, ?, ?]::TEXT[])
            ORDER BY 
                "thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"bob", "dave", "adam", "pete", "bob", "dave", "adam", "pete", 50},
	},
	{
		name: "list_op_implicit_input_enum_oneof",
		keelSchema: `
            model Thing {
                fields {
                    category Category
                }
                actions {
                    list listThings(category)
                }
                @permission(expression: true, actions: [list])
            }
			enum Category {
				Technical
				Food
				Lifestyle
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"category": map[string]any{
					"oneOf": []any{"Technical", "Food"}}}},
		expectedTemplate: `
            SELECT
                DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
								(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."category" = ANY(ARRAY[?, ?]::TEXT[])) AS totalCount
            FROM 
                "thing" 
            WHERE
				"thing"."category" = ANY(ARRAY[?, ?]::TEXT[])
            ORDER BY 
                "thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"Technical", "Food", "Technical", "Food", 50},
	},
	{
		name: "list_op_implicit_input_timestamp_after",
		keelSchema: `
			model Thing {
				actions {
					list listThings(createdAt)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"after": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."created_at" > ?) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."created_at" > ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), 50},
	},
	{
		name: "list_op_implicit_input_timestamp_onorafter",
		keelSchema: `
			model Thing {
				actions {
					list listThings(createdAt)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"onOrAfter": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."created_at" >= ?) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."created_at" >= ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), 50},
	},
	{
		name: "list_op_implicit_input_timestamp_after",
		keelSchema: `
			model Thing {
				actions {
					list listThings(createdAt)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"before": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."created_at" < ?) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."created_at" < ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), 50},
	},
	{
		name: "list_op_implicit_input_timestamp_onorbefore",
		keelSchema: `
			model Thing {
				actions {
					list listThings(createdAt)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"createdAt": map[string]any{
					"onOrBefore": time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC)}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."created_at" <= ?) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."created_at" <= ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), time.Date(2020, 11, 19, 9, 0, 30, 0, time.UTC), 50},
	},
	{
		name: "list_op_expression_text_in",
		keelSchema: `
			model Thing {
				fields {
                    title Text
                }
				actions {
					list listThings() {
						@where(thing.title in ["title1", "title2"])
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."title" = ANY(ARRAY[?, ?]::TEXT[])) AS totalCount
			FROM 
				"thing" 
			WHERE 
				"thing"."title" = ANY(ARRAY[?, ?]::TEXT[])
			ORDER BY 
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"title1", "title2", "title1", "title2", 50},
	},
	{
		name: "list_op_expression_text_in_field",
		keelSchema: `
			model RepeatedThing {
				fields {
					name Text
					thing Thing
				}
			}
			model Thing {
				fields {
                    title Text
					repeatedThings RepeatedThing[]
                }
				actions {
					list listRepeatedThings() {
						@where(thing.title in thing.repeatedThings.name)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listRepeatedThings",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, 
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "thing"."id") 
					FROM 
						"thing" 
					LEFT JOIN "repeated_thing" AS "thing$repeated_things" ON 
						"thing$repeated_things"."thing_id" = "thing"."id" 
					WHERE 
						"thing"."title" IS NOT DISTINCT FROM "thing$repeated_things"."name") AS totalCount FROM "thing" 
			LEFT JOIN "repeated_thing" AS "thing$repeated_things" ON 
				"thing$repeated_things"."thing_id" = "thing"."id" 
			WHERE 
				"thing"."title" IS NOT DISTINCT FROM "thing$repeated_things"."name" 
			ORDER BY 
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{50},
	},
	{
		name: "list_op_expression_text_not_in_field",
		keelSchema: `
			model RepeatedThing {
				fields {
					name Text
					thing Thing
				}
			}
			model Thing {
				fields {
                    title Text
					repeatedThings RepeatedThing[]
                }
				actions {
					list listRepeatedThings() {
						@where(thing.title not in thing.repeatedThings.name)
					} 
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listRepeatedThings",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("thing"."id") "thing".*, 
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "thing"."id") 
					FROM 
						"thing" 
					LEFT JOIN "repeated_thing" AS "thing$repeated_things" ON 
						"thing$repeated_things"."thing_id" = "thing"."id" 
					WHERE 
						"thing"."title" IS DISTINCT FROM "thing$repeated_things"."name") AS totalCount FROM "thing" 
			LEFT JOIN "repeated_thing" AS "thing$repeated_things" ON 
				"thing$repeated_things"."thing_id" = "thing"."id" 
			WHERE 
				"thing"."title" IS DISTINCT FROM "thing$repeated_things"."name" 
			ORDER BY 
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{50},
	},
	{
		name: "list_op_expression_text_notin",
		keelSchema: `
			model Thing {
				fields {
                    title Text
                }
				actions {
					list listThings() {
						@where(thing.title not in ["title1", "title2"])
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE NOT "thing"."title" = ANY(ARRAY[?, ?]::TEXT[])) AS totalCount
			FROM 
				"thing" 
			WHERE 
				NOT "thing"."title" = ANY(ARRAY[?, ?]::TEXT[])
			ORDER BY 
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"title1", "title2", "title1", "title2", 50},
	},
	{
		name: "list_op_expression_number_in",
		keelSchema: `
			model Thing {
				fields {
                    age Number
                }
				actions {
					list listThings() {
						@where(thing.age in [10, 20])
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."age" = ANY(ARRAY[?, ?]::INTEGER[])) AS totalCount
			FROM 
				"thing" 
			WHERE 
				"thing"."age" = ANY(ARRAY[?, ?]::INTEGER[])
			ORDER BY 
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{int64(10), int64(20), int64(10), int64(20), 50},
	},
	{
		name: "list_op_expression_number_notin",
		keelSchema: `
			model Thing {
				fields {
                    age Number
                }
				actions {
					list listThings() {
						@where(thing.age not in [10, 20])
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE NOT "thing"."age" = ANY(ARRAY[?, ?]::INTEGER[])) AS totalCount
			FROM 
				"thing" 
			WHERE 
				NOT "thing"."age" = ANY(ARRAY[?, ?]::INTEGER[])
			ORDER BY 
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{int64(10), int64(20), int64(10), int64(20), 50},
	},
	{
		name: "list_op_expression_model_in_backlink",
		keelSchema: `
			model Account {
				fields {
					username Text @unique
					identity Identity @unique
					followers Follow[]
					following Follow[]
				}

				actions {
					list accountsFollowed() {
						@where(account in ctx.identity.account.following.followee)
            			@orderBy(username: asc)
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}
			model Follow {
				fields {
					followee Account @relation(followers)
					follower Account @relation(following)
				}
				@unique([follower, followee])
			}`,
		actionName: "accountsFollowed",
		input:      map[string]any{},
		identity:   identity,
		expectedTemplate: `
			SELECT
				DISTINCT ON("account"."username", "account"."id") "account".*,
				CASE WHEN LEAD("account"."id") OVER (ORDER BY "account"."username" ASC, "account"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("account"."username", "account"."id")) FROM "account" WHERE "account"."id" IN (SELECT "identity$account$following"."followee_id" FROM "identity" LEFT JOIN "account" AS "identity$account" ON "identity$account"."identity_id" = "identity"."id" LEFT JOIN "follow" AS "identity$account$following" ON "identity$account$following"."follower_id" = "identity$account"."id" WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$account$following"."followee_id" IS DISTINCT FROM NULL )) AS totalCount
			FROM
				"account"
			WHERE "account"."id" IN
				(SELECT "identity$account$following"."followee_id"
				FROM "identity"
				LEFT JOIN "account" AS "identity$account" ON "identity$account"."identity_id" = "identity"."id"
				LEFT JOIN "follow" AS "identity$account$following" ON "identity$account$following"."follower_id" = "identity$account"."id"
				WHERE
					"identity"."id" IS NOT DISTINCT FROM ? AND
					"identity$account$following"."followee_id" IS DISTINCT FROM NULL )
			ORDER BY "account"."username" ASC, "account"."id" ASC LIMIT ?`,
		expectedArgs: []any{"identityId", "identityId", 50},
	},
	{
		name: "list_op_expression_model_id_in_backlink",
		keelSchema: `
			model Account {
				fields {
					username Text @unique
					identity Identity @unique
					followers Follow[]
					following Follow[]
				}

				actions {
					list accountsFollowed() {
						@where(account.id in ctx.identity.account.following.account.id)
            			@orderBy(username: asc)
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}
			model Follow {
				fields {
					account Account @relation(followers)
					follower Account @relation(following)
				}
				@unique([follower, account])
			}`,
		actionName: "accountsFollowed",
		input:      map[string]any{},
		identity:   identity,
		expectedTemplate: `
			SELECT
				DISTINCT ON("account"."username", "account"."id") "account".*,
				CASE WHEN LEAD("account"."id") OVER (ORDER BY "account"."username" ASC, "account"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("account"."username", "account"."id")) FROM "account" WHERE "account"."id" IN (SELECT "identity$account$following$account"."id" FROM "identity" LEFT JOIN "account" AS "identity$account" ON "identity$account"."identity_id" = "identity"."id" LEFT JOIN "follow" AS "identity$account$following" ON "identity$account$following"."follower_id" = "identity$account"."id" LEFT JOIN "account" AS "identity$account$following$account" ON "identity$account$following$account"."id" = "identity$account$following"."account_id" WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$account$following$account"."id" IS DISTINCT FROM NULL )) AS totalCount
			FROM
				"account"
			WHERE
				"account"."id" IN
					(SELECT "identity$account$following$account"."id"
					FROM "identity"
					LEFT JOIN "account" AS "identity$account" ON "identity$account"."identity_id" = "identity"."id"
					LEFT JOIN "follow" AS "identity$account$following" ON "identity$account$following"."follower_id" = "identity$account"."id"
					LEFT JOIN "account" AS "identity$account$following$account" ON "identity$account$following$account"."id" = "identity$account$following"."account_id"
					WHERE
						"identity"."id" IS NOT DISTINCT FROM ? AND
						"identity$account$following$account"."id" IS DISTINCT FROM NULL )
			ORDER BY
				"account"."username" ASC,
				"account"."id" ASC LIMIT ?`,
		expectedArgs: []any{"identityId", "identityId", 50},
	},
	{
		name: "list_op_expression_model_id_not_in_backlink",
		keelSchema: `
			model Account {
				fields {
					username Text @unique
					identity Identity @unique @relation(primaryAccount)
					followers Follow[]
					following Follow[]
				}

				actions {
					list accountsNotFollowed() {
						@where(account.identity != ctx.identity)
           				@where(account.id not in ctx.identity.primaryAccount.following.followee.id)
            			@orderBy(username: asc)
						@permission(expression: ctx.isAuthenticated)
					}
				}
			}
			model Follow {
				fields {
					followee Account @relation(followers)
					follower Account @relation(following)
				}
				@unique([follower, followee])
			}`,
		actionName: "accountsNotFollowed",
		input:      map[string]any{},
		identity:   identity,
		expectedTemplate: `
			SELECT
				DISTINCT ON("account"."username", "account"."id") "account".*,
				CASE WHEN LEAD("account"."id") OVER (ORDER BY "account"."username" ASC, "account"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("account"."username", "account"."id")) FROM "account" WHERE "account"."identity_id" IS DISTINCT FROM ? AND "account"."id" NOT IN (SELECT "identity$primary_account$following$followee"."id" FROM "identity" LEFT JOIN "account" AS "identity$primary_account" ON "identity$primary_account"."identity_id" = "identity"."id" LEFT JOIN "follow" AS "identity$primary_account$following" ON "identity$primary_account$following"."follower_id" = "identity$primary_account"."id" LEFT JOIN "account" AS "identity$primary_account$following$followee" ON "identity$primary_account$following$followee"."id" = "identity$primary_account$following"."followee_id" WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$primary_account$following$followee"."id" IS DISTINCT FROM NULL )) AS totalCount
			FROM
				"account"
			WHERE
				"account"."identity_id" IS DISTINCT FROM ? AND
				"account"."id" NOT IN
					(SELECT "identity$primary_account$following$followee"."id"
					FROM "identity"
					LEFT JOIN "account" AS "identity$primary_account" ON "identity$primary_account"."identity_id" = "identity"."id"
					LEFT JOIN "follow" AS "identity$primary_account$following" ON "identity$primary_account$following"."follower_id" = "identity$primary_account"."id"
					LEFT JOIN "account" AS "identity$primary_account$following$followee" ON "identity$primary_account$following$followee"."id" = "identity$primary_account$following"."followee_id"
					WHERE
						"identity"."id" IS NOT DISTINCT FROM ? AND
						"identity$primary_account$following$followee"."id" IS DISTINCT FROM NULL )
			ORDER BY
				"account"."username" ASC,
				"account"."id" ASC LIMIT ?`,
		expectedArgs: []any{"identityId", "identityId", "identityId", "identityId", 50},
	},
	{
		name: "get_op_context_user_backlink_model_with_relation_attribute",
		keelSchema: `
			model CompanyUser {
				fields {
					identity Identity @unique @relation(primaryUser)
				}
			}
			model Record {
				fields {
					name Text
					owner CompanyUser
				}
				actions {
					get record(id) {
						@where(record.owner == ctx.identity.primaryUser)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "record",
		input:      map[string]any{"id": "xyz"},
		expectedTemplate: `
			SELECT
				DISTINCT ON("record"."id") "record".* FROM "record"
			WHERE
				"record"."id" IS NOT DISTINCT FROM ? AND
				"record"."owner_id" IS NOT DISTINCT FROM (
					SELECT "identity$primary_user"."id"
					FROM "identity"
					LEFT JOIN "company_user" AS "identity$primary_user" ON "identity$primary_user"."identity_id" = "identity"."id"
					WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$primary_user"."id" IS DISTINCT FROM NULL)`,
		identity:     identity,
		expectedArgs: []any{"xyz", "identityId"},
	},
	{
		name: "get_op_context_user_backlink_field",
		keelSchema: `
			model CompanyUser {
				fields {
					identity Identity @unique @relation(primaryUser)
					isActive Boolean
				}
			}
			model Record {
				fields {
					name Text
				}
				actions {
					get record(id) {
						@where(ctx.identity.primaryUser.isActive)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "record",
		input:      map[string]any{"id": "xyz"},
		expectedTemplate: `
			SELECT
				DISTINCT ON("record"."id") "record".* FROM "record"
			WHERE
				"record"."id" IS NOT DISTINCT FROM ? AND
					(SELECT "identity$primary_user"."is_active"
					FROM "identity"
					LEFT JOIN "company_user" AS "identity$primary_user" ON "identity$primary_user"."identity_id" = "identity"."id"
					WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$primary_user"."is_active" IS DISTINCT FROM NULL) IS NOT DISTINCT FROM ?`,
		identity:     identity,
		expectedArgs: []any{"xyz", "identityId", true},
	},
	{
		name: "list_op_implicit_input_nested_model_id",
		keelSchema: `
			model BankAccount {
				fields {
					entity Entity @unique
					balance Number @default(0)
				}
			}
			model Entity {
				fields {
					name Text
					users EntityUser[]
					account BankAccount
				}
			}
			model EntityUser {
				fields {
					name Text
					identity Identity @unique @relation(user)
					entity Entity
				}
				actions {
					list bankAccountUsers(entity.account.id) {
						 @orderBy(name: asc)
					}
				}
			}`,
		actionName: "bankAccountUsers",
		input:      map[string]any{"where": map[string]any{"entity": map[string]any{"account": map[string]any{"id": map[string]any{"equals": "123"}}}}},
		identity:   identity,
		expectedTemplate: `
			SELECT
				DISTINCT ON("entity_user"."name", "entity_user"."id") "entity_user".*,
				CASE WHEN LEAD("entity_user"."id") OVER (ORDER BY "entity_user"."name" ASC, "entity_user"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("entity_user"."name", "entity_user"."id")) FROM "entity_user" LEFT JOIN "entity" AS "entity_user$entity" ON "entity_user$entity"."id" = "entity_user"."entity_id" LEFT JOIN "bank_account" AS "entity_user$entity$account" ON "entity_user$entity$account"."entity_id" = "entity_user$entity"."id" WHERE "entity_user$entity$account"."id" IS NOT DISTINCT FROM ?) AS totalCount
			FROM
				"entity_user"
			LEFT JOIN
				"entity" AS "entity_user$entity" ON "entity_user$entity"."id" = "entity_user"."entity_id"
			LEFT JOIN
				"bank_account" AS "entity_user$entity$account" ON "entity_user$entity$account"."entity_id" = "entity_user$entity"."id"
			WHERE
				"entity_user$entity$account"."id" IS NOT DISTINCT FROM ?
			ORDER BY
				"entity_user"."name" ASC,
				"entity_user"."id" ASC LIMIT ?`,
		expectedArgs: []any{"123", "123", 50},
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
				actions {
					list listThings(parent.name)
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{
				"parent": map[string]any{
					"name": map[string]any{
						"equals": "bob"}}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" LEFT JOIN "parent" AS "thing$parent" ON "thing$parent"."id" = "thing"."parent_id" WHERE "thing$parent"."name" IS NOT DISTINCT FROM ?) AS totalCount
			FROM
				"thing"
			LEFT JOIN
				"parent" AS "thing$parent"
					ON "thing$parent"."id" = "thing"."parent_id"
			WHERE
				"thing$parent"."name" IS NOT DISTINCT FROM ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"bob", "bob", 50},
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
				actions {
					list listThings() {
						@where(thing.parent.isActive == false)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" LEFT JOIN "parent" AS "thing$parent" ON "thing$parent"."id" = "thing"."parent_id" WHERE "thing$parent"."is_active" IS NOT DISTINCT FROM ?) AS totalCount
			FROM
				"thing"
			LEFT JOIN
				"parent" AS "thing$parent"
					ON "thing$parent"."id" = "thing"."parent_id"
			WHERE
				"thing$parent"."is_active" IS NOT DISTINCT FROM ?
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{false, false, 50},
	},
	{
		name: "list_op_orderby",
		keelSchema: `
			model Thing {
				fields {
					name Text
					views Number
				}
				actions {
					list listThings() {
						@orderBy(name: asc, views: desc)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."name", "thing"."views", "thing"."id") "thing".*,
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."name" ASC, "thing"."views" DESC, "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("thing"."name", "thing"."views", "thing"."id")) FROM "thing" ) AS totalCount
			FROM
				"thing"
			ORDER BY
				"thing"."name" ASC,
				"thing"."views" DESC,
				"thing"."id" ASC
			LIMIT ?`,
		expectedArgs: []any{50},
	},
	{
		name: "list_op_orderby_with_after",
		keelSchema: `
			model Thing {
				fields {
					name Text
					views Number
				}
				actions {
					list listThings() {
						@orderBy(name: asc, views: desc)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"after": "xyz",
			"where": map[string]any{}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."name", "thing"."views", "thing"."id") "thing".*,
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."name" ASC, "thing"."views" DESC, "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("thing"."name", "thing"."views", "thing"."id")) FROM "thing" ) AS totalCount
			FROM
				"thing"
			WHERE
				(
					"thing"."name" > (SELECT "thing"."name" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? )
					OR
					( "thing"."name" IS NOT DISTINCT FROM (SELECT "thing"."name" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) AND "thing"."views" < (SELECT "thing"."views" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) )
					OR
					( "thing"."name" IS NOT DISTINCT FROM (SELECT "thing"."name" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) AND "thing"."views" IS NOT DISTINCT FROM (SELECT "thing"."views" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) AND "thing"."id" > (SELECT "thing"."id" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) )
				)
			ORDER BY
				"thing"."name" ASC,
				"thing"."views" DESC,
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"xyz", "xyz", "xyz", "xyz", "xyz", "xyz", 50},
	},
	{
		name: "list_op_sortable",
		keelSchema: `
			model Thing {
				fields {
					name Text
					views Number
				}
				actions {
					list listThings() {
						@sortable(name, views)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{},
			"orderBy": []any{
				map[string]any{"name": "asc"},
				map[string]any{"views": "desc"}},
		},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."name", "thing"."views", "thing"."id") "thing".*,
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."name" ASC, "thing"."views" DESC, "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("thing"."name", "thing"."views", "thing"."id")) FROM "thing" ) AS totalCount
			FROM
				"thing"
			ORDER BY
				"thing"."name" ASC,
				"thing"."views" DESC,
				"thing"."id" ASC
			LIMIT ?`,
		expectedArgs: []any{50},
	},
	{
		name: "list_op_sortable_with_after",
		keelSchema: `
			model Thing {
				fields {
					name Text
					views Number
				}
				actions {
					list listThings() {
						@sortable(name, views)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"after": "xyz",
			"where": map[string]any{},
			"orderBy": []any{
				map[string]any{"name": "asc"},
				map[string]any{"views": "desc"}},
		},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."name", "thing"."views", "thing"."id") "thing".*,
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."name" ASC, "thing"."views" DESC, "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("thing"."name", "thing"."views", "thing"."id")) FROM "thing" ) AS totalCount
			FROM
				"thing"
			WHERE
				(
					"thing"."name" > (SELECT "thing"."name" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? )
					OR
					( "thing"."name" IS NOT DISTINCT FROM (SELECT "thing"."name" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) AND "thing"."views" < (SELECT "thing"."views" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) )
					OR
					( "thing"."name" IS NOT DISTINCT FROM (SELECT "thing"."name" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) AND "thing"."views" IS NOT DISTINCT FROM (SELECT "thing"."views" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) AND "thing"."id" > (SELECT "thing"."id" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? ) )
				)
			ORDER BY
				"thing"."name" ASC,
				"thing"."views" DESC,
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"xyz", "xyz", "xyz", "xyz", "xyz", "xyz", 50},
	},
	{
		name: "list_op_sortable_overriding_orderby",
		keelSchema: `
			model Thing {
				fields {
					name Text
					views Number
				}
				actions {
					list listThings() {
						@orderBy(name: desc)
						@sortable(name, views)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{},
			"orderBy": []any{
				map[string]any{"name": "asc"},
				map[string]any{"views": "desc"}},
		},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."name", "thing"."views", "thing"."id") "thing".*,
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."name" ASC, "thing"."views" DESC, "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("thing"."name", "thing"."views", "thing"."id")) FROM "thing" ) AS totalCount
			FROM
				"thing"
			ORDER BY
				"thing"."name" ASC,
				"thing"."views" DESC,
				"thing"."id" ASC
			LIMIT ?`,
		expectedArgs: []any{50},
	},
	{
		name: "list_op_sortable_and_orderby",
		keelSchema: `
			model Thing {
				fields {
					name Text
					views Number
				}
				actions {
					list listThings() {
						@sortable(name, views)
						@orderBy(name: asc)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"where": map[string]any{},
			"orderBy": []any{
				map[string]any{"views": "desc"}},
		},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."name", "thing"."views", "thing"."id") "thing".*,
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."name" ASC, "thing"."views" DESC, "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT ("thing"."name", "thing"."views", "thing"."id")) FROM "thing" ) AS totalCount
			FROM
				"thing"
			ORDER BY
				"thing"."name" ASC,
				"thing"."views" DESC,
				"thing"."id" ASC
			LIMIT ?`,
		expectedArgs: []any{50},
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
					parent Parent
				}
				actions {
					create createThing() with (name, age, parent.id)
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createThing",
		input: map[string]any{
			"name":   "bob",
			"age":    21,
			"parent": map[string]any{"id": "123"},
		},
		expectedTemplate: `
			WITH
				new_1_thing AS
					(INSERT INTO "thing"
						(age, name, parent_id)
					VALUES
						(?, ?, ?)
					RETURNING *)
			SELECT * FROM new_1_thing`,
		expectedArgs: []any{21, "bob", "123"},
	},
	{
		name: "create_op_many_reln_optional_input_not_provided",
		keelSchema: `
		model Customer {
			fields {
				name Text
				orders Order[]
			}

			actions {
				create createCustomer() with (name, orders.deliveryAddress?)
			}

			@permission(
				actions: [get, list, update, delete, create],
				expression: true
			)
		}
		model Order {
			fields {
				deliveryAddress Text
				customer Customer?
			}
		}
		`,
		actionName: "createCustomer",
		input: map[string]any{
			"name": "fred",
		},
		expectedTemplate: `
			WITH new_1_customer AS (
				INSERT INTO "customer" (name) VALUES (?) RETURNING *)
			SELECT * FROM new_1_customer`,
		expectedArgs: []any{"fred"},
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
				actions {
					update updateThing(id) with (name, age, parent.id)
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "updateThing",
		input: map[string]any{
			"where": map[string]any{
				"id": "789",
			},
			"values": map[string]any{
				"name": "bob",
				"age":  21,
				"parent": map[string]any{
					"id": "123",
				},
			},
		},
		expectedTemplate: `
			UPDATE
				"thing"
			SET
				age = ?,
				name = ?,
				parent_id = ?
			WHERE
				"thing"."id" IS NOT DISTINCT FROM ?
			RETURNING
				"thing".*`,
		expectedArgs: []any{21, "bob", "123", "789"},
	},
	{
		name: "delete_op_by_id",
		keelSchema: `
			model Thing {
				actions {
					delete deleteThing(id)
				}
				@permission(expression: true, actions: [delete])
			}`,
		actionName: "deleteThing",
		input:      map[string]any{"id": "123"},
		expectedTemplate: `
			DELETE FROM
				"thing"
			WHERE
				"thing"."id" IS NOT DISTINCT FROM ?
			RETURNING "thing"."id"`,
		expectedArgs: []any{"123"},
	},
	{
		name: "delete_op_relationship_condition",
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
				actions {
					delete deleteThing(id) {
						@where(thing.parent.name == "XYZ")
					}
				}
				@permission(expression: true, actions: [delete])
			}`,
		actionName: "deleteThing",
		input:      map[string]any{"id": "123"},
		expectedTemplate: `
			DELETE FROM
				"thing"
			USING
				"parent" AS "thing$parent"
			WHERE
				"thing"."id" IS NOT DISTINCT FROM ? AND
				"thing$parent"."name" IS NOT DISTINCT FROM ?
			RETURNING "thing"."id"`,
		expectedArgs: []any{"123", "XYZ"},
	},
	{
		name: "list_op_forward_paging",
		keelSchema: `
			model Thing {
				actions {
					list listThings()
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"first": 2,
			"after": "123",
		},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" ) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."id" > (SELECT "thing"."id" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? )
			ORDER BY
				"thing"."id" ASC
			LIMIT ?`,
		expectedArgs: []any{"123", 2},
	},
	{
		name: "list_op_backwards_paging",
		keelSchema: `
			model Thing {
				actions {
					list listThings()
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThings",
		input: map[string]any{
			"last":   2,
			"before": "123",
		},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*,
				CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" ) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."id" < (SELECT "thing"."id" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? )
			ORDER BY
				"thing"."id" DESC LIMIT ?`,
		expectedArgs: []any{"123", 2},
	},
	{
		name: "list_multiple_conditions_no_parenthesis",
		keelSchema: `
			model Thing {
				fields {
					first Text
					second Number
					third Boolean
				}
				actions {
					list listThing() {
						@where(thing.first == "first" and thing.second == 10 or thing.third == true and thing.second > 100)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThing",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE ( "thing"."first" IS NOT DISTINCT FROM ? AND "thing"."second" IS NOT DISTINCT FROM ? OR "thing"."third" IS NOT DISTINCT FROM ? AND "thing"."second" > ? )) AS totalCount
			FROM
				"thing"
			WHERE
				( "thing"."first" IS NOT DISTINCT FROM ? AND
				"thing"."second" IS NOT DISTINCT FROM ? OR
				"thing"."third" IS NOT DISTINCT FROM ? AND
				"thing"."second" > ? )
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"first", int64(10), true, int64(100), "first", int64(10), true, int64(100), 50},
	},
	{
		name: "list_multiple_conditions_parenthesis_on_ands",
		keelSchema: `
			model Thing {
				fields {
					first Text
					second Number
					third Boolean
				}
				actions {
					list listThing() {
						@where((thing.first == "first" and thing.second == 10) or (thing.third == true and thing.second > 100))
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThing",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE ( ( "thing"."first" IS NOT DISTINCT FROM ? AND "thing"."second" IS NOT DISTINCT FROM ? ) OR ( "thing"."third" IS NOT DISTINCT FROM ? AND "thing"."second" > ? ) )) AS totalCount
			FROM
				"thing"
			WHERE
				( ( "thing"."first" IS NOT DISTINCT FROM ? AND "thing"."second" IS NOT DISTINCT FROM ? )
					OR
				( "thing"."third" IS NOT DISTINCT FROM ? AND "thing"."second" > ? ) )
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"first", int64(10), true, int64(100), "first", int64(10), true, int64(100), 50},
	},
	{
		name: "list_multiple_conditions_parenthesis_on_ors",
		keelSchema: `
			model Thing {
				fields {
					first Text
					second Number
					third Boolean
				}
				actions {
					list listThing() {
						@where((thing.first == "first" or thing.second == 10) and (thing.third == true or thing.second > 100))
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThing",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE ( ( "thing"."first" IS NOT DISTINCT FROM ? OR "thing"."second" IS NOT DISTINCT FROM ? ) AND ( "thing"."third" IS NOT DISTINCT FROM ? OR "thing"."second" > ? ) )) AS totalCount
			FROM
				"thing"
			WHERE
				( ( "thing"."first" IS NOT DISTINCT FROM ? OR "thing"."second" IS NOT DISTINCT FROM ? )
					AND
				( "thing"."third" IS NOT DISTINCT FROM ? OR "thing"."second" > ? ) )
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"first", int64(10), true, int64(100), "first", int64(10), true, int64(100), 50},
	},
	{
		name: "list_multiple_conditions_nested_parenthesis",
		keelSchema: `
			model Thing {
				fields {
					first Text
					second Number
					third Boolean
				}
				actions {
					list listThing() {
						@where(thing.first == "first" or (thing.second == 10 and (thing.third == true or thing.second > 100)))
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThing",
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE ( "thing"."first" IS NOT DISTINCT FROM ? OR ( "thing"."second" IS NOT DISTINCT FROM ? AND ( "thing"."third" IS NOT DISTINCT FROM ? OR "thing"."second" > ? ) ) )) AS totalCount
			FROM
				"thing"
			WHERE
				( "thing"."first" IS NOT DISTINCT FROM ? OR
					( "thing"."second" IS NOT DISTINCT FROM ? AND
						( "thing"."third" IS NOT DISTINCT FROM ? OR "thing"."second" > ? ) ) )
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"first", int64(10), true, int64(100), "first", int64(10), true, int64(100), 50},
	},
	{
		name: "list_multiple_conditions_implicit_and_explicit",
		keelSchema: `
			model Thing {
				fields {
					first Text
					second Number
					third Boolean
				}
				actions {
					list listThing(first, explicitSecond: Number) {
						@where(thing.second == explicitSecond or thing.third == false)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThing",
		input: map[string]any{
			"where": map[string]any{
				"first": map[string]any{
					"equals": "first"},
				"explicitSecond": int64(10)}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."first" IS NOT DISTINCT FROM ? AND ( "thing"."second" IS NOT DISTINCT FROM ? OR "thing"."third" IS NOT DISTINCT FROM ? )) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."first" IS NOT DISTINCT FROM ? AND
				( "thing"."second" IS NOT DISTINCT FROM ? OR "thing"."third" IS NOT DISTINCT FROM ? )
			ORDER BY
				"thing"."id" ASC LIMIT ?`,
		expectedArgs: []any{"first", int64(10), false, "first", int64(10), false, 50},
	},
	{
		name: "list_multiple_conditions_implicit_and_explicit_and_paging",
		keelSchema: `
			model Thing {
				fields {
					first Text
					second Number
					third Boolean
				}
				actions {
					list listThing(first, explicitSecond: Number) {
						@where(thing.second == explicitSecond or thing.third == false)
					}
				}
				@permission(expression: true, actions: [list])
			}`,
		actionName: "listThing",
		input: map[string]any{
			"first": 2,
			"after": "123",
			"where": map[string]any{
				"first": map[string]any{
					"equals": "first"},
				"explicitSecond": int64(10)}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("thing"."id") "thing".*, CASE WHEN LEAD("thing"."id") OVER (ORDER BY "thing"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext,
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."first" IS NOT DISTINCT FROM ? AND ( "thing"."second" IS NOT DISTINCT FROM ? OR "thing"."third" IS NOT DISTINCT FROM ? )) AS totalCount
			FROM
				"thing"
			WHERE
				"thing"."first" IS NOT DISTINCT FROM ? AND
				( "thing"."second" IS NOT DISTINCT FROM ? OR "thing"."third" IS NOT DISTINCT FROM ? ) AND
				"thing"."id" > (SELECT "thing"."id" FROM "thing" WHERE "thing"."id" IS NOT DISTINCT FROM ? )
			ORDER BY
				"thing"."id" ASC
			LIMIT ?`,
		expectedArgs: []any{"first", int64(10), false, "first", int64(10), false, "123", 2},
	},
	{
		name: "update_with_expression",
		keelSchema: `
			model Parent {
				fields {
					name Text
				}
			}
			model Thing {
				fields {
					name Text
					code Text @unique
				}
				actions {
					update updateThing(id) with (name) {
						@where(thing.code == "XYZ" or thing.code == "ABC")
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "updateThing",
		input: map[string]any{
			"where": map[string]any{
				"id": "789",
			},
			"values": map[string]any{
				"name": "bob",
			},
		},
		expectedTemplate: `
			UPDATE
				"thing"
			SET
				name = ?
			WHERE
				"thing"."id" IS NOT DISTINCT FROM ? AND
				( "thing"."code" IS NOT DISTINCT FROM ? OR "thing"."code" IS NOT DISTINCT FROM ? )
			RETURNING
				"thing".*`,
		expectedArgs: []any{"bob", "789", "XYZ", "ABC"},
	},
	{
		name: "delete_with_expression",
		keelSchema: `
			model Parent {
				fields {
					name Text
				}
			}
			model Thing {
				fields {
					name Text
					code Text @unique
				}
				actions {
					delete deleteThing(id) {
						@where(thing.code == "XYZ" or thing.code == "ABC")
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "deleteThing",
		input: map[string]any{
			"id": "789",
		},
		expectedTemplate: `
			DELETE FROM
				"thing"
			WHERE
				"thing"."id" IS NOT DISTINCT FROM ? AND
				( "thing"."code" IS NOT DISTINCT FROM ? OR "thing"."code" IS NOT DISTINCT FROM ? )
			RETURNING
				"thing"."id"`,
		expectedArgs: []any{"789", "XYZ", "ABC"},
	},
	{
		name: "create_relationships_1_to_M",
		keelSchema: `
			model Order {
				fields {
					onPromotion Boolean
					items OrderItem[]
				}
				actions {
					create createOrder() with (onPromotion, items.quantity, items.product.id)
				}
				@permission(expression: true, actions: [create])
			}
			model Product {
				fields {
					name Text
				}
			}
			model OrderItem {
				fields {
					order Order
					quantity Text
					product Product
				}
			}`,
		actionName: "createOrder",
		input: map[string]any{
			"onPromotion": true,
			"items": []any{
				map[string]any{
					"quantity": 2,
					"product": map[string]any{
						"id": "xyz",
					},
				},
				map[string]any{
					"quantity": 4,
					"product": map[string]any{
						"id": "abc",
					},
				},
			},
		},
		expectedTemplate: `
			WITH
				new_1_order AS
					(INSERT INTO "order"
						(on_promotion)
					VALUES
						(?)
					RETURNING *),
				new_1_order_item AS
					(INSERT INTO "order_item"
						(order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), ?, ?)
					RETURNING *),
				new_2_order_item AS
					(INSERT INTO "order_item"
						(order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), ?, ?)
					RETURNING *)
			SELECT * FROM new_1_order`,
		expectedArgs: []any{
			true,     // new_1_order
			"xyz", 2, // new_1_order_item
			"abc", 4, // new_2_order_item
		},
	},
	{
		name: "create_relationships_M_to_1_to_M",
		keelSchema: `
			model Order {
				fields {
					product Product
				}
				actions {
					create createOrder() with (product.name, product.attributes.name, product.attributes.status) {
						@set(order.product.createdOnOrder = true)
					}
				}
				@permission(expression: true, actions: [create])
			}
			model Product {
				fields {
					name Text
					isActive Boolean @default(true)
					createdOnOrder Boolean @default(false)
					attributes ProductAttribute[]
				}
			}
			model ProductAttribute {
				fields {
					product Product
					name Text
					status AttributeStatus
				}
			}
			enum AttributeStatus {
				NotApplicable
				Unknown
				Yes
				No
			}`,
		actionName: "createOrder",
		input: map[string]any{
			"product": map[string]any{
				"name": "Child Bicycle",
				"attributes": []any{
					map[string]any{
						"name":   "FDA approved",
						"status": "NotApplicable",
					},
					map[string]any{
						"name":   "Toy-safety-council approved",
						"status": "Yes",
					},
				},
			},
		},
		expectedTemplate: `
			WITH
				new_1_product AS
					(INSERT INTO "product"
						(created_on_order, name)
					VALUES
						(?, ?)
					RETURNING *),
				new_1_product_attribute AS
					(INSERT INTO "product_attribute"
						(name, product_id, status)
					VALUES
						(?, (SELECT id FROM new_1_product), ?)
					RETURNING *),
				new_2_product_attribute AS
					(INSERT INTO "product_attribute"
						(name, product_id, status)
					VALUES
						(?, (SELECT id FROM new_1_product), ?)
					RETURNING *),
				new_1_order AS
					(INSERT INTO "order"
						(product_id)
					VALUES
						((SELECT id FROM new_1_product))
					RETURNING *)
			SELECT * FROM new_1_order`,
		expectedArgs: []any{
			true, "Child Bicycle", // new_1_product
			"FDA approved", "NotApplicable", // new_1_product_attribute
			"Toy-safety-council approved", "Yes", // new_2_product_attribute
		},
	},
	{
		name: "create_relationships_1_to_M_to_1",
		keelSchema: `
			model Order {
				fields {
					items OrderItem[]
				}
				actions {
					create createOrder() with (items.quantity, items.product.name)
				}
				@permission(expression: true, actions: [create])
			}
			model Product {
				fields {
					name Text
					isActive Boolean @default(true)
					createdOnOrder Boolean @default(false)
				}
			}
			model OrderItem {
				fields {
					order Order
					quantity Text
					product Product
					isReturned Boolean @default(false)
				}
			}`,
		actionName: "createOrder",
		input: map[string]any{
			"items": []any{
				map[string]any{
					"quantity": 2,
					"product": map[string]any{
						"name": "Hair dryer",
					},
				},
				map[string]any{
					"quantity": 4,
					"product": map[string]any{
						"name": "Hair clips",
					},
				},
			},
		},
		expectedTemplate: `
			WITH
				new_1_order AS
					(INSERT INTO "order"
					DEFAULT VALUES
					RETURNING *),
				new_1_product AS
					(INSERT INTO "product"
						(name)
					VALUES
						(?)
					RETURNING *),
				new_1_order_item AS
					(INSERT INTO "order_item"
						(order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), (SELECT id FROM new_1_product), ?)
					RETURNING *),
				new_2_product AS
					(INSERT INTO "product"
						(name)
					VALUES
						(?)
					RETURNING *),
				new_2_order_item AS
					(INSERT INTO "order_item"
						(order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), (SELECT id FROM new_2_product), ?)
					RETURNING *)
			SELECT * FROM new_1_order`,
		expectedArgs: []any{
			"Hair dryer", // new_1_product
			2,            //new_1_order_item
			"Hair clips", // new_2_product
			4,            //new_2_order_item
		},
	},
	{
		name: "create_relationships_M_to_1_multiple",
		keelSchema: `
			model Order {
				fields {
					product1 Product
					product2 Product
				}
				actions {
					create createOrder() with (product1.name, product2.name)
				}
				@permission(expression: true, actions: [create])
			}
			model Product {
				fields {
					name Text
					isActive Boolean @default(true)
				}
			}`,
		actionName: "createOrder",
		input: map[string]any{
			"product1": map[string]any{
				"name": "Child Bicycle",
			},
			"product2": map[string]any{
				"name": "Adult Bicycle",
			},
		},
		expectedTemplate: `
			WITH
				new_1_product AS
					(INSERT INTO "product"
						(name)
					VALUES
						(?)
					RETURNING *),
				new_2_product AS
					(INSERT INTO "product"
						(name)
					VALUES
						(?)
					RETURNING *),
				new_1_order AS
					(INSERT INTO "order"
						(product_1_id, product_2_id)
					VALUES
						((SELECT id FROM new_1_product), (SELECT id FROM new_2_product))
					RETURNING *)
			SELECT * FROM new_1_order`,
		expectedArgs: []any{
			"Child Bicycle", // new_1_product
			"Adult Bicycle", // new_2_product
		},
	},
	{
		name: "create_relationships_1_to_M_multiple",
		keelSchema: `
			model Order {
				fields {
					items OrderItem[]
					freeItems OrderItem[]
				}
				actions {
					create createOrder() with (items.quantity, items.product.id, freeItems.quantity, freeItems.product.id)
				}
				@permission(expression: true, actions: [create])
			}
			model Product {
				fields {
					name Text
				}
			}
			model OrderItem {
				fields {
					order Order? @relation(items)
					freeOnOrder Order? @relation(freeItems)
					quantity Text
					product Product
				}
			}`,
		actionName: "createOrder",
		input: map[string]any{
			"items": []any{
				map[string]any{
					"quantity": 2,
					"product": map[string]any{
						"id": "paid1",
					},
				},
				map[string]any{
					"quantity": 4,
					"product": map[string]any{
						"id": "paid2",
					},
				},
			},
			"freeItems": []any{
				map[string]any{
					"quantity": 6,
					"product": map[string]any{
						"id": "free1",
					},
				},
				map[string]any{
					"quantity": 8,
					"product": map[string]any{
						"id": "free2",
					},
				},
			},
		},
		expectedTemplate: `
			WITH
				new_1_order AS
					(INSERT INTO "order"
					DEFAULT VALUES
					RETURNING *),
				new_1_order_item AS
					(INSERT INTO "order_item"
						(order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), ?, ?)
					RETURNING *),
				new_2_order_item AS
					(INSERT INTO "order_item"
						(order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), ?, ?)
					RETURNING *),
				new_3_order_item AS
					(INSERT INTO "order_item"
						(free_on_order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), ?, ?)
					RETURNING *),
				new_4_order_item AS
					(INSERT INTO "order_item"
						(free_on_order_id, product_id, quantity)
					VALUES
						((SELECT id FROM new_1_order), ?, ?)
					RETURNING *)
			SELECT * FROM new_1_order`,
		expectedArgs: []any{
			"paid1", 2, // new_1_order_item
			"paid2", 4, // new_2_order_item
			"free1", 6, // new_3_order_item
			"free2", 8, // new_4_order_item
		},
	},
	{
		name: "update_by_unique_key",
		keelSchema: `
			model Product {
				fields {
					barcode Text @unique
					isActive Boolean @default(true)
				}
				actions {
					update deactivateProduct(barcode) {
						@where(product.isActive)
						@set(product.isActive = false)
						@permission(expression: true)
					}
				}
			}`,
		actionName: "deactivateProduct",
		input: map[string]any{
			"where": map[string]any{
				"barcode": "123",
			},
		},
		expectedTemplate: `
			UPDATE "product"
			SET is_active = ?
			WHERE
				"product"."barcode" IS NOT DISTINCT FROM ? AND
				"product"."is_active" IS NOT DISTINCT FROM ?
			RETURNING "product".*`,
		expectedArgs: []any{false, "123", true},
	},
	{
		name: "update_by_unique_composite_key",
		keelSchema: `
			model Brand {
				fields {
					code Text @unique
					products Product[]
				}
			}
			model Product {
				fields {
					productCode Text
					brand Brand
					isActive Boolean @default(true)
				}
				actions {
					update deactivateProduct(productCode, brand.code) {
						@set(product.isActive = false)
						@permission(expression: true)
					}
				}
				@unique([productCode, brand])
			}`,
		actionName: "deactivateProduct",
		input: map[string]any{
			"where": map[string]any{
				"productCode": "prodcode",
				"brandCode":   "brand",
			},
		},
		expectedTemplate: `
			UPDATE "product"
			SET is_active = ?
			FROM "brand" AS "product$brand"
			WHERE
				"product$brand"."id" = "product"."brand_id" AND
				"product"."product_code" IS NOT DISTINCT FROM ? AND
				"product$brand"."code" IS NOT DISTINCT FROM ?
			RETURNING "product".*`,
		expectedArgs: []any{false, "prodcode", "brand"},
	},
	{
		name: "update_by_unique_composite_key_and_filters",
		keelSchema: `
			model Supplier {
				fields {
					products Product[]
					isRegistered Boolean
				}
			}
			model Brand {
				fields {
					code Text @unique
					products Product[]
				}
			}
			model Product {
				fields {
					productCode Text
					brand Brand
					supplier Supplier
					isActive Boolean @default(true)
				}
				actions {
					update deactivateProduct(productCode, brand.code) {
						@where(product.isActive)
						@where(product.supplier.isRegistered)
						@set(product.isActive = false)
						@permission(expression: true)
					}
				}
				@unique([productCode, brand])
			}`,
		actionName: "deactivateProduct",
		input: map[string]any{
			"where": map[string]any{
				"productCode": "prodcode",
				"brandCode":   "brand",
			},
		},
		expectedTemplate: `
			UPDATE "product"
			SET is_active = ?
			FROM "brand" AS "product$brand"
			LEFT JOIN "supplier" AS "product$supplier" ON "product$supplier"."id" = "product"."supplier_id"
			WHERE
				"product$brand"."id" = "product"."brand_id" AND
				"product"."product_code" IS NOT DISTINCT FROM ? AND
				"product$brand"."code" IS NOT DISTINCT FROM ? AND
				"product"."is_active" IS NOT DISTINCT FROM ? AND
				"product$supplier"."is_registered" IS NOT DISTINCT FROM ?
			RETURNING "product".*`,
		expectedArgs: []any{false, "prodcode", "brand", true, true},
	},
	{
		name: "delete_by_unique_key",
		keelSchema: `
			model Product {
				fields {
					barcode Text @unique
					isActive Boolean @default(true)
				}
				actions {
					delete deleteProduct(barcode) {
						@where(product.isActive)
						@permission(expression: true)
					}
				}
			}`,
		actionName: "deleteProduct",
		input: map[string]any{
			"barcode": "123",
		},
		expectedTemplate: `
			DELETE FROM "product"
			WHERE
				"product"."barcode" IS NOT DISTINCT FROM ? AND
				"product"."is_active" IS NOT DISTINCT FROM ?
			RETURNING "product"."id"`,
		expectedArgs: []any{"123", true},
	},
	{
		name: "delete_by_unique_composite_key",
		keelSchema: `
			model Brand {
				fields {
					code Text @unique
					products Product[]
				}
			}
			model Product {
				fields {
					productCode Text
					brand Brand
					isActive Boolean @default(true)
				}
				actions {
					delete deleteProduct(productCode, brand.code) {
						@permission(expression: true)
					}
				}
				@unique([productCode, brand])
			}`,
		actionName: "deleteProduct",
		input: map[string]any{
			"productCode": "prodcode",
			"brandCode":   "brand",
		},
		expectedTemplate: `
			DELETE FROM "product"
			USING "brand" AS "product$brand"
			WHERE
				"product"."product_code" IS NOT DISTINCT FROM ? AND
				"product$brand"."code" IS NOT DISTINCT FROM ?
			RETURNING "product"."id"`,
		expectedArgs: []any{"prodcode", "brand"},
	},
	{
		name: "delete_by_unique_composite_key_and_filters",
		keelSchema: `
			model Supplier {
				fields {
					products Product[]
					isLocked Boolean
				}
			}
			model Brand {
				fields {
					code Text @unique
					products Product[]
				}
			}
			model Product {
				fields {
					productCode Text
					brand Brand
					supplier Supplier
					isActive Boolean @default(true)
				}
				actions {
					delete deleteProduct(productCode, brand.code) {
						@where(product.isActive)
						@where(product.supplier.isLocked)
						@permission(expression: true)
					}
				}
				@unique([productCode, brand])
			}`,
		actionName: "deleteProduct",
		input: map[string]any{
			"productCode": "prodcode",
			"brandCode":   "brand",
		},
		expectedTemplate: `
			DELETE FROM "product"
			USING "brand" AS "product$brand", "supplier" AS "product$supplier"
			WHERE
				"product"."product_code" IS NOT DISTINCT FROM ? AND
				"product$brand"."code" IS NOT DISTINCT FROM ? AND
				"product"."is_active" IS NOT DISTINCT FROM ? AND
				"product$supplier"."is_locked" IS NOT DISTINCT FROM ?
			RETURNING "product"."id"`,
		expectedArgs: []any{"prodcode", "brand", true, true},
	},
	{
		name: "create_set_ctx_identity_fields",
		keelSchema: `
			model Person {
				fields {
					email Text
					created Timestamp
					emailVerified Boolean
					externalId Text
					issuer Text
				}
				actions {
					create createPerson() {
						@set(person.email = ctx.identity.email)
						@set(person.created = ctx.identity.createdAt)
						@set(person.emailVerified = ctx.identity.emailVerified)
						@set(person.externalId = ctx.identity.externalId)
						@set(person.issuer = ctx.identity.issuer)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "createPerson",
		input:      map[string]any{},
		identity:   identity,
		expectedTemplate: `
			WITH
				select_identity (column_0, column_1, column_2, column_3, column_4) AS (
					SELECT "identity"."email", "identity"."created_at", "identity"."email_verified", "identity"."external_id", "identity"."issuer"
					FROM "identity"
					WHERE "identity"."id" IS NOT DISTINCT FROM ?),
				new_1_person AS (
					INSERT INTO "person" (created, email, email_verified, external_id, issuer)
					VALUES (
						(SELECT column_1 FROM select_identity),
						(SELECT column_0 FROM select_identity),
						(SELECT column_2 FROM select_identity),
						(SELECT column_3 FROM select_identity),
						(SELECT column_4 FROM select_identity))
					RETURNING *)
			SELECT *, set_identity_id(?) AS __keel_identity_id FROM new_1_person`,
		expectedArgs: []any{identity[parser.FieldNameId].(string), identity[parser.FieldNameId].(string)},
	},
	{
		name: "update_set_ctx_identity_fields",
		keelSchema: `
			model Person {
				fields {
					email Text
					created Timestamp
					emailVerified Boolean
					externalId Text
					issuer Text
				}
				actions {
					update updatePerson(id) {
						@set(person.email = ctx.identity.email)
						@set(person.created = ctx.identity.createdAt)
						@set(person.emailVerified = ctx.identity.emailVerified)
						@set(person.externalId = ctx.identity.externalId)
						@set(person.issuer = ctx.identity.issuer)
					}
				}
				@permission(expression: true, actions: [create])
			}`,
		actionName: "updatePerson",
		input:      map[string]any{"where": map[string]any{"id": "xyz"}},
		identity:   identity,
		expectedTemplate: `
			WITH
				select_identity (column_0, column_1, column_2, column_3, column_4) AS (
					SELECT "identity"."email", "identity"."created_at", "identity"."email_verified", "identity"."external_id", "identity"."issuer"
					FROM "identity"
					WHERE "identity"."id" IS NOT DISTINCT FROM ?)
			UPDATE "person" SET
				created = (SELECT column_1 FROM select_identity),
				email = (SELECT column_0 FROM select_identity),
				email_verified = (SELECT column_2 FROM select_identity),
				external_id = (SELECT column_3 FROM select_identity),
				issuer = (SELECT column_4 FROM select_identity)
			WHERE "person"."id" IS NOT DISTINCT FROM ?
			RETURNING "person".*, set_identity_id(?) AS __keel_identity_id`,
		expectedArgs: []any{identity[parser.FieldNameId].(string), "xyz", identity[parser.FieldNameId].(string)},
	},
	{
		name: "create_array",
		keelSchema: `
			model Post {
				fields {
					title Text
					tags Text[]
				}
				actions {
					create createPost() with (title, tags)
				}
			}`,
		actionName: "createPost",
		input:      map[string]any{"title": "Hello world", "tags": []string{"science", "politics"}},
		expectedTemplate: `
			WITH new_1_post AS (
				INSERT INTO "post" (tags, title) 
				VALUES (ARRAY[?, ?]::TEXT[], ?)
				RETURNING *) 
			SELECT * FROM new_1_post`,
		expectedArgs: []any{"science", "politics", "Hello world"},
	},
	{
		name: "create_array_set_attribute_empty",
		keelSchema: `
			model Post {
				fields {
					title Text
					tags Text[]
				}
				actions {
					create createPost() with (title) {
						@set(post.tags = [])
					}
				}
			}`,
		actionName: "createPost",
		input:      map[string]any{"title": "Hello world"},
		expectedTemplate: `
			WITH new_1_post AS (
				INSERT INTO "post" (tags, title) 
				VALUES ('{}', ?)
				RETURNING *) 
			SELECT * FROM new_1_post`,
		expectedArgs: []any{"Hello world"},
	},
	{
		name: "create_array_set_attribute",
		keelSchema: `
			model Post {
				fields {
					title Text
					tags Text[]
				}
				actions {
					create createPost() with (title) {
						@set(post.tags = ["science", "technology"])
					}
				}
			}`,
		actionName: "createPost",
		input:      map[string]any{"title": "Hello world"},
		expectedTemplate: `
			WITH new_1_post AS (
				INSERT INTO "post" (tags, title) 
				VALUES (ARRAY[?, ?]::TEXT[], ?)
				RETURNING *) 
			SELECT * FROM new_1_post`,
		expectedArgs: []any{"science", "technology", "Hello world"},
	},
	{
		name: "list_array_literal_expression",
		keelSchema: `
			model Post {
				fields {
					title Text
					tags Text[]
				}
				actions {
					list listPosts() {
						@where(post.title in ["1", "2"])
					}
				}
			}`,
		actionName: "listPosts",
		expectedTemplate: `
			SELECT 
			DISTINCT ON("post"."id") "post".*, CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, (SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE "post"."title" = ANY(ARRAY[?, ?]::TEXT[])) AS totalCount 
			FROM "post" 
			WHERE "post"."title" = ANY(ARRAY[?, ?]::TEXT[]) 
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"1", "2", "1", "2", 50},
	},
	{
		name: "list_array_implicit_equals",
		keelSchema: `
			model Post {
				fields {
					tags Text[]
				}
				actions {
					list listPosts(tags)
				}
			}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"tags": map[string]any{"equals": []any{"science"}}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE "post"."tags" IS NOT DISTINCT FROM ARRAY[?]::TEXT[]) AS totalCount 	
			FROM "post" 
			WHERE "post"."tags" IS NOT DISTINCT FROM ARRAY[?]::TEXT[] 
			ORDER BY "post"."id" ASC LIMIT ?`,
		expectedArgs: []any{"science", "science", 50},
	},
	{
		name: "list_array_implicit_equals_empty_array",
		keelSchema: `
			model Post {
				fields {
					tags Text[]
				}
				actions {
					list listPosts(tags)
				}
			}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"tags": map[string]any{"equals": []any{}}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE "post"."tags" IS NOT DISTINCT FROM '{}') AS totalCount 	
			FROM "post" 
			WHERE "post"."tags" IS NOT DISTINCT FROM '{}'
			ORDER BY "post"."id" ASC LIMIT ?`,
		expectedArgs: []any{50},
	},
	{
		name: "list_array_implicit_not_equals",
		keelSchema: `
			model Post {
				fields {
					tags Text[]
				}
				actions {
					list listPosts(tags)
				}
			}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"tags": map[string]any{"notEquals": []any{"science"}}}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE "post"."tags" IS DISTINCT FROM ARRAY[?]::TEXT[]) AS totalCount 	
			FROM "post" 
			WHERE "post"."tags" IS DISTINCT FROM ARRAY[?]::TEXT[] 
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"science", "science", 50},
	},
	{
		name: "list_array_implicit_any_equals",
		keelSchema: `
		model Post {
			fields {
				tags Text[]
			}
			actions {
				list listPosts(tags)
			}
		}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"tags": map[string]any{"any": map[string]any{"equals": "science"}}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE ? = ANY("post"."tags")) AS totalCount 	
			FROM "post" 
			WHERE ? = ANY("post"."tags")
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"science", "science", 50},
	},
	{
		name: "list_array_implicit_all_equals",
		keelSchema: `
		model Post {
			fields {
				tags Text[]
			}
			actions {
				list listPosts(tags)
			}
		}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"tags": map[string]any{"all": map[string]any{"equals": "science"}}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE (? = ALL("post"."tags") AND "post"."tags" IS DISTINCT FROM '{}')) AS totalCount 	
			FROM "post" 
			WHERE (? = ALL("post"."tags") AND "post"."tags" IS DISTINCT FROM '{}')
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"science", "science", 50},
	},
	{
		name: "list_array_implicit_all_less_than",
		keelSchema: `
		model Post {
			fields {
				votes Number[]
			}
			actions {
				list listPosts(votes)
			}
		}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"votes": map[string]any{"all": map[string]any{"lessThan": 5}}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE ? > ALL("post"."votes")) AS totalCount 	
			FROM "post" 
			WHERE ? > ALL("post"."votes")
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{5, 5, 50},
	},
	{
		name: "list_array_implicit_any_less_than",
		keelSchema: `
		model Post {
			fields {
				votes Number[]
			}
			actions {
				list listPosts(votes)
			}
		}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"votes": map[string]any{"any": map[string]any{"lessThan": 5}}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE ? > ANY("post"."votes")) AS totalCount 	
			FROM "post" 
			WHERE ? > ANY("post"."votes")
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{5, 5, 50},
	},
	{
		name: "list_array_implicit_all_after",
		keelSchema: `
		model Post {
			fields {
				editedAt Timestamp[]
			}
			actions {
				list listPosts(editedAt)
			}
		}`,
		actionName: "listPosts",
		input:      map[string]any{"where": map[string]any{"editedAt": map[string]any{"all": map[string]any{"after": "2024-01-01T00:12:00Z"}}}},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE ? < ALL("post"."edited_at")) AS totalCount 	
			FROM "post" 
			WHERE ? < ALL("post"."edited_at")
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"2024-01-01T00:12:00Z", "2024-01-01T00:12:00Z", 50},
	},
	{
		name: "list_array_expression_in",
		keelSchema: `
			model Post {
				fields {
					tags Text[]
				}
				actions {
					list listSciencePosts() {
						@where("science" in post.tags)
					}
				}
			}`,
		actionName: "listSciencePosts",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE ? = ANY("post"."tags")) AS totalCount 	
			FROM "post" 
			WHERE ? = ANY("post"."tags")
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"science", "science", 50},
	},
	{
		name: "list_array_expression_not_in",
		keelSchema: `
			model Post {
				fields {
					tags Text[]
				}
				actions {
					list listSciencePosts() {
						@where("science" not in post.tags)
					}
				}
			}`,
		actionName: "listSciencePosts",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE (NOT ? = ANY("post"."tags") OR "post"."tags" IS NOT DISTINCT FROM NULL)) AS totalCount 	
			FROM "post" 
			WHERE (NOT ? = ANY("post"."tags") OR "post"."tags" IS NOT DISTINCT FROM NULL)
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"science", "science", 50},
	},
	{
		name: "list_array_expression_equals",
		keelSchema: `
			model Post {
				fields {
					tags Text[]
				}
				actions {
					list listSciencePosts() {
						@where(["science", "tech"] == post.tags)
					}
				}
			}`,
		actionName: "listSciencePosts",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE ARRAY[?, ?]::TEXT[] IS NOT DISTINCT FROM "post"."tags") AS totalCount 	
			FROM "post" 
			WHERE ARRAY[?, ?]::TEXT[] IS NOT DISTINCT FROM "post"."tags"
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"science", "tech", "science", "tech", 50},
	},
	{
		name: "list_array_expression_not_equals",
		keelSchema: `
			model Post {
				fields {
					tags Text[]
				}
				actions {
					list listSciencePosts() {
						@where(["science", "tech"] != post.tags)
					}
				}
			}`,
		actionName: "listSciencePosts",
		input:      map[string]any{},
		expectedTemplate: `
			SELECT
				DISTINCT ON("post"."id") "post".*, 
				CASE WHEN LEAD("post"."id") OVER (ORDER BY "post"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT "post"."id") FROM "post" WHERE ARRAY[?, ?]::TEXT[] IS DISTINCT FROM "post"."tags") AS totalCount 	
			FROM "post" 
			WHERE ARRAY[?, ?]::TEXT[] IS DISTINCT FROM "post"."tags"
			ORDER BY "post"."id" ASC 
			LIMIT ?`,
		expectedArgs: []any{"science", "tech", "science", "tech", 50},
	},
	{
		name: "list_nested_array_expression_in",
		keelSchema: `
			model Collection {
				fields {
					name Text
					books Book[]
				}
				actions {
					list listInCollection(genre: Text) {
						@where(genre in collection.books.genres)
						@orderBy(name: asc)
					}
				}
			}
			model Book {
				fields {
					col Collection
					genres Text[]
				}
			}`,
		actionName: "listInCollection",
		input:      map[string]any{"where": map[string]any{"genre": "fantasy"}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("collection"."name", "collection"."id") "collection".*, 
				CASE WHEN LEAD("collection"."id") OVER (ORDER BY "collection"."name" ASC, "collection"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT ("collection"."name", "collection"."id")) 
					FROM "collection" 
					LEFT JOIN "book" AS "collection$books" ON "collection$books"."col_id" = "collection"."id" 
					WHERE ? = ANY("collection$books"."genres")) AS totalCount 
			FROM "collection" 
			LEFT JOIN "book" AS "collection$books" ON "collection$books"."col_id" = "collection"."id" 
			WHERE ? = ANY("collection$books"."genres") 
			ORDER BY "collection"."name" ASC, "collection"."id" ASC LIMIT ?`,
		expectedArgs: []any{"fantasy", "fantasy", 50},
	},
	{
		name: "list_nested_array_expression_not_in",
		keelSchema: `
			model Collection {
				fields {
					name Text
					books Book[]
				}
				actions {
					list listInCollection(genre: Text) {
						@where(genre not in collection.books.genres)
						@orderBy(name: asc)
					}
				}
			}
			model Book {
				fields {
					col Collection
					genres Text[]
				}
			}`,
		actionName: "listInCollection",
		input:      map[string]any{"where": map[string]any{"genre": "fantasy"}},
		expectedTemplate: `
			SELECT 
				DISTINCT ON("collection"."name", "collection"."id") "collection".*, 
				CASE WHEN LEAD("collection"."id") OVER (ORDER BY "collection"."name" ASC, "collection"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, 
				(SELECT COUNT(DISTINCT ("collection"."name", "collection"."id")) 
					FROM "collection" 
					LEFT JOIN "book" AS "collection$books" ON "collection$books"."col_id" = "collection"."id" 
					WHERE (NOT ? = ANY("collection$books"."genres") OR "collection$books"."genres" IS NOT DISTINCT FROM NULL)) AS totalCount 
			FROM "collection" 
			LEFT JOIN "book" AS "collection$books" ON "collection$books"."col_id" = "collection"."id" 
			WHERE (NOT ? = ANY("collection$books"."genres") OR "collection$books"."genres" IS NOT DISTINCT FROM NULL) 
			ORDER BY "collection"."name" ASC, "collection"."id" ASC LIMIT ?`,
		expectedArgs: []any{"fantasy", "fantasy", 50},
	},
	{
		name: "list_nested_array_expression_not_in",
		keelSchema: `
			model Collection {
				fields {
					name Text
					books Book[]
				}
				actions {
					list suggestedCollections() {
						@where(ctx.identity.person.favouriteGenre in collection.books.genres)
						@orderBy(name: asc)
						@permission(expression: true)
					}
				}
			}
			model Book {
				fields {
					col Collection
					genres Text[]
				}
			}
			model Person {
				fields {
					favouriteGenre Text
					identity Identity @unique
				}
			}`,
		actionName: "suggestedCollections",
		input:      map[string]any{},
		identity:   identity,
		expectedTemplate: `
			SELECT 
				DISTINCT ON("collection"."name", "collection"."id") "collection".*, 
				CASE WHEN LEAD("collection"."id") OVER (ORDER BY "collection"."name" ASC, "collection"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, (
					SELECT COUNT(DISTINCT ("collection"."name", "collection"."id")) 
					FROM "collection" LEFT JOIN "book" AS "collection$books" ON "collection$books"."col_id" = "collection"."id" 
					WHERE (SELECT "identity$person"."favourite_genre" FROM "identity" LEFT JOIN "person" AS "identity$person" ON "identity$person"."identity_id" = "identity"."id" WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$person"."favourite_genre" IS DISTINCT FROM NULL) = ANY("collection$books"."genres")) AS totalCount 
				FROM "collection" 
				LEFT JOIN "book" AS "collection$books" ON "collection$books"."col_id" = "collection"."id" 
				WHERE (SELECT "identity$person"."favourite_genre" 
						FROM "identity" 
						LEFT JOIN "person" AS "identity$person" ON "identity$person"."identity_id" = "identity"."id" 
						WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$person"."favourite_genre" IS DISTINCT FROM NULL) = ANY("collection$books"."genres") 
				ORDER BY "collection"."name" ASC, "collection"."id" ASC LIMIT ?`,
		expectedArgs: []any{identity["id"].(string), identity["id"].(string), 50},
	},
	{
		name: "list_nested_array_expression_not_in1",
		keelSchema: `
			model Collection {
				fields {
					name Text
					books Book[]
				}
			}
			model Book {
				fields {
					col Collection
					genre Text
				}
				actions {
					list suggestedBooks() {
						@where(book.genre in ctx.identity.person.favouriteGenres)
						@permission(expression: true)
					}
				}
			}
			model Person {
				fields {
					favouriteGenres Text[]
					identity Identity @unique
				}
			}`,
		actionName: "suggestedBooks",
		input:      map[string]any{},
		identity:   identity,
		expectedTemplate: `
			SELECT 
				DISTINCT ON("book"."id") "book".*, 
				CASE WHEN LEAD("book"."id") OVER (ORDER BY "book"."id" ASC) IS NOT NULL THEN true ELSE false END AS hasNext, (
					SELECT COUNT(DISTINCT "book"."id") 
					FROM "book" 
					WHERE "book"."genre" IN (
						SELECT unnest("identity$person"."favourite_genres") 
						FROM "identity" 
						LEFT JOIN "person" AS "identity$person" ON "identity$person"."identity_id" = "identity"."id" 
						WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$person"."favourite_genres" IS DISTINCT FROM NULL)) AS totalCount 
				FROM "book" 
				WHERE "book"."genre" IN (
					SELECT unnest("identity$person"."favourite_genres") 
					FROM "identity" 
					LEFT JOIN "person" AS "identity$person" ON "identity$person"."identity_id" = "identity"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$person"."favourite_genres" IS DISTINCT FROM NULL) 
				ORDER BY "book"."id" ASC LIMIT ?`,
		expectedArgs: []any{identity["id"].(string), identity["id"].(string), 50},
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
				statement, err = actions.GenerateGetStatement(query, scope, testCase.input)
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

// Generates a scope and query builder
func generateQueryScope(ctx context.Context, schemaString string, actionName string) (*actions.Scope, *actions.QueryBuilder, *proto.Action, error) {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(schemaString, config.Empty)
	if err != nil {
		return nil, nil, nil, err
	}

	action := proto.FindAction(schema, actionName)
	if action == nil {
		return nil, nil, nil, fmt.Errorf("action not found in schema: %s", actionName)
	}

	model := proto.FindModel(schema.Models, action.ModelName)
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

func TestInsertStatement(t *testing.T) {
	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	query.AddWriteValues(map[string]*actions.QueryOperand{"name": actions.Value("Fred")})
	query.Select(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.InsertStatement(context.Background())

	expected := `
		WITH new_1_person AS (INSERT INTO "person" (name) VALUES (?) RETURNING *)
		SELECT * FROM new_1_person`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestUpdateStatement(t *testing.T) {
	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	query.AddWriteValue(actions.Field("name"), actions.Value("Fred"))
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.Select(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.UpdateStatement(context.Background())

	expected := `
		UPDATE "person" SET name = ? WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING "person".*`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestDeleteStatement(t *testing.T) {
	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.Select(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.DeleteStatement(context.Background())

	expected := `
		DELETE FROM "person" WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING "person".*`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestInsertStatementWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx = withIdentity(ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	query.AddWriteValues(map[string]*actions.QueryOperand{"name": actions.Value("Fred")})
	query.Select(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.InsertStatement(ctx)

	expected := `
		WITH new_1_person AS (INSERT INTO "person" (name) VALUES (?) RETURNING *)
		SELECT
			*,
			set_identity_id(?) AS __keel_identity_id,
			set_trace_id(?) AS __keel_trace_id
		FROM new_1_person`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
	require.Equal(t, "Fred", stmt.SqlArgs()[0])
	require.Equal(t, "2V1gEtq4GEhvtRofqwiN9ZfapxN", stmt.SqlArgs()[1])
	require.Equal(t, "71f835dc7ac2750bed2135c7b30dc7fe", stmt.SqlArgs()[2])
}

func TestUpdateStatementWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx = withIdentity(ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	query.AddWriteValue(actions.Field("name"), actions.Value("Fred"))
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.Select(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.UpdateStatement(ctx)

	expected := `
		UPDATE "person" SET name = ? WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING
			"person".*,
			set_identity_id(?) AS __keel_identity_id,
			set_trace_id(?) AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
	require.Equal(t, "Fred", stmt.SqlArgs()[0])
	require.Equal(t, "1234", stmt.SqlArgs()[1])
	require.Equal(t, "2V1gEtq4GEhvtRofqwiN9ZfapxN", stmt.SqlArgs()[2])
	require.Equal(t, "71f835dc7ac2750bed2135c7b30dc7fe", stmt.SqlArgs()[3])
}

func TestUpdateStatementNoReturnsWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx = withIdentity(ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	query.AddWriteValue(actions.Field("name"), actions.Value("Fred"))
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.Select(actions.AllFields())
	stmt := query.UpdateStatement(ctx)

	expected := `
		UPDATE "person" SET name = ? WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING
			set_identity_id(?) AS __keel_identity_id,
			set_trace_id(?) AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
	require.Equal(t, "Fred", stmt.SqlArgs()[0])
	require.Equal(t, "1234", stmt.SqlArgs()[1])
	require.Equal(t, "2V1gEtq4GEhvtRofqwiN9ZfapxN", stmt.SqlArgs()[2])
	require.Equal(t, "71f835dc7ac2750bed2135c7b30dc7fe", stmt.SqlArgs()[3])
}

func TestDeleteStatementWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx = withIdentity(ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.Select(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.DeleteStatement(ctx)

	expected := `
		DELETE FROM "person" WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING
			"person".*,
			set_identity_id(?) AS __keel_identity_id,
			set_trace_id(?) AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
	require.Equal(t, "1234", stmt.SqlArgs()[0])
	require.Equal(t, "2V1gEtq4GEhvtRofqwiN9ZfapxN", stmt.SqlArgs()[1])
	require.Equal(t, "71f835dc7ac2750bed2135c7b30dc7fe", stmt.SqlArgs()[2])
}

func TestDeleteStatementNoReturnWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx = withIdentity(ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(model)
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.Select(actions.AllFields())
	stmt := query.DeleteStatement(ctx)

	expected := `
		DELETE FROM "person" WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING
			set_identity_id(?) AS __keel_identity_id,
			set_trace_id(?) AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
	require.Equal(t, "1234", stmt.SqlArgs()[0])
	require.Equal(t, "2V1gEtq4GEhvtRofqwiN9ZfapxN", stmt.SqlArgs()[1])
	require.Equal(t, "71f835dc7ac2750bed2135c7b30dc7fe", stmt.SqlArgs()[2])
}

func withIdentity(ctx context.Context) context.Context {
	identity := auth.Identity{"id": identityId}
	return auth.WithIdentity(ctx, identity)
}

const (
	identityId = "2V1gEtq4GEhvtRofqwiN9ZfapxN"
	traceId    = "71f835dc7ac2750bed2135c7b30dc7fe"
	spanId     = "b4c9e2a6a0d84702"
)

func withTracing(t *testing.T, ctx context.Context) context.Context {
	traceIdBytes, err := hex.DecodeString(traceId)
	require.NoError(t, err)
	spanIdBytes, err := hex.DecodeString(spanId)
	require.NoError(t, err)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID(traceIdBytes),
		SpanID:     trace.SpanID(spanIdBytes),
		TraceFlags: trace.FlagsSampled,
	})
	require.True(t, spanContext.IsValid())
	return trace.ContextWithSpanContext(ctx, spanContext)
}

func TestParseArray(t *testing.T) {
	scanTests := []struct {
		in  string
		out []string
	}{
		{"{one,two}", []string{"one", "two"}},
		{`{"one, sdf",two}`, []string{"one, sdf", "two"}},
		{`{"\"one\"",two}`, []string{`"one"`, "two"}},
		{`{"\\one\\",two}`, []string{`\one\`, "two"}},
		{`{"{one}",two}`, []string{`{one}`, "two"}},
		{`{"one two"}`, []string{`one two`}},
		{`{"one,two"}`, []string{`one,two`}},
		{`{abcdef:83bf98cc-fec9-4e77-b4cf-99f9fb6655fa-0NH:zxcvzxc:wers:vxdfw-asdf-asdf}`, []string{"abcdef:83bf98cc-fec9-4e77-b4cf-99f9fb6655fa-0NH:zxcvzxc:wers:vxdfw-asdf-asdf"}},
		{`{"",two}`, []string{"", "two"}},
		{`{" ","NULL"}`, []string{" ", "NULL"}},
		{`{"something",NULL}`, []string{"something", "NULL"}},
		{`{}`, []string{}},
	}

	for _, testCase := range scanTests {
		res, err := actions.ParsePostgresArray(testCase.in, func(s string) (string, error) { return s, nil })
		assert.NoError(t, err)
		assert.Equal(t, testCase.out, res)
	}
}
