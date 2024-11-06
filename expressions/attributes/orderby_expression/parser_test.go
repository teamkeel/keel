package orderby_expression

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
)

func TestValid(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	expression := `person.name == "Keelson"`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewOrderByExpressionParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}
