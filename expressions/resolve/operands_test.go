package resolve_test

import (
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
	assert.Equal(t, "post.isActive", ident[0].String())
}

func TestOperands_Variable(t *testing.T) {
	expression, err := parser.ParseExpression("isActive == true")
	assert.NoError(t, err)

	ident, err := resolve.IdentOperands(expression)
	assert.NoError(t, err)

	assert.Len(t, ident, 1)
	assert.Equal(t, "isActive", ident[0].String())
}

func TestOperands_Complex(t *testing.T) {
	expression, err := parser.ParseExpression(`isPublic == true || (post.hasAdminAccess == true && ctx.identity.user.isAdmin)`)
	assert.NoError(t, err)

	ident, err := resolve.IdentOperands(expression)
	assert.NoError(t, err)

	assert.Len(t, ident, 3)
	assert.Equal(t, "isPublic", ident[0].String())
	assert.Equal(t, "post.hasAdminAccess", ident[1].String())
	assert.Equal(t, "ctx.identity.user.isAdmin", ident[2].String())
}

func TestOperands_Functions(t *testing.T) {
	expression, err := parser.ParseExpression(`isPublic == true && SUM(account.transactions.amount) > 100`)
	assert.NoError(t, err)

	ident, err := resolve.IdentOperands(expression)
	assert.NoError(t, err)

	assert.Len(t, ident, 2)
	assert.Equal(t, "isPublic", ident[0].String())
	assert.Equal(t, "account.transactions.amount", ident[1].String())
}

func TestOperands_FunctionsMultipleArgs(t *testing.T) {
	expression, err := parser.ParseExpression(`isPublic == true && SUMIF(account.transactions.amount, account.transactions.isDeleted == false && account.transactions.amount > 0) > 100`)
	assert.NoError(t, err)

	ident, err := resolve.IdentOperands(expression)
	assert.NoError(t, err)

	assert.Len(t, ident, 3)
	assert.Equal(t, "isPublic", ident[0].String())
	assert.Equal(t, "account.transactions.amount", ident[1].String())
	assert.Equal(t, "account.transactions.isDeleted", ident[2].String())
}
