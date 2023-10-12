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
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/schema"
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
	identity *auth.Identity
	// Expected SQL template generated (with ? placeholders for values)
	expectedTemplate string
	// OPTIONAL: Expected ordered argument slice
	expectedArgs []any
}

var identity = &auth.Identity{
	Id:    "identityId",
	Email: "keelson@keel.xyz",
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
		name: "create_op_set_attribute",
		keelSchema: `
			model Person {
				fields {
					name Text
					age Number
					isActive Boolean
				}
				actions {
					create createPerson() {
						@set(person.name = "Bob")
						@set(person.age = 100)
						@set(person.isActive = true)
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
						(age, is_active, name) 
					VALUES 
						(?, ?, ?) 
					RETURNING *) 
			SELECT * FROM new_1_person`,
		expectedArgs: []any{int64(100), true, "Bob"},
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
				}
				actions {
					update updatePerson(id) {
						@set(person.name = "Bob")
						@set(person.age = 100)
						@set(person.isActive = true)
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
			    age = ?, is_active = ?, name = ?
			WHERE
				"person"."id" IS NOT DISTINCT FROM ?
			RETURNING 
				"person".*`,
		expectedArgs: []any{int64(100), true, "Bob", "xyz"},
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
								(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."name" IN (?, ?, ?, ?)) AS totalCount
            FROM 
                "thing" 
            WHERE
                "thing"."name" IN (?, ?, ?, ?)
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
								(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."category" IN (?, ?)) AS totalCount
            FROM 
                "thing" 
            WHERE
                "thing"."category" IN (?, ?)
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
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."title" IN (?, ?)) AS totalCount
			FROM 
				"thing" 
			WHERE 
				"thing"."title" IN (?, ?)
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
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."title" NOT IN (?, ?)) AS totalCount
			FROM 
				"thing" 
			WHERE 
				"thing"."title" NOT IN (?, ?)
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
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."age" IN (?, ?)) AS totalCount
			FROM 
				"thing" 
			WHERE 
				"thing"."age" IN (?, ?)
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
				(SELECT COUNT(DISTINCT "thing"."id") FROM "thing" WHERE "thing"."age" NOT IN (?, ?)) AS totalCount
			FROM 
				"thing" 
			WHERE 
				"thing"."age" NOT IN (?, ?)
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
		WITH new_1_customer AS (INSERT INTO "customer" (name) VALUES (?) RETURNING *) SELECT * FROM new_1_customer`,
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
				"thing"."id" ASC LIMIT ?`,
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
}

func TestQueryBuilder(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
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
func generateQueryScope(ctx context.Context, schemaText string, actionName string) (*actions.Scope, *actions.QueryBuilder, *proto.Action, error) {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(schemaText)
	if err != nil {
		return nil, nil, nil, err
	}

	action := proto.FindAction(schema, actionName)
	if action == nil {
		return nil, nil, nil, fmt.Errorf("action not found in schema: %s", actionName)
	}

	model := proto.FindModel(schema.Models, action.ModelName)
	query := actions.NewQuery(context.Background(), model)
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
	query := actions.NewQuery(context.Background(), model)
	query.AddWriteValues(map[string]any{"name": "Fred"})
	query.AppendSelect(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.InsertStatement()

	expected := `
		WITH new_1_person AS (INSERT INTO "person" (name) VALUES (?) RETURNING *) 
		SELECT * FROM new_1_person`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestUpdateStatement(t *testing.T) {
	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(context.Background(), model)
	query.AddWriteValue(actions.Field("name"), "Fred")
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.AppendSelect(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.UpdateStatement()

	expected := `
		UPDATE "person" SET name = ? WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING "person".*`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestDeleteStatement(t *testing.T) {
	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(context.Background(), model)
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.AppendSelect(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.DeleteStatement()

	expected := `
		DELETE FROM "person" WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING "person".*`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestInsertStatementWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx, _ = withIdentity(t, ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(ctx, model)
	query.AddWriteValues(map[string]any{"name": "Fred"})
	query.AppendSelect(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.InsertStatement()

	expected := `
		WITH new_1_person AS (INSERT INTO "person" (name) VALUES (?) RETURNING *) 
		SELECT 
			*, 
			set_identity_id('2V1gEtq4GEhvtRofqwiN9ZfapxN') AS __keel_identity_id, 
			set_trace_id('71f835dc7ac2750bed2135c7b30dc7fe') AS __keel_trace_id
		FROM new_1_person`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestUpdateStatementWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx, _ = withIdentity(t, ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(ctx, model)
	query.AddWriteValue(actions.Field("name"), "Fred")
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.AppendSelect(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.UpdateStatement()

	expected := `
		UPDATE "person" SET name = ? WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING 
			"person".*, 
			set_identity_id('2V1gEtq4GEhvtRofqwiN9ZfapxN') AS __keel_identity_id, 
			set_trace_id('71f835dc7ac2750bed2135c7b30dc7fe') AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestUpdateStatementNoReturnsWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx, _ = withIdentity(t, ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(ctx, model)
	query.AddWriteValue(actions.Field("name"), "Fred")
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.AppendSelect(actions.AllFields())
	stmt := query.UpdateStatement()

	expected := `
		UPDATE "person" SET name = ? WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING 
			set_identity_id('2V1gEtq4GEhvtRofqwiN9ZfapxN') AS __keel_identity_id, 
			set_trace_id('71f835dc7ac2750bed2135c7b30dc7fe') AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestDeleteStatementWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx, _ = withIdentity(t, ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(ctx, model)
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.AppendSelect(actions.AllFields())
	query.AppendReturning(actions.AllFields())
	stmt := query.DeleteStatement()

	expected := `
		DELETE FROM "person" WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING 
			"person".*, 
			set_identity_id('2V1gEtq4GEhvtRofqwiN9ZfapxN') AS __keel_identity_id, 
			set_trace_id('71f835dc7ac2750bed2135c7b30dc7fe') AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func TestDeleteStatementNoReturnWithAuditing(t *testing.T) {
	ctx := context.Background()
	ctx, _ = withIdentity(t, ctx)
	ctx = withTracing(t, ctx)

	model := &proto.Model{Name: "Person"}
	query := actions.NewQuery(ctx, model)
	err := query.Where(actions.IdField(), actions.Equals, actions.Value("1234"))
	require.NoError(t, err)
	query.AppendSelect(actions.AllFields())
	stmt := query.DeleteStatement()

	expected := `
		DELETE FROM "person" WHERE "person"."id" IS NOT DISTINCT FROM ? RETURNING 
			set_identity_id('2V1gEtq4GEhvtRofqwiN9ZfapxN') AS __keel_identity_id, 
			set_trace_id('71f835dc7ac2750bed2135c7b30dc7fe') AS __keel_trace_id`

	require.Equal(t, clean(expected), clean(stmt.SqlTemplate()))
}

func withIdentity(t *testing.T, ctx context.Context) (context.Context, *auth.Identity) {
	identity := &auth.Identity{Id: identityId}
	return auth.WithIdentity(ctx, identity), identity
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
