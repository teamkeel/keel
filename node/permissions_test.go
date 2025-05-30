package node

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/schema"
)

type fixture struct {
	name     string
	schema   string
	expected string
	sql      string
}

func TestPermissionFnBuilder(t *testing.T) {
	t.Parallel()
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
export const permissionFns = {
	createPost: [],
}
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
export const permissionFns = {
	createPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "person"."id" 
				FROM "person" 
				WHERE true AND "person"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."publish_date" <= ${ctx.now()} AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	updatePost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				LEFT JOIN "identity" AS "post$identity" ON "post"."identity_id" = "post$identity"."id" 
				WHERE "post$identity"."email" IS NOT DISTINCT FROM ${"adam@keel.xyz"} AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."view_count" < ${10} AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	updatePost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."identity_id" IS NOT DISTINCT FROM ${ctx.identity ? ctx.identity.id : ''} AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	updatePost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				LEFT JOIN "identity" AS "post$identity" ON "post"."identity_id" = "post$identity"."id" 
				WHERE "post$identity"."email" IS NOT DISTINCT FROM ${ctx.identity ? ctx.identity.email : ''} AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	createPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE ${ctx.isAuthenticated}::boolean AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE ${ctx.headers["secretkey"] || ""} IS NOT DISTINCT FROM "post"."secret_key" AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
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
export const permissionFns = {
	getPost: [
		async (records, ctx, db) => {
			const { rows } = await sql%s.execute(db);
			return rows.length === records.length;
		},
	],
}
			`,
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE ${ctx.secrets["SECRET_KEY"] || ""} IS NOT DISTINCT FROM "post"."secret_key" AND "post"."id" IN (${(records.length > 0) ? sql.join(records.map(x => x.id)) : []})
			`,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			w := codegen.Writer{}
			builder := schema.Builder{}
			config := `
secrets:
  - name: SECRET_KEY`

			schema, err := builder.MakeFromString(fixture.schema, config)

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
