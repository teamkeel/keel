package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/expressions/visitor"
)

func TestIdent_ModelField(t *testing.T) {
	ident, err := resolve.AsIdent("post.name")
	assert.NoError(t, err)

	assert.Equal(t, "post.name", strings.Join(ident, "."))
}

func TestIdent_Variable(t *testing.T) {
	ident, err := resolve.AsIdent("name")
	assert.NoError(t, err)

	assert.Equal(t, "name", strings.Join(ident, "."))
}

func TestIdent_Literal(t *testing.T) {
	_, err := resolve.AsIdent("123")
	assert.ErrorIs(t, err, resolve.ErrExpressionNotValidIdent)
}

func TestIdent_Operator(t *testing.T) {
	_, err := resolve.AsIdent("post.age + 1")
	assert.ErrorIs(t, err, resolve.ErrExpressionNotValidIdent)
}

func TestIdent_Empty(t *testing.T) {
	_, err := resolve.AsIdent("")
	assert.ErrorIs(t, err, visitor.ErrExpressionNotParseable)
}

func TestIdent_InvalidExpression(t *testing.T) {
	_, err := resolve.AsIdent("post.age,")
	assert.ErrorIs(t, err, visitor.ErrExpressionNotParseable)
}
