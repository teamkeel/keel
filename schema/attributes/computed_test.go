package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

func TestComputed_SumFunction(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Invoice {
			fields {
				items Item[]
				total Decimal @computed(SUM(invoice.items.total))
			}
		}
		model Item {
			fields {
				invoice Invoice
				product Product
				quantity Number
				total Decimal? @computed(item.quantity * item.product.price)
			}
		}
		model Product {
			fields {
				price Decimal
				items Item[]
			}
		}`})

	model := query.Model(schema, "Invoice")
	expression := model.Sections[0].Fields[1].Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateComputedExpression(schema, model, model.Sections[0].Fields[1], expression)
	require.NoError(t, err)
	require.Len(t, issues, 0)
}

func TestComputed_AggregatedSumFunction(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Invoice {
			fields {
				items Item[]
				total Decimal @computed(SUM(invoice.items.total, item.isDeleted == false))
			}
		}
		model Item {
			fields {
				invoice Invoice
				product Product
				quantity Number
				total Decimal? @computed(item.quantity * item.product.price)
				isDeleted Boolean @default(false)
			}
		}
		model Product {
			fields {
				price Decimal
				items Item[]
			}
		}`})

	model := query.Model(schema, "Invoice")
	expression := model.Sections[0].Fields[1].Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateComputedExpression(schema, model, model.Sections[0].Fields[1], expression)
	require.NoError(t, err)
	require.Len(t, issues, 0)
}
