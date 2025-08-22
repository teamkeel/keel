package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

var model = &parser.ModelNode{EntityNode: parser.EntityNode{Name: parser.NameNode{Value: "Product"}}}

func TestFieldLookups_ByModel(t *testing.T) {
	expression, err := parser.ParseExpression("product == someProduct")
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Len(t, lookups[0], 1)
	assert.Equal(t, "product.id", lookups[0][0].String())
}

func TestFieldLookups_ById(t *testing.T) {
	expression, err := parser.ParseExpression("product.id == someId")
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Len(t, lookups[0], 1)
	assert.Equal(t, "product.id", lookups[0][0].String())
}

func TestFieldLookups_Comparison(t *testing.T) {
	expression, err := parser.ParseExpression("product.rating > 3")
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 0)
}

func TestFieldLookups_Variables(t *testing.T) {
	expression, err := parser.ParseExpression("product.sku == mySku")
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Len(t, lookups[0], 1)
	assert.Equal(t, "product.sku", lookups[0][0].String())
}

func TestFieldLookups_VariablesInverse(t *testing.T) {
	expression, err := parser.ParseExpression("mySku == product.sku")
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Len(t, lookups[0], 1)
	assert.Equal(t, "product.sku", lookups[0][0].String())
}

func TestFieldLookups_NotEquals(t *testing.T) {
	expression, err := parser.ParseExpression("product.sku != mySku")
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 0)
}

func TestFieldLookups_WithAnd(t *testing.T) {
	expression, err := parser.ParseExpression(`product.sku == mySku && product.name == "test"`)
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Len(t, lookups[0], 2)
	assert.Equal(t, "product.sku", lookups[0][0].String())
	assert.Equal(t, "product.name", lookups[0][1].String())
}

func TestFieldLookups_WithComparison(t *testing.T) {
	expression, err := parser.ParseExpression(`product.sku == mySku && product.rating > 3`)
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Len(t, lookups[0], 1)
	assert.Equal(t, "product.sku", lookups[0][0].String())
}

func TestFieldLookups_WithOr(t *testing.T) {
	expression, err := parser.ParseExpression(`product.sku == mySku || product.name == "test"`)
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 2)
	assert.Len(t, lookups[0], 1)
	assert.Len(t, lookups[1], 1)
	assert.Equal(t, "product.sku", lookups[0][0].String())
	assert.Equal(t, "product.name", lookups[1][0].String())
}

func TestFieldLookups_Complex(t *testing.T) {
	expression, err := parser.ParseExpression(`product.sku == mySku || product.name == "test" && product.id == "123"`)
	assert.NoError(t, err)

	lookups, err := resolve.FieldLookups(model, expression)
	assert.NoError(t, err)

	assert.Len(t, lookups, 2)
	assert.Len(t, lookups[0], 1)
	assert.Len(t, lookups[1], 2)
	assert.Equal(t, "product.sku", lookups[0][0].String())
	assert.Equal(t, "product.name", lookups[1][0].String())
	assert.Equal(t, "product.id", lookups[1][1].String())
}
