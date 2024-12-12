package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
)

func TestToValue_String(t *testing.T) {
	v, err := resolve.ToValue[string](`"keel"`)
	assert.NoError(t, err)
	assert.Equal(t, "keel", v)
}

func TestToValue_NotString(t *testing.T) {
	_, err := resolve.ToValue[string](`1`)
	assert.ErrorContains(t, err, "value is of type 'int64' and cannot assert type 'string'")
}

func TestToValue_Number(t *testing.T) {
	v, err := resolve.ToValue[int64](`1 + 1`)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), v)
}

func TestToValue_Float(t *testing.T) {
	v, err := resolve.ToValue[float64](`1.5 + 1.1`)
	assert.NoError(t, err)
	assert.Equal(t, float64(2.6), v)
}

func TestToValue_Boolean(t *testing.T) {
	v, err := resolve.ToValue[bool](`true`)
	assert.NoError(t, err)
	assert.Equal(t, true, v)
}

func TestToValueArray_StringArray(t *testing.T) {
	v, err := resolve.ToValueArray[string](`["keel", "weave"]`)
	assert.NoError(t, err)
	assert.Equal(t, "keel", v[0])
	assert.Equal(t, "weave", v[1])
}
