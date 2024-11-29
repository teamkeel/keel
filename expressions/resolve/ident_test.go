package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
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
