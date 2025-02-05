package actions_test

import (
	"context"
	"strings"
	"testing"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/test-go/testify/assert"
)

const setTestSchema = `
model Item {
	fields {
		price Decimal
		product Product
		quantity Number
		
	}
	actions {
		create createItem() with (product.id, quantity) {
			#placeholder#
		}
	}
}

model Product {
	fields {
		name Text
		standardPrice Decimal
		sku Text
	}
}`

type setTestCase struct {
	// Name given to the test case
	name string
	// Valid keel schema for this test case
	keelSchema string
	// action name to run test upon
	field string
	// Input map for action
	expectedSql string
}

var setTestCases = []setTestCase{

	{
		name:        "adding field with literal",
		keelSchema:  setTestSchema,
		field:       "@set(item.price = item.product.standardPrice)",
		expectedSql: `r."price" + 100`,
	},
}

func TestGeneratedSelect(t *testing.T) {
	t.Parallel()
	for _, testCase := range setTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			raw := strings.Replace(testCase.keelSchema, "#placeholder#", testCase.field, 1)

			schemaFiles :=
				&reader.Inputs{
					SchemaFiles: []*reader.SchemaFile{
						{
							Contents: raw,
							FileName: "schema.keel",
						},
					},
				}

			builder := &schema.Builder{}
			schema, err := builder.MakeFromInputs(schemaFiles)
			assert.NoError(t, err)

			model := schema.Models[0]
			actionName := "createItem"
			action := schema.FindAction(actionName)

			expression, err := parser.ParseExpression(action.SetExpressions[0].Source)
			assert.NoError(t, err)

			_, rhs, err := expression.ToAssignmentExpression()
			assert.NoError(t, err)

			sql, err := resolve.RunCelVisitor(rhs, actions.GenerateSelectQuery(context.Background(), schema, model, action, map[string]any{}))
			assert.NoError(t, err)

			assert.Equal(t, testCase.expectedSql, sql, "expected `%s` but got `%s`", testCase.expectedSql, sql)
		})
	}
}
