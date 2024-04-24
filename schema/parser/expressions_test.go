package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/schema/parser"
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
		"negative integer":       "a = -1",
		"decimal number":         "a = 1.580000", // %f uses a default precision of 6 digits after the decimal point
	}

	for name, fixture := range fixtures {
		t.Run(name, func(t *testing.T) {
			expr, err := parser.ParseExpression(fixture)
			assert.NoError(t, err)

			str, err := expr.ToString()
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

	expr, err := parser.ParseExpression(source)
	assert.NoError(t, err)

	output, err := expr.ToString()
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
		"1.12":    true,

		"a == b":          false,
		"true or a == b":  false,
		"true and a == b": false,
		"(a == b)":        false,
		"a = b":           false,
	}

	for input, expected := range fixtures {
		t.Run(input, func(t *testing.T) {
			expr, err := parser.ParseExpression(input)
			assert.NoError(t, err)

			assert.Equal(t, expected, expr.IsValue())
		})
	}
}

func TestIsAssignment(t *testing.T) {
	fixtures := map[string]bool{
		"a":       false,
		"1":       false,
		"-1":      false,
		"true":    false,
		"false":   false,
		"null":    false,
		"42":      false,
		"[1,2,3]": false,
		"1.23":    false,
		"-1.23":   false,

		"a == b":          false,
		"true or a == b":  false,
		"true and a == b": false,
		"(a == b)":        false,
		"a = b":           true,
		"a = -1":          true,
		"a += 1":          false,
		"a -= 1":          false,
		"a = 1.23":        true,
		"a = -1.23":       true,
	}

	for input, expected := range fixtures {
		t.Run(input, func(t *testing.T) {
			expr, err := parser.ParseExpression(input)
			assert.NoError(t, err)

			actual := expr.IsAssignment()
			assert.Equal(t, expected, actual)
		})
	}
}

func TestLogicalExpressions(t *testing.T) {
	fixtures := map[string]bool{
		"a":       false,
		"1":       false,
		"true":    false,
		"false":   false,
		"null":    false,
		"42":      false,
		"[1,2,3]": false,
		"a = b":   false,

		"a == b":                   true,
		"a.b.c == b":               true,
		"a == b.c":                 true,
		"a.b == b.c or a.c == b.c": true,
		"a.b == b.c and b.c == x.x or a.b.c == b.c.a": true,
		"a > b":              true,
		"a >= b":             true,
		"a < b":              true,
		"a <= b":             true,
		"a in b":             true,
		"a != b":             true,
		"a not in b":         true,
		"a.b.c not in b.c.a": true,
	}

	for input, expected := range fixtures {
		t.Run(input, func(t *testing.T) {
			expr, err := parser.ParseExpression(input)
			assert.NoError(t, err)

			for _, cond := range expr.Conditions() {
				if expected {
					assert.Equal(t, parser.LogicalCondition, cond.Type())
				} else {
					assert.NotEqual(t, parser.LogicalCondition, cond.Type())
				}
			}
		})
	}
}
