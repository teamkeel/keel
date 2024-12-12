package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
)

func TestIdentArray_Variables(t *testing.T) {
	operands, err := resolve.AsIdentArray("[one,two]")
	assert.NoError(t, err)

	assert.Len(t, operands, 2)
	assert.Equal(t, "one", strings.Join(operands[0], "."))
	assert.Equal(t, "two", strings.Join(operands[1], "."))
}

func TestIdentArray_Enums(t *testing.T) {
	operands, err := resolve.AsIdentArray("[MyEnum.One, MyEnum.Two]")
	assert.NoError(t, err)

	assert.Len(t, operands, 2)
	assert.Equal(t, "MyEnum.One", strings.Join(operands[0], "."))
	assert.Equal(t, "MyEnum.Two", strings.Join(operands[1], "."))
}

func TestIdentArray_Empty(t *testing.T) {
	operands, err := resolve.AsIdentArray("[]")
	assert.NoError(t, err)

	assert.Len(t, operands, 0)
}
