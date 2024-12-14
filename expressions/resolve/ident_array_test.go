package resolve_test

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
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
	assert.Equal(t, "one", strings.Join(operands[0].Fragments, "."))
	assert.Equal(t, "two", strings.Join(operands[1].Fragments, "."))
	assert.Equal(t, lexer.Position{Offset: 1, Column: 1, Line: 1}, operands[0].Pos)
	assert.Equal(t, lexer.Position{Offset: 4, Column: 4, Line: 1}, operands[0].EndPos)
	assert.Equal(t, lexer.Position{Offset: 5, Column: 5, Line: 1}, operands[1].Pos)
	assert.Equal(t, lexer.Position{Offset: 8, Column: 8, Line: 1}, operands[1].EndPos)

}

func TestIdentArray_Enums(t *testing.T) {
	expression, err := parser.ParseExpression("[MyEnum.One, MyEnum.Two]")
	assert.NoError(t, err)

	operands, err := resolve.AsIdentArray(expression)
	assert.NoError(t, err)

	assert.Len(t, operands, 2)
	assert.Equal(t, "MyEnum.One", strings.Join(operands[0].Fragments, "."))
	assert.Equal(t, "MyEnum.Two", strings.Join(operands[1].Fragments, "."))
	assert.Equal(t, lexer.Position{Offset: 1, Column: 2, Line: 1}, operands[0].Pos)
	assert.Equal(t, lexer.Position{Offset: 11, Column: 12, Line: 1}, operands[0].EndPos)
	assert.Equal(t, lexer.Position{Offset: 13, Column: 14, Line: 1}, operands[1].Pos)
	assert.Equal(t, lexer.Position{Offset: 23, Column: 24, Line: 1}, operands[1].EndPos)
}

func TestIdentArray_Empty(t *testing.T) {
	expression, err := parser.ParseExpression("[]")
	assert.NoError(t, err)

	operands, err := resolve.AsIdentArray(expression)
	assert.NoError(t, err)

	assert.Len(t, operands, 0)
}
