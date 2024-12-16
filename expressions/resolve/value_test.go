package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

func TestToValue_String(t *testing.T) {
	expression, err := parser.ParseExpression(`"keel"`)
	assert.NoError(t, err)

	v, _, err := resolve.ToValue[string](expression)
	assert.NoError(t, err)
	assert.Equal(t, "keel", v)
}

func TestToValue_NotString(t *testing.T) {
	expression, err := parser.ParseExpression(`1`)
	assert.NoError(t, err)

	_, _, err = resolve.ToValue[string](expression)
	assert.ErrorContains(t, err, "value is of type 'int64' and cannot assert type 'string'")
}

func TestToValue_Number(t *testing.T) {
	expression, err := parser.ParseExpression(`1 + 1`)
	assert.NoError(t, err)

	v, _, err := resolve.ToValue[int64](expression)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), v)
}

func TestToValue_Float(t *testing.T) {
	expression, err := parser.ParseExpression(`1.5 + 1.1`)
	assert.NoError(t, err)

	v, _, err := resolve.ToValue[float64](expression)
	assert.NoError(t, err)
	assert.Equal(t, float64(2.6), v)
}

func TestToValue_Boolean(t *testing.T) {
	expression, err := parser.ParseExpression(`true`)
	assert.NoError(t, err)

	v, _, err := resolve.ToValue[bool](expression)
	assert.NoError(t, err)
	assert.Equal(t, true, v)
}

func TestToValueArray_StringArray(t *testing.T) {
	expression, err := parser.ParseExpression(`["keel", "weave"]`)
	assert.NoError(t, err)

	v, err := resolve.ToValueArray[string](expression)
	assert.NoError(t, err)
	assert.Equal(t, "keel", v[0])
	assert.Equal(t, "weave", v[1])
}

func TestToValueArray_Null(t *testing.T) {
	expression, err := parser.ParseExpression(`null`)
	assert.NoError(t, err)

	v, isNull, err := resolve.ToValue[any](expression)
	assert.NoError(t, err)
	assert.True(t, isNull)
	assert.Equal(t, nil, v)
}
