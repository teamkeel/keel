package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

func TestOperands_ModelField(t *testing.T) {
	expression, err := parser.ParseExpression("post.isActive == true")
	assert.NoError(t, err)

	ident, err := resolve.IdentOperands(expression)
	assert.NoError(t, err)

	assert.Len(t, ident, 1)
	assert.Equal(t, "post.isActive", strings.Join(ident[0].Fragments, "."))
}

func TestOperands_Variable(t *testing.T) {
	expression, err := parser.ParseExpression("isActive == true")
	assert.NoError(t, err)

	ident, err := resolve.IdentOperands(expression)
	assert.NoError(t, err)

	assert.Len(t, ident, 1)
	assert.Equal(t, "isActive", strings.Join(ident[0].Fragments, "."))
}

func TestOperands_Complex(t *testing.T) {
	expression, err := parser.ParseExpression(`isPublic == true || (post.hasAdminAccess == true && ctx.identity.user.isAdmin)`)
	assert.NoError(t, err)

	ident, err := resolve.IdentOperands(expression)
	assert.NoError(t, err)

	assert.Len(t, ident, 3)
	assert.Equal(t, "isPublic", strings.Join(ident[0].Fragments, "."))
	assert.Equal(t, "post.hasAdminAccess", strings.Join(ident[1].Fragments, "."))
	assert.Equal(t, "ctx.identity.user.isAdmin", strings.Join(ident[2].Fragments, "."))
}
