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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'string' but it is 'int'", issues[0])
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'list(string)' but it is 'string'", issues[0])
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestDefault_InvalidNumber(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
	model Person {
		fields {
			age Number @default(1.5)
		}
	}`})

	model := query.Model(schema, "Person")
	field := query.Field(model, "age")
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'int' but it is 'double'", issues[0])
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'string' but it is 'int'", issues[0])
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'bool' but it is 'int'", issues[0])
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to '_==_' (in container '')", issues[0])
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to 'ctx' (in container '')", issues[0])
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
	defaultAttr := field.Attributes[0]

	expression := defaultAttr.Arguments[0].Expression.String()

	parser, err := attributes.NewDefaultExpressionParser(schema, field)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to '_+_' (in container '')", issues[0])
}
