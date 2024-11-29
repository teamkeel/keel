package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
)

func TestString_Valid(t *testing.T) {
	s, err := resolve.AsString(`"keel"`)
	assert.NoError(t, err)

	assert.Equal(t, "keel", s)
}
