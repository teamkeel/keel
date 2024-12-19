package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

func TestIdent_ModelField(t *testing.T) {
	expression, err := parser.ParseExpression("post.name")
	assert.NoError(t, err)

	ident, err := resolve.AsIdent(expression)
	assert.NoError(t, err)

	assert.Equal(t, "post.name", ident.String())
}

func TestIdent_Variable(t *testing.T) {
	expression, err := parser.ParseExpression("name")
	assert.NoError(t, err)

	ident, err := resolve.AsIdent(expression)
	assert.NoError(t, err)

	assert.Equal(t, "name", ident.String())
}

func TestIdent_Literal(t *testing.T) {
	expression, err := parser.ParseExpression("123")
	assert.NoError(t, err)

	_, err = resolve.AsIdent(expression)
	assert.ErrorIs(t, err, resolve.ErrExpressionNotValidIdent)
}

func TestIdent_Operator(t *testing.T) {
	expression, err := parser.ParseExpression("post.age + 1")
	assert.NoError(t, err)

	_, err = resolve.AsIdent(expression)
	assert.ErrorIs(t, err, resolve.ErrExpressionNotValidIdent)
}

func TestIdent_Empty(t *testing.T) {
	expression, err := parser.ParseExpression("")
	assert.NoError(t, err)

	_, err = resolve.AsIdent(expression)
	assert.ErrorIs(t, err, resolve.ErrExpressionNotParseable)
}
