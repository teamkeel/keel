package expressions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions"
)

func TestRoundTrip(t *testing.T) {
	fixtures := map[string]string{
		"single ident":           "a",
		"array of values":        "[a, 2, true, false, null, \"literal\"]",
		"equals":                 "a == b",
		"not equals":             "a != b",
		"greater than or equals": "a >= b",
		"less than or equals":    "a <= b",
		"greater than":           "a > b",
		"less than":              "a < b",
		"not in":                 "a not in b",
		"in":                     "a in b",
		"increment by":           "a += b",
		"decrement by":           "a -= b",
		"assignment":             "a = b",
		"or condition":           "a == b or a > c",
		"and condition":          "a == b and a > c",
		"mixed or/and":           "a == b or a < c and a > d",
		"parenthesis":            "(a == b or a < c) and a > d",
		"dot notation":           "a.b.c == d.e.f",
	}

	for name, fixture := range fixtures {
		t.Run(name, func(t *testing.T) {
			expr, err := expressions.Parse(fixture)
			assert.NoError(t, err)

			str, err := expressions.ToString(expr)
			assert.NoError(t, err)
			assert.Equal(t, fixture, str)
		})
	}
}

func TestToString(t *testing.T) {
	source := `
	a   ==   b   or
	(
		(c  <    d)  and
		(e  >    f)
	)
	`

	expr, err := expressions.Parse(source)
	assert.NoError(t, err)

	output, err := expressions.ToString(expr)
	assert.NoError(t, err)

	assert.Equal(t, "a == b or ((c < d) and (e > f))", output)
}

func TestIsValue(t *testing.T) {
	fixtures := map[string]bool{
		"a":       true,
		"1":       true,
		"true":    true,
		"false":   true,
		"null":    true,
		"42":      true,
		"[1,2,3]": true,

		"a == b":          false,
		"true or a == b":  false,
		"true and a == b": false,
		"(a == b)":        false,
		"a = b":           false,
	}

	for input, expected := range fixtures {
		t.Run(input, func(t *testing.T) {
			expr, err := expressions.Parse(input)
			assert.NoError(t, err)

			assert.Equal(t, expected, expressions.IsValue(expr))
		})
	}
}
