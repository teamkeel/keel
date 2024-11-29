package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
)

func TestOperands_ModelField(t *testing.T) {
	ident, err := resolve.IdentOperands("post.isActive == true")
	assert.NoError(t, err)

	assert.Len(t, ident, 1)
	assert.Equal(t, "post.isActive", strings.Join(ident[0], "."))
}

func TestOperands_Variable(t *testing.T) {
	ident, err := resolve.IdentOperands("isActive == true")
	assert.NoError(t, err)

	assert.Len(t, ident, 1)
	assert.Equal(t, "isActive", strings.Join(ident[0], "."))
}

func TestOperands_Complex(t *testing.T) {
	ident, err := resolve.IdentOperands(`isPublic == true || (post.hasAdminAccess == true && ctx.identity.user.isAdmin)`)
	assert.NoError(t, err)

	assert.Len(t, ident, 3)
	assert.Equal(t, "isPublic", strings.Join(ident[0], "."))
	assert.Equal(t, "post.hasAdminAccess", strings.Join(ident[1], "."))
	assert.Equal(t, "ctx.identity.user.isAdmin", strings.Join(ident[2], "."))
}
