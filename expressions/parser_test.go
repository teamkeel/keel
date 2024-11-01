package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

func TestSQLGen(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, err := ToSQL(schema, schema.FindModel("Person"), `person.name == "Keelson"`)
	require.NoError(t, err)
	require.Equal(t, `'person'.'name' = "Keelson"`, sql)
}

func TestValid(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	issues, err := Validate(schema, schema.FindModel("Person"), `person.name == "Keelson"`, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Empty(t, issues)
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

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	issues, err := Validate(schema, schema.FindModel("Person"), `person.org.companyName == "Keel"`, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
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

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	issues, err := Validate(schema, schema.FindModel("Person"), `person.org.companyName`, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
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

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	issues, err := Validate(schema, schema.FindModel("Person"), `person.firstName == "Keelson"`, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undefined field 'firstName'", issues[0])
}

func TestRemovedFunction(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
		}
	}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	issues, err := Validate(schema, schema.FindModel("Person"), `person.name.startsWith("Keelson")`, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to 'startsWith' (in container '')", issues[0])
}

func TestInvalidOperandTypes(t *testing.T) {
	var keelSchema = `
	model Person {
		fields {
			name Text
			age Number
		}
	}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	issues, err := Validate(schema, schema.FindModel("Person"), `person.name + person.age`, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
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

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	issues, err := Validate(schema, schema.FindModel("Person"), `person.age + 10`, &proto.TypeInfo{Type: proto.Type_TYPE_BOOL})
	require.NoError(t, err)
	require.Len(t, issues, 0)
}
