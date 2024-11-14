package expressions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
)

func TestTextEquality_Valid(t *testing.T) {
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

func TestEnums_Valid(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status == Status.Married`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestEnums_InvalidValue(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status == Status.NotExists`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undefined field 'NotExists'", issues[0])
}

func TestEnums_NoValue(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status == Status`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_==_' applied to '(Status, Status_EnumDefinition)'", issues[0])
}

func TestEnums_WrongEnum(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
				employment Employment
			}
		}
		enum Status {
			Married
			Single
		}
		enum Employment {
			Permanent
			Temporary
			Unemployed
		}`})

	expression := `person.status == Employment.Permanent`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_==_' applied to '(Status, Employment)'", issues[0])
}

func parse(t *testing.T, s *reader.SchemaFile) []*parser.AST {
	schema, err := parser.Parse(s)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return []*parser.AST{schema}
}
