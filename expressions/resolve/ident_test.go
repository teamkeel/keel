package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/expressions/visitor"
	"github.com/teamkeel/keel/schema/parser"
)

func TestIdent_ModelField(t *testing.T) {
	expression, err := parser.ParseExpression("post.name")
	assert.NoError(t, err)

	ident, err := resolve.AsIdent(expression)
	assert.NoError(t, err)

	assert.Equal(t, "post.name", ident.ToString())
	assert.Equal(t, 1, ident.Pos.Column)
	assert.Equal(t, 1, ident.Pos.Line)
	assert.Equal(t, 0, ident.Pos.Offset)
	assert.Equal(t, 10, ident.EndPos.Column)
	assert.Equal(t, 1, ident.EndPos.Line)
	assert.Equal(t, 9, ident.EndPos.Offset)
}

func TestIdent_Variable(t *testing.T) {
	expression, err := parser.ParseExpression("name")
	assert.NoError(t, err)

	ident, err := resolve.AsIdent(expression)
	assert.NoError(t, err)

	assert.Equal(t, "name", ident.ToString())
	assert.Equal(t, 0, ident.Pos.Column)
	assert.Equal(t, 1, ident.Pos.Line)
	assert.Equal(t, 0, ident.Pos.Offset)
	assert.Equal(t, 4, ident.EndPos.Column)
	assert.Equal(t, 1, ident.EndPos.Line)
	assert.Equal(t, 4, ident.EndPos.Offset)
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
	assert.ErrorIs(t, err, visitor.ErrExpressionNotParseable)
}
