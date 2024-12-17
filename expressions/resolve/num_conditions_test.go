package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

func TestNumConditions_Field(t *testing.T) {
	expression, err := parser.ParseExpression("post.isActive")
	assert.NoError(t, err)
	count, err := resolve.NumConditions(expression)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestNumConditions_Arithmetic(t *testing.T) {
	expression, err := parser.ParseExpression("1 + 1")
	assert.NoError(t, err)
	count, err := resolve.NumConditions(expression)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestNumConditions_ArithmeticMultiple(t *testing.T) {
	expression, err := parser.ParseExpression("post.isActive or true")
	assert.NoError(t, err)
	count, err := resolve.NumConditions(expression)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestNumConditions_Complex(t *testing.T) {
	expression, err := parser.ParseExpression(`product.sku == mySku || product.name == "test" && (product.id == "123" || true)`)
	assert.NoError(t, err)
	count, err := resolve.NumConditions(expression)
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}
