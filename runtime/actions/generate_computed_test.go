package actions_test

import (
	"strings"
	"testing"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/test-go/testify/assert"
)

const testSchema = `
model Item {
	fields {
		price Decimal?
		product Product
		quantity Number
		isActive Boolean
		#placeholder#
	}
}
model Product {
	fields {
		name Text
		standardPrice Decimal
		sku Text
		agent Agent
	}
}
model Agent {
	fields {
		commission Decimal
	}
}`

type computedTestCase struct {
	// Name given to the test case
	name string
	// Valid keel schema for this test case
	keelSchema string
	// action name to run test upon
	field string
	// Input map for action
	expectedSql string
}

var computedTestCases = []computedTestCase{

	{
		name:        "adding field with literal",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.price + 100)",
		expectedSql: `r."price" + 100`,
	},
	{
		name:        "subtracting field with literal",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.price - 100)",
		expectedSql: `r."price" - 100`,
	},
	{
		name:        "dividing field with literal",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.price / 100)",
		expectedSql: `r."price" / 100`,
	},
	{
		name:        "multiplying field with literal",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.price * 100)",
		expectedSql: `r."price" * 100`,
	},
	{
		name:        "multiply fields on same model",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.price * item.quantity)",
		expectedSql: `r."price" * r."quantity"`,
	},
	{
		name:        "parenthesis",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.quantity * (1 + item.quantity) / (100 * ((item.price + 1) * 1)))",
		expectedSql: `r."quantity" * (1 + r."quantity") / (100 * ((r."price" + 1) * 1))`,
	},
	{
		name:        "no parenthesis",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.quantity * 1 + item.quantity / 100 * item.price + 1)",
		expectedSql: `r."quantity" * 1 + r."quantity" / 100 * r."price" + 1`,
	},
	{
		name:        "unnecessary parenthesis",
		keelSchema:  testSchema,
		field:       "total Decimal @computed((item.quantity * 1) + (item.quantity / 100 * item.price) + 1)",
		expectedSql: `r."quantity" * 1 + r."quantity" / 100 * r."price" + 1`,
	},
	{
		name:        "bool greater than",
		keelSchema:  testSchema,
		field:       "isExpensive Boolean @computed(item.price > 100)",
		expectedSql: `r."price" > 100`,
	},
	{
		name:        "bool greater or equals",
		keelSchema:  testSchema,
		field:       "isExpensive Boolean @computed(item.price >= 100)",
		expectedSql: `r."price" >= 100`,
	},
	{
		name:        "bool less than",
		keelSchema:  testSchema,
		field:       "isCheap Boolean @computed(item.price < 100)",
		expectedSql: `r."price" < 100`,
	},
	{
		name:        "bool less or equals",
		keelSchema:  testSchema,
		field:       "isCheap Boolean @computed(item.price <= 100)",
		expectedSql: `r."price" <= 100`,
	},
	{
		name:        "bool is not null",
		keelSchema:  testSchema,
		field:       "hasPrice Boolean @computed(item.price != null)",
		expectedSql: `r."price" IS DISTINCT FROM NULL`,
	},
	{
		name:        "bool is null",
		keelSchema:  testSchema,
		field:       "noPrice Boolean @computed(item.price == null)",
		expectedSql: `r."price" IS NOT DISTINCT FROM NULL`,
	},
	{
		name:        "bool with and",
		keelSchema:  testSchema,
		field:       "isExpensive Boolean @computed(item.price > 100 && item.isActive)",
		expectedSql: `r."price" > 100 AND r."is_active"`,
	},
	{
		name:        "bool with or",
		keelSchema:  testSchema,
		field:       "isExpensive Boolean @computed(item.price > 100 || item.isActive)",
		expectedSql: `(r."price" > 100 OR r."is_active")`,
	},
	{
		name:        "negation",
		keelSchema:  testSchema,
		field:       "isExpensive Boolean @computed(item.price > 100 || !item.isActive)",
		expectedSql: `(r."price" > 100 OR NOT r."is_active")`,
	},
	{
		name:        "negation and parenthesis",
		keelSchema:  testSchema,
		field:       "isExpensive Boolean @computed(!(item.price > 100 || item.isActive))",
		expectedSql: `NOT (r."price" > 100 OR r."is_active")`,
	},
	{
		name:        "1:M relationship",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.product.standardPrice * item.quantity)",
		expectedSql: `(SELECT "product"."standard_price" FROM "product" WHERE "product"."id" IS NOT DISTINCT FROM r."product_id") * r."quantity"`,
	},
	{
		name:        "nested 1:M relationship",
		keelSchema:  testSchema,
		field:       "total Decimal @computed(item.product.standardPrice * item.quantity + item.product.agent.commission)",
		expectedSql: `(SELECT "product"."standard_price" FROM "product" WHERE "product"."id" IS NOT DISTINCT FROM r."product_id") * r."quantity" + (SELECT "product$agent"."commission" FROM "product" LEFT JOIN "agent" AS "product$agent" ON "product$agent"."id" = "product"."agent_id" WHERE "product"."id" IS NOT DISTINCT FROM r."product_id")`,
	},
}

func TestGeneratedComputed(t *testing.T) {
	t.Parallel()
	for _, testCase := range computedTestCases {
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
			fieldName := strings.Split(testCase.field, " ")[0]
			field := proto.FindField(schema.Models, model.Name, fieldName)

			expression, err := parser.ParseExpression(field.ComputedExpression.Source)
			assert.NoError(t, err)

			sql, err := resolve.RunCelVisitor(expression, actions.GenerateComputedFunction(schema, model, field))
			assert.NoError(t, err)

			assert.Equal(t, testCase.expectedSql, sql, "expected `%s` but got `%s`", testCase.expectedSql, sql)
		})
	}
}
