package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
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

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestValidNullable(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text?
		}
	}`

	expression := `ctx.identity == null`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestInvalidNullable(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	expression := `ctx.identity == null`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_==_' applied to '(string, null)'", issues[0])
}

func TestValidRelationship(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
			org Organisation
		}
	}
	model Organisation {
		fields {
			companyName Text
		}
	}`

	expression := `person.org.companyName == "Keel"`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestInvalidOutputType(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
			org Organisation
		}
	}
	model Organisation {
		fields {
			companyName Text
		}
	}`

	expression := `person.org.companyName`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'TYPE_BOOL'", issues[0])
}

func TestInvalidField(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	expression := `person.firstName == "Keelson"`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undefined field 'firstName'", issues[0])
}

func TestInvalidOperandTypes(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
			age Number
		}
	}`

	expression := `person.name + person.age`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_+_' applied to '(string, int)'", issues[0])
}

func TestValidArithmetic(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
			age Number
		}
	}`

	expression := `person.age + 10`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Len(t, issues, 0)
}

func TestValidFunction(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	expression := `upper(person.name) == upper("Keelson")`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestValidFunctionWithExpression(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			firstName Text
			lastName Text
		}
	}`

	expression := `upper(person.firstName + " " + person.lastName) == "KEEL KEELSON"`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestInvalidFunctionArgumentType(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
			age Number
		}
	}`

	expression := `upper(person.age) == "40"`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for 'upper' applied to '(int)'", issues[0])
}

func TestInvalidFunction(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	expression := `person.name.startsWith("Keelson")`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	parser, err := NewParser(schema, schema.FindModel("Person"))
	require.NoError(t, err)

	issues, err := parser.Validate(expression, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to 'startsWith' (in container '')", issues[0])
}
