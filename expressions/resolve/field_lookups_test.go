package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

var model = &parser.ModelNode{Name: parser.NameNode{Value: "Product"}}

func TestFieldLookups_ByModel(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, "product == someProduct")
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Equal(t, "product.id", strings.Join(lookups[0], "."))
}

func TestFieldLookups_ById(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, "product.id == someId")
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Equal(t, "product.id", strings.Join(lookups[0], "."))
}

func TestFieldLookups_Comparison(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, "product.rating > 3")
	assert.NoError(t, err)

	assert.Len(t, lookups, 0)
}

func TestFieldLookups_Variables(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, "product.sku == mySku")
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Equal(t, "product.sku", strings.Join(lookups[0], "."))
}

func TestFieldLookups_VariablesInverse(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, "mySku == product.sku")
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Equal(t, "product.sku", strings.Join(lookups[0], "."))
}

func TestFieldLookups_NotEquals(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, "product.sku != mySku")
	assert.NoError(t, err)

	assert.Len(t, lookups, 0)
}

func TestFieldLookups_WithAnd(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, `product.sku == mySku && product.name == "test"`)
	assert.NoError(t, err)

	assert.Len(t, lookups, 2)
	assert.Equal(t, "product.sku", strings.Join(lookups[0], "."))
	assert.Equal(t, "product.name", strings.Join(lookups[1], "."))
}

func TestFieldLookups_WithComparison(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, `product.sku == mySku && product.rating > 3`)
	assert.NoError(t, err)

	assert.Len(t, lookups, 1)
	assert.Equal(t, "product.sku", strings.Join(lookups[0], "."))
}

func TestFieldLookups_WithOr(t *testing.T) {
	lookups, err := resolve.FieldLookups(model, `product.sku == mySku || product.name == "test"`)
	assert.NoError(t, err)

	assert.Len(t, lookups, 0)
}
