package permissions_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/permissions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
)

func TestToSQL(t *testing.T) {

	type Fixture struct {
		name   string
		schema string
		action string
		sql    string
		values []permissions.Value
	}

	fixtures := []Fixture{
		{
			name: "equals_false",
			schema: `
				model Post {
					fields {
						public Boolean
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.public == false,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post"."public" IS NOT DISTINCT FROM false) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "equals_true",
			schema: `
				model Post {
					fields {
						public Boolean
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.public == true,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post"."public" IS NOT DISTINCT FROM true) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "equals_string",
			schema: `
				model Post {
					fields {
						title Text
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.title == "Foo",
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post"."title" IS NOT DISTINCT FROM ?) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type:        permissions.ValueString,
					StringValue: `"Foo"`,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "equals_number",
			schema: `
				model Post {
					fields {
						viewCount Number
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.viewCount < 10,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post"."view_count" < ?) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type:        permissions.ValueNumber,
					NumberValue: 10,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "equals_null",
			schema: `
				model Post {
					fields {
						identity Identity?
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.identity == null,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
			SELECT "post"."id", ("post"."identity_id" IS NOT DISTINCT FROM null) AS "result" 
			FROM "post" 
			WHERE "post"."id" IN (?) 
			GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "not_equals_null",
			schema: `
				model Post {
					fields {
						identity Identity?
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.identity != null,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post"."identity_id" IS DISTINCT FROM null) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "equals_enum",
			schema: `
				enum Visibility {
					Public
					Private
				}
				model Project {
					fields {
						visibility Visibility
					}
					operations {
						get getProject(id)
					}
					@permission(
						expression: project.visibility == Visibility.Public,
						actions: [get]
					)
				}
			`,
			action: "getProject",
			sql: `
				SELECT "project"."id", ("project"."visibility" IS NOT DISTINCT FROM ?) AS "result" 
				FROM "project" 
				WHERE "project"."id" IN (?) 
				GROUP BY "project"."id"
			`,
			values: []permissions.Value{
				{
					Type:        permissions.ValueString,
					StringValue: "Public",
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "equals_header",
			schema: `
				model Post {
					fields {
						secretKey Text
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: ctx.headers.secretkey == post.secretKey,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", (? IS NOT DISTINCT FROM "post"."secret_key") AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type:      permissions.ValueHeader,
					HeaderKey: "secretkey",
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "equals_secret",
			schema: `
				model Post {
					fields {
						secretKey Text
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: ctx.secrets.SECRET_KEY == post.secretKey,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", (? IS NOT DISTINCT FROM "post"."secret_key") AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type:      permissions.ValueSecret,
					SecretKey: "SECRET_KEY",
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "belongs_to_join",
			schema: `
				model Author {
					fields {
						identity Identity
					}
				}
				model Post {
					fields {
						author Author
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.author.identity == ctx.identity,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post$author"."identity_id" IS NOT DISTINCT FROM ?) AS "result" 
				FROM "post" 
				LEFT JOIN "author" AS "post$author" ON "post"."author_id" = "post$author"."id" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id", "post$author"."identity_id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueIdentityID,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "belongs_to_join_multiple",
			schema: `
				model Account {
					fields {
						identity Identity
					}
				}
				model Author {
					fields {
						account Account
					}
				}
				model Post {
					fields {
						author Author
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.author.account.identity == ctx.identity,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post$author$account"."identity_id" IS NOT DISTINCT FROM ?) AS "result" 
				FROM "post" 
				LEFT JOIN "author" AS "post$author" ON "post"."author_id" = "post$author"."id" 
				LEFT JOIN "account" AS "post$author$account" ON "post$author"."account_id" = "post$author$account"."id" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id", "post$author$account"."identity_id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueIdentityID,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "has_many_join",
			schema: `
				model Account {
					fields {
						identity Identity
						project Project
					}
				}
				model Project {
					fields {
						accounts Account[]
					}
					operations {
						get getProject(id)
					}
					@permission(
						expression: ctx.identity in project.accounts.identity,
						actions: [get]
					)
				}
			`,
			action: "getProject",
			sql: `
				SELECT "project"."id", (? = ANY(ARRAY_AGG("project$accounts"."identity_id"))) AS "result" 
				FROM "project" 
				LEFT JOIN "account" AS "project$accounts" ON "project"."id" = "project$accounts"."project_id" 
				WHERE "project"."id" IN (?) 
				GROUP BY "project"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueIdentityID,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "grouped_expression",
			schema: `
				model Project {
					fields {
						identity Identity
						public Boolean
					}
					operations {
						get getProject(id)
					}
					@permission(
						expression: project.identity == ctx.identity or (project.public and ctx.isAuthenticated == false),
						actions: [get]
					)
				}
			`,
			action: "getProject",
			sql: `
				SELECT "project"."id", ("project"."identity_id" IS NOT DISTINCT FROM ? or ("project"."public" and ?::boolean IS NOT DISTINCT FROM false)) AS "result" 
				FROM "project" 
				WHERE "project"."id" IN (?) 
				GROUP BY "project"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueIdentityID,
				},
				{
					Type: permissions.ValueIsAuthenticated,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "ctx_now",
			schema: `
				model Post {
					fields {
						publishDate Date
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.publishDate <= ctx.now,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post"."publish_date" <= ?) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueNow,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "multiple_permission_rules",
			schema: `
				model Post {
					fields {
						identity Identity
						publishDate Date
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.publishDate <= ctx.now,
						actions: [get]
					)
					@permission(
						expression: post.identity == ctx.identity,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post"."publish_date" <= ?) or ("post"."identity_id" IS NOT DISTINCT FROM ?) AS "result" 
				FROM "post" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueNow,
				},
				{
					Type: permissions.ValueIdentityID,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "deduped_joins",
			schema: `
				model Account {
					fields {
						identity Identity
						postsArePublic Boolean
					}
				}
				model Post {
					fields {
						account Account
					}
					operations {
						get getPost(id)
					}
					@permission(
						expression: post.account.identity == ctx.identity,
						actions: [get]
					)
					@permission(
						expression: post.account.postsArePublic,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT "post"."id", ("post$account"."identity_id" IS NOT DISTINCT FROM ?) or ("post$account"."posts_are_public") AS "result" 
				FROM "post" 
				LEFT JOIN "account" AS "post$account" ON "post"."account_id" = "post$account"."id" 
				WHERE "post"."id" IN (?) 
				GROUP BY "post"."id", "post$account"."identity_id", "post$account"."posts_are_public"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueIdentityID,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "quoted_identifiers",
			schema: `
				model Select {
					fields {
						by Identity
					}
				}
				model Table {
					fields {
						group Select
					}
				}
				model Join {
					fields {
						inner Table
					}
					operations {
						get getJoin(id)
					}
					@permission(
						expression: join.inner.group.by == ctx.identity,
						actions: [get]
					)
				}
			`,
			action: "getJoin",
			sql: `
				SELECT "join"."id", ("join$inner$group"."by_id" IS NOT DISTINCT FROM ?) AS "result" 
				FROM "join" 
				LEFT JOIN "table" AS "join$inner" ON "join"."inner_id" = "join$inner"."id" 
				LEFT JOIN "select" AS "join$inner$group" ON "join$inner"."group_id" = "join$inner$group"."id" 
				WHERE "join"."id" IN (?) 
				GROUP BY "join"."id", "join$inner$group"."by_id"
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueIdentityID,
				},
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			builder := &schema.Builder{
				Config: &config.ProjectConfig{
					Secrets: []config.Input{
						{
							Name: "SECRET_KEY",
						},
					},
				},
			}

			s, err := builder.MakeFromString(fixture.schema)
			require.NoError(t, err)

			var model *proto.Model
			var action *proto.Operation
			for _, m := range s.Models {
				for _, a := range m.Operations {
					if a.Name == fixture.action {
						action = a
						model = m
					}
				}
			}

			sql, values, err := permissions.ToSQL(s, model, action)
			require.NoError(t, err)

			// Assert SQL is as expecte
			assert.Equal(t, clean(fixture.sql), clean(sql))
			assert.Len(t, values, len(fixture.values))

			// Assert values as expected
			for i, v := range fixture.values {
				assert.Equal(t, v.Type, values[i].Type)
				switch v.Type {
				case permissions.ValueString:
					assert.Equal(t, v.StringValue, values[i].StringValue)
				case permissions.ValueNumber:
					assert.Equal(t, v.NumberValue, values[i].NumberValue)
				case permissions.ValueHeader:
					assert.Equal(t, v.HeaderKey, values[i].HeaderKey)
				case permissions.ValueSecret:
					assert.Equal(t, v.SecretKey, values[i].SecretKey)
				}
			}

			dbConnInfo := &db.ConnectionInfo{
				Host:     "localhost",
				Port:     "8001",
				Username: "postgres",
				Database: "keel",
				Password: "postgres",
			}

			// Setup test database
			ctx := context.Background()
			dbName := testhelpers.DbNameForTestName(t.Name())
			database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, s, dbName)
			require.NoError(t, err)
			defer database.Close()

			// Can use nil for all values as we're only testing SQL is valid
			vals := []any{}
			for range values {
				vals = append(vals, nil)
			}

			// Execute the query runs without error
			_, err = database.ExecuteQuery(context.Background(), sql, vals...)
			require.NoError(t, err)
		})
	}
}

func clean(sql string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
}