package node

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
)

type fixture struct {
	name     string
	schema   string
	expected string
	sql      string
}

func TestPermissionFnBuilder(t *testing.T) {
	fixtures := []fixture{
		{
			name: "NoPermissionRules",
			schema: `
				model Person {
					fields {
						name Text
					}

					actions {
						create createPost() with(name) @function
					}
				}
			`,
			expected: `
const permissionFns = {
	createPost: [],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "ValueRecordIDs",
			schema: `
				model Person {
					fields {
						name Text
					}

					actions {
						create createPost() with(name) @function
					}

					@permission(expression: true, actions: [list, update, get, create, delete])
				}
			`,
			expected: `
const permissionFns = {
	createPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "person"."id", (true) AS "result" 
				FROM "person" 
				WHERE "person"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "person"."id"
			`,
		},
		{
			name: "ValueNow",
			schema: `
				model Post {
					fields {
						publishDate Date
					}

					actions {
						get getPost(id) @function
					}

					@permission(expression: post.publishDate <= ctx.now , actions: [get])
				}
			`,
			expected: `
const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", ("post"."publish_date" <= ${ctx.now()}) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "post"."id"
			`,
		},
		{
			name: "ValueString",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					actions {
						update updatePost() with(title) {
							@permission(expression: post.identity.email == "adam@keel.xyz")
							@function
						}
					}
				}
			`,
			expected: `
const permissionFns = {
	updatePost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", ("post$identity"."email" IS NOT DISTINCT FROM ${"adam@keel.xyz"}) AS "result" 
				FROM "post" 
				LEFT JOIN "identity" AS "post$identity" ON "post"."identity_id" = "post$identity"."id" 
				WHERE "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "post"."id", "post$identity"."email"
			`,
		},
		{
			name: "ValueNumber",
			schema: `
				model Post {
					fields {
						viewCount Number
					}
					actions {
						get getPost(id) {
							@permission(expression: post.viewCount < 10)
							@function
						}
					}
				}
			`,
			expected: `
const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", ("post"."view_count" < ${10}) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "post"."id"
			`,
		},
		{
			name: "ValueIdentityID",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					actions {
						update updatePost() with(title) {
							@permission(expression: post.identity == ctx.identity)
							@function
						}
					}
				}
			`,
			expected: `
const permissionFns = {
	updatePost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", ("post"."identity_id" IS NOT DISTINCT FROM ${ctx.identity ? ctx.identity.id : ''}) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "post"."id"
			`,
		},
		{
			name: "ValueIdentityEmail",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					actions {
						update updatePost() with(title) {
							@permission(expression: post.identity.email == ctx.identity.email)
							@function
						}
					}
				}
			`,
			expected: `
const permissionFns = {
	updatePost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", ("post$identity"."email" IS NOT DISTINCT FROM ${ctx.identity ? ctx.identity.email : ''}) AS "result" 
				FROM "post" 
				LEFT JOIN "identity" AS "post$identity" ON "post"."identity_id" = "post$identity"."id" 
				WHERE "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "post"."id", "post$identity"."email"
			`,
		},
		{
			name: "ValueIsAuthenticated",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					actions {
						update createPost() with(title) {
							@permission(expression: ctx.isAuthenticated)
							@function
						}
					}
				}
			`,
			expected: `
const permissionFns = {
	createPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", (${ctx.isAuthenticated}::boolean) AS "result" 
				FROM "post"
				WHERE "post"."id" 
				IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "post"."id"
			`,
		},
		{
			name: "ValueHeader",
			schema: `
				model Post {
					fields {
						secretKey Text
					}
					actions {
						get getPost(id) {
							@permission(expression: ctx.headers.secretkey == post.secretKey)
							@function
						}
					}
				}
			`,
			expected: `
const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", (${ctx.headers["secretkey"] || ""} IS NOT DISTINCT FROM "post"."secret_key") AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}) 
				GROUP BY "post"."id"
			`,
		},
		{
			name: "ValueSecret",
			schema: `
				model Post {
					fields {
						secretKey Text
					}
					actions {
						get getPost(id) {
							@permission(expression: ctx.secrets.SECRET_KEY == post.secretKey)
							@function
						}
					}
				}
			`,
			expected: `
const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length && rows.every(x => x.result);
		},
	],
}
module.exports.permissionFns = permissionFns;
			`,
			sql: `
				SELECT "post"."id", (${ctx.secrets["SECRET_KEY"] || ""} IS NOT DISTINCT FROM "post"."secret_key") AS "result"
				FROM "post"
				WHERE "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
				GROUP BY "post"."id"
			`,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			w := codegen.Writer{}
			builder := schema.Builder{
				Config: &config.ProjectConfig{
					Secrets: []config.Input{
						{
							Name: "SECRET_KEY",
						},
					},
				},
			}
			schema, err := builder.MakeFromString(fixture.schema)

			require.NoError(t, err)
			writePermissions(&w, schema)

			expected := fixture.expected
			if fixture.sql != "" {
				sql := strings.Join(strings.Fields(strings.TrimSpace(fixture.sql)), " ")
				expected = fmt.Sprintf(fixture.expected, "`"+sql+"`")
			}

			assert.Equal(t, normalise(expected), normalise(w.String()))
		})
	}
}
