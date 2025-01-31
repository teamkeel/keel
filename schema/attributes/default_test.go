package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

func TestDefault_ValidString(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			name Text @default("Keelson")
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "name")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestDefault_InvalidString(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			name Text @default(1)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "name")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)

	require.Equal(t, "expression expected to resolve to type Text but it is Number", issues[0].Message)
}

func TestDefault_ValidStringArray(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			names Text[] @default(["Keelson", "Weave"])
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "names")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestDefault_InValidStringArray(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			names Text[] @default("Keelson")
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "names")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Text[] but it is Text", issues[0].Message)
}

func TestDefault_ValidNumber(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			age Number @default(123)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "age")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestDefault_ValidNumberFromDecimal(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			age Number @default(1.5)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "age")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 0)
}

func TestDefault_ValidDecimalFromNumber(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			age Decimal @default(1)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "age")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 0)
}

func TestDefault_ValidID(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			personId ID @default("123")
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "personId")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestDefault_InvalidID(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			personId ID @default(123)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "personId")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Text but it is Number", issues[0].Message)
}

func TestDefault_ValidBooleanb(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			isEmployed Boolean @default(false)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "isEmployed")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestDefault_InvalidBoolean(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			isEmployed Boolean @default(1)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "isEmployed")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Boolean but it is Number", issues[0].Message)
}

func TestDefault_InvalidWithOperators(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			isEmployed Boolean @default(true == true)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "isEmployed")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "operator '==' not supported in this context", issues[0].Message)
}

func TestDefault_InvalidWithCtx(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			isEmployed Boolean @default(ctx.isAuthenticated)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "isEmployed")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "unknown identifier 'ctx'", issues[0].Message)
}

func TestDefault_InvalidArithmetic(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			num Number @default(1 + 1)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "num")
	expression := field.Attributes[0].Arguments[0].Expression

	issues, err := attributes.ValidateDefaultExpression(schema, field, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "operator '+' not supported in this context", issues[0].Message)
}
