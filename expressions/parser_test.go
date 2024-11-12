package orderby_expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
)

func TestValid(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			name Text
		}
	}`})

	expression := `person.name == 'Keel'`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func parse(t *testing.T, s *reader.SchemaFile) []*parser.AST {
	schema, err := parser.Parse(s)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return []*parser.AST{schema}
}
