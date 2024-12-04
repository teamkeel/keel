package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
)

func TestDefinitiveIdentLookups_Variables(t *testing.T) {
	lookups, err := resolve.DefinitiveLookups("[one,two]")
	assert.NoError(t, err)

	assert.Len(t, lookups, 2)
	assert.Equal(t, "one", strings.Join(lookups[0], "."))
	assert.Equal(t, "two", strings.Join(lookups[1], "."))
}
