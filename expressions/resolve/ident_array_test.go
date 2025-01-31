package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

func TestIdentArray_Variables(t *testing.T) {
	expression, err := parser.ParseExpression("[one,two]")
	assert.NoError(t, err)

	operands, err := resolve.AsIdentArray(expression)
	assert.NoError(t, err)

	assert.Len(t, operands, 2)
	assert.Equal(t, "one", operands[0].String())
	assert.Equal(t, "two", operands[1].String())
}

func TestIdentArray_Enums(t *testing.T) {
	expression, err := parser.ParseExpression("[MyEnum.One, MyEnum.Two]")
	assert.NoError(t, err)

	operands, err := resolve.AsIdentArray(expression)
	assert.NoError(t, err)

	assert.Len(t, operands, 2)
	assert.Equal(t, "MyEnum.One", operands[0].String())
	assert.Equal(t, "MyEnum.Two", operands[1].String())
}

func TestIdentArray_Empty(t *testing.T) {
	expression, err := parser.ParseExpression("[]")
	assert.NoError(t, err)

	operands, err := resolve.AsIdentArray(expression)
	assert.NoError(t, err)

	assert.Len(t, operands, 0)
}
