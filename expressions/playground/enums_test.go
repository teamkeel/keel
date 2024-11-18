package playground

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/stretchr/testify/require"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func TestValid(t *testing.T) {
	typeProvider := NewTypeProvider()

	env, err := cel.NewCustomEnv(
		cel.ClearMacros(),
		cel.CustomTypeProvider(typeProvider),
		cel.Function(operators.Equals,
			cel.Overload(overloads.Equals, []*types.Type{types.NewOpaqueType("Status"), types.NewOpaqueType("Status")}, types.BoolType, decls.OverloadIsNonStrict()),
		),
		cel.Variable("person", types.NewObjectType("Person")),
		cel.Constant("Status", types.NewObjectType("StatusEnumDefinition"), nil),
		cel.EagerlyValidateDeclarations(true))
	require.NoError(t, err)

	_, issues := env.Compile("person.status == Status")
	require.Len(t, issues.Errors(), 0)
}

type typeProvider struct{}

var _ types.Provider = new(typeProvider)

func NewTypeProvider() *typeProvider {
	return &typeProvider{}
}

func (p *typeProvider) FindStructType(structType string) (*types.Type, bool) {
	if structType == "StatusEnumDefinition" {
		return types.NewObjectType("StatusEnumDefinition"), true
	}

	if structType == "Person" {
		return types.NewObjectType("Person"), true
	}

	return nil, false
}

func (p *typeProvider) FindStructFieldType(structType, fieldName string) (*types.FieldType, bool) {
	if structType == "StatusEnumDefinition" {
		switch fieldName {
		case "Married":
			return &types.FieldType{Type: types.NewOpaqueType("Status")}, true
		case "Single":
			return &types.FieldType{Type: types.NewOpaqueType("Status")}, true
		default:
			return nil, false
		}
	}

	if structType == "Person" {
		switch fieldName {
		case "status":
			return &types.FieldType{Type: types.NewOpaqueType("Status")}, true
		default:
			return nil, false
		}
	}

	return nil, false
}

func (p *typeProvider) EnumValue(enumName string) ref.Val {
	return types.NewErr("unknown enum name '%s'", enumName)
}

func (p *typeProvider) FindIdent(identName string) (ref.Val, bool) {
	return nil, false
}

func (p *typeProvider) FindStructFieldNames(structType string) ([]string, bool) {
	panic("not implemented")
}

func (p *typeProvider) FindType(typeName string) (*expr.Type, bool) {
	panic("not implemented")
}

func (p *typeProvider) FindFieldType(messageType string, fieldName string) (*types.FieldType, bool) {
	panic("not implemented")
}

func (p *typeProvider) NewValue(typeName string, fields map[string]ref.Val) ref.Val {
	panic("not implemented")
}
