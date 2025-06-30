package actions_test

import (
	"context"
	"testing"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/test-go/testify/assert"
)

const ctxTestSchema = `
model User {
	fields {
		name Text
		team CompanyTeam
		identity Identity @unique
	}
}
model CompanyTeam {
	fields {
		name Text
		users User[]
	}
}`

type ctxTestCase struct {
	name        string
	keelSchema  string
	expression  string
	expectedSql string
}

var ctxTestCases = []ctxTestCase{
	{
		name:        "ctx identity backlink",
		keelSchema:  ctxTestSchema,
		expression:  `ctx.identity.user.team.name == "myTeam"`,
		expectedSql: `SELECT COUNT(*) FROM "identity" LEFT JOIN "user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id" LEFT JOIN "company_team" AS "identity$user$team" ON "identity$user$team"."id" = "identity$user"."team_id" WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$user$team"."name" IS NOT DISTINCT FROM ?`,
	},
	{
		name:        "multiple ctx identity backlink conditions",
		keelSchema:  ctxTestSchema,
		expression:  `ctx.identity.user.team.name == "myTeam" && ctx.identity.user.name == "John"`,
		expectedSql: `SELECT COUNT(*) FROM "identity" LEFT JOIN "user" AS "identity$user" ON "identity$user"."identity_id" = "identity"."id" LEFT JOIN "company_team" AS "identity$user$team" ON "identity$user$team"."id" = "identity$user"."team_id" WHERE "identity"."id" IS NOT DISTINCT FROM ? AND "identity$user$team"."name" IS NOT DISTINCT FROM ? AND "identity"."id" IS NOT DISTINCT FROM ? AND "identity$user"."name" IS NOT DISTINCT FROM ?`,
	},
}

func TestGeneratedCtxQuery(t *testing.T) {
	for _, testCase := range ctxTestCases {
		schemaFiles :=
			&reader.Inputs{
				SchemaFiles: []*reader.SchemaFile{
					{
						Contents: testCase.keelSchema,
						FileName: "schema.keel",
					},
				},
			}

		builder := &schema.Builder{}
		schema, err := builder.MakeFromInputs(schemaFiles)
		assert.NoError(t, err)

		expression, err := parser.ParseExpression(testCase.expression)
		assert.NoError(t, err)

		query := actions.NewQuery(proto.FindModel(schema.GetModels(), "Identity"))
		query.SelectClause("COUNT(*)")

		query, err = resolve.RunCelVisitor(expression, actions.GenerateCtxQuery(context.Background(), query, schema))
		assert.NoError(t, err)

		stmt := query.SelectStatement()
		assert.Equal(t, testCase.expectedSql, stmt.SqlTemplate(), "expected `%s` but got `%s`", testCase.expectedSql, stmt.SqlTemplate())
	}
}
