package permissions_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/permissions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
)

func TestToSQL(t *testing.T) {
	t.Parallel()
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."public" IS NOT DISTINCT FROM false AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."public" IS NOT DISTINCT FROM true AND "post"."id" IN (?)
			`,
			values: []permissions.Value{
				{
					Type: permissions.ValueRecordIDs,
				},
			},
		},
		{
			name: "boolean_expression",
			schema: `
				model Post {
					fields {
						isPublic Boolean
					}
					actions {
						get getPost(id)
					}
					@permission(
						expression: post.isPublic,
						actions: [get]
					)
				}
			`,
			action: "getPost",
			sql: `
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."is_public" AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."title" IS NOT DISTINCT FROM ? AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."view_count" < ? AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."identity_id" IS NOT DISTINCT FROM null AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id"
				FROM "post" 
				WHERE "post"."identity_id" IS DISTINCT FROM null AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "project"."id" 
				FROM "project" 
				WHERE "project"."visibility" IS NOT DISTINCT FROM ? AND "project"."id" IN (?)
			`,
			values: []permissions.Value{
				{
					Type:        permissions.ValueString,
					StringValue: "\"Public\"",
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE ? IS NOT DISTINCT FROM "post"."secret_key" AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE ? IS NOT DISTINCT FROM "post"."secret_key" AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" LEFT JOIN "author" AS "post$author" ON "post"."author_id" = "post$author"."id" 
				WHERE "post$author"."identity_id" IS NOT DISTINCT FROM ? AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				LEFT JOIN "author" AS "post$author" ON "post"."author_id" = "post$author"."id" 
				LEFT JOIN "account" AS "post$author$account" ON "post$author"."account_id" = "post$author$account"."id" 
				WHERE "post$author$account"."identity_id" IS NOT DISTINCT FROM ? AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "project"."id" 
				FROM "project" 
				LEFT JOIN "account" AS "project$accounts" ON "project"."id" = "project$accounts"."project_id"
				WHERE ? IS NOT DISTINCT FROM "project$accounts"."identity_id" AND "project"."id" IN (?)
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
					actions {
						get getProject(id)
					}
					@permission(
						expression: project.identity == ctx.identity || (project.public && ctx.isAuthenticated == false),
						actions: [get]
					)
				}
			`,
			action: "getProject",
			sql: `
				SELECT DISTINCT "project"."id" FROM "project" 
				WHERE 
					("project"."identity_id" IS NOT DISTINCT FROM ? or 
					"project"."public" and ?::boolean IS NOT DISTINCT FROM false) AND "project"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE "post"."publish_date" <= ? AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				WHERE ("post"."publish_date" <= ? or "post"."identity_id" IS NOT DISTINCT FROM ?) AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "post"."id" 
				FROM "post" 
				LEFT JOIN "account" AS "post$account" ON "post"."account_id" = "post$account"."id" 
				WHERE ("post$account"."identity_id" IS NOT DISTINCT FROM ? or "post$account"."posts_are_public") AND "post"."id" IN (?)
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
					actions {
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
				SELECT DISTINCT "join"."id" 
				FROM "join" 
				LEFT JOIN "table" AS "join$inner" ON "join"."inner_id" = "join$inner"."id" 
				LEFT JOIN "select" AS "join$inner$group" ON "join$inner"."group_id" = "join$inner$group"."id" 
				WHERE "join$inner$group"."by_id" IS NOT DISTINCT FROM ? AND "join"."id" IN (?)
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
			name: "identity_backlink_compare_model",
			schema: `
				model UserProfile {
					fields {
						identity Identity @unique
					}
				}

				model Post {
					fields {
						profile UserProfile
					}

					actions {
						get getPost(id)
					}
				}

				model Comment {
					fields {
						post Post
						text Text
					}

					actions {
						create addComment() with (post.id, text)
					}

					@permission(
						expression: ctx.identity.userProfile == comment.post.profile,
						actions: [create]
					)
				}
			`,
			action: "addComment",
			sql: `
				SELECT DISTINCT "comment"."id" 
				FROM "comment" 
				LEFT JOIN "post" AS "comment$post" 
				ON "comment"."post_id" = "comment$post"."id" 
				WHERE 
					(SELECT "identity$user_profile"."id" 
					FROM "identity" 
					LEFT JOIN "user_profile" AS "identity$user_profile" 
					ON "identity"."id" = "identity$user_profile"."identity_id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ?) 
					IS NOT DISTINCT FROM "comment$post"."profile_id"
				AND "comment"."id" IN (?)
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
			name: "identity_backlink_compare_field",
			schema: `
				model UserProfile {
					fields {
						identity Identity @unique
					}
				}

				model Post {
					fields {
						profile UserProfile
					}

					actions {
						get getPost(id)
					}
				}

				model Comment {
					fields {
						post Post
						text Text
					}

					actions {
						create addComment() with (post.id, text)
					}

					@permission(
						expression: ctx.identity.userProfile.id == comment.post.profile.id,
						actions: [create]
					)
				}
			`,
			action: "addComment",
			sql: `
				SELECT DISTINCT "comment"."id" 
				FROM "comment" 
				LEFT JOIN "post" AS "comment$post" 
				ON "comment"."post_id" = "comment$post"."id" 
				LEFT JOIN "user_profile" AS "comment$post$profile" ON "comment$post"."profile_id" = "comment$post$profile"."id" 
				WHERE 
					(SELECT "identity$user_profile"."id" 
					FROM "identity" 
					LEFT JOIN "user_profile" AS "identity$user_profile" 
					ON "identity"."id" = "identity$user_profile"."identity_id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ?) IS NOT DISTINCT FROM "comment$post$profile"."id" 
				AND "comment"."id" IN (?)
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
			name: "identity_backlink_model_many_to_many",
			schema: `
				model User {
					fields {
						identity Identity @unique
						orgs UserOrg[]
					}
					actions {
						list listUserByOrg(orgs.org.id)
					}
					@permission(
						expression: ctx.identity.user in user.orgs.org.users.user,
						actions: [list]
					)
				}

				model Org {
					fields {
						users UserOrg[]
					}
				}

				model UserOrg {
					fields {
						user User
						org Org
					}

					@unique([user, org])
				}
			`,
			action: "listUserByOrg",
			sql: `
				SELECT DISTINCT "user"."id" 
				FROM "user" 
				LEFT JOIN "user_org" AS "user$orgs" ON "user"."id" = "user$orgs"."user_id" 
				LEFT JOIN "org" AS "user$orgs$org" ON "user$orgs"."org_id" = "user$orgs$org"."id" 
				LEFT JOIN "user_org" AS "user$orgs$org$users" ON "user$orgs$org"."id" = "user$orgs$org$users"."org_id" 
				WHERE (SELECT "identity$user"."id" 
					FROM "identity" 
					LEFT JOIN "user" AS "identity$user" ON "identity"."id" = "identity$user"."identity_id"
					WHERE "identity"."id" IS NOT DISTINCT FROM ?) IS NOT DISTINCT FROM "user$orgs$org$users"."user_id" 
				AND "user"."id" IN (?)
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
			name: "identity_backlink_literal_in_subquery",
			schema: `
				model User {
					fields {
						identity Identity @unique
						roles UserRole[]
					}
				}
				model UserRole {
					fields {
						user User
						role Role
					}
					@unique([user, role])
				}
				model Role {
					fields {
						name Text
						permissions RolePermission[]
					}
				}
				model RolePermission {
					fields {
						role Role
						permission Permission
					}
				}
				model Permission {
					fields {
						name Text
					}
				}
				model Account {
					fields {
						name Text
					}
					actions {
						list listAccount() {
							@function
							@permission(expression: "account:list" in ctx.identity.user.roles.role.permissions.permission.name)
						}
					}
				}

			`,
			action: "listAccount",
			sql: `
				SELECT DISTINCT "account"."id" 
				FROM "account" 
				WHERE ? IN 
					(SELECT "identity$user$roles$role$permissions$permission"."name" 
					FROM "identity" 
					LEFT JOIN "user" AS "identity$user" ON "identity"."id" = "identity$user"."identity_id" 
					LEFT JOIN "user_role" AS "identity$user$roles" ON "identity$user"."id" = "identity$user$roles"."user_id" 
					LEFT JOIN "role" AS "identity$user$roles$role" ON "identity$user$roles"."role_id" = "identity$user$roles$role"."id" 
					LEFT JOIN "role_permission" AS "identity$user$roles$role$permissions" ON "identity$user$roles$role"."id" = "identity$user$roles$role$permissions"."role_id" 
					LEFT JOIN "permission" AS "identity$user$roles$role$permissions$permission" ON "identity$user$roles$role$permissions"."permission_id" = "identity$user$roles$role$permissions$permission"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ?) 
				AND "account"."id" IN (?)
			`,
			values: []permissions.Value{
				{
					Type:        permissions.ValueString,
					StringValue: "\"account:list\"",
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
			name: "identity_backlink_field_in_subquery",
			schema: `
				model User {
					fields {
						identity Identity @unique
						roles UserRole[]
					}
				}
				model UserRole {
					fields {
						user User
						role Role
					}
					@unique([user, role])
				}
				model Role {
					fields {
						name Text
						permissions RolePermission[]
					}
				}
				model RolePermission {
					fields {
						role Role
						permission Permission
					}
				}
				model Permission {
					fields {
						name Text
					}
				}
				model Account {
					fields {
						name Text
						permission Text
					}
					actions {
						list listAccount() {
							@function
							@permission(expression: account.permission in ctx.identity.user.roles.role.permissions.permission.name)
						}
					}
				}

			`,
			action: "listAccount",
			sql: `
				SELECT DISTINCT "account"."id" 
				FROM "account" 
				WHERE "account"."permission" IN 
					(SELECT "identity$user$roles$role$permissions$permission"."name" 
					FROM "identity" 
					LEFT JOIN "user" AS "identity$user" ON "identity"."id" = "identity$user"."identity_id" 
					LEFT JOIN "user_role" AS "identity$user$roles" ON "identity$user"."id" = "identity$user$roles"."user_id" 
					LEFT JOIN "role" AS "identity$user$roles$role" ON "identity$user$roles"."role_id" = "identity$user$roles$role"."id" 
					LEFT JOIN "role_permission" AS "identity$user$roles$role$permissions" ON "identity$user$roles$role"."id" = "identity$user$roles$role$permissions"."role_id" 
					LEFT JOIN "permission" AS "identity$user$roles$role$permissions$permission" ON "identity$user$roles$role$permissions"."permission_id" = "identity$user$roles$role$permissions$permission"."id" 
					WHERE "identity"."id" IS NOT DISTINCT FROM ?) 
				AND "account"."id" IN (?)
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
			t.Parallel()
			builder := &schema.Builder{}

			config := `
secrets: 
  - name: SECRET_KEY
`

			s, err := builder.MakeFromString(fixture.schema, config)
			require.NoError(t, err)

			var model *proto.Model
			var action *proto.Action
			for _, m := range s.GetModels() {
				for _, a := range m.GetActions() {
					if a.GetName() == fixture.action {
						action = a
						model = m
					}
				}
			}

			perms := proto.PermissionsForAction(s, action)

			sql, values, err := permissions.ToSQL(s, model, perms)
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

			ctx := t.Context()

			ctx, err = testhelpers.WithTracing(ctx)
			require.NoError(t, err)

			dbName := testhelpers.DbNameForTestName(t.Name())
			database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, s, dbName, true)
			require.NoError(t, err)
			defer database.Close()

			// Can use nil for all values as we're only testing SQL is valid
			vals := []any{}
			for range values {
				vals = append(vals, nil)
			}

			// Execute the query runs without error
			_, err = database.ExecuteQuery(t.Context(), sql, vals...)
			require.NoError(t, err)
		})
	}
}

func clean(sql string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
}
