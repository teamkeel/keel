package expressions

import (
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/teamkeel/keel/proto"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// typeProvider supplies the CEL context with the relevant Keel types and identifiers
type typeProvider struct {
	schema *proto.Schema
	types.Provider
}

var _ types.Provider = new(typeProvider)

func NewTypeProvider(schema *proto.Schema) *typeProvider {
	return &typeProvider{schema: schema}
}

func (p *typeProvider) EnumValue(enumName string) ref.Val {
	return types.NewErr("unknown enum name '%s'", enumName)
}

func (p *typeProvider) FindIdent(identName string) (ref.Val, bool) {
	return nil, false
}

func (p *typeProvider) FindType(typeName string) (*expr.Type, bool) {
	return decls.NewTypeType(decls.NewObjectType(typeName)), true
}

func (p *typeProvider) FindStructType(structType string) (*types.Type, bool) {

	switch {
	case p.schema.FindModel(structType) != nil:
		return types.NewObjectType(structType), true
	case structType == "Context":
		return types.NewObjectType(structType), true
	}

	return nil, false
}

func (p *typeProvider) FindStructFieldNames(structType string) ([]string, bool) {
	panic("not implemented")
}

func (p *typeProvider) FindStructFieldType(structType, fieldName string) (*types.FieldType, bool) {
	if structType == "Context" {
		switch fieldName {
		case "identity":
			return &types.FieldType{Type: types.StringType}, true
		case "isAuthenticated":
			return &types.FieldType{Type: types.BoolType}, true
		case "headers":
			return &types.FieldType{Type: types.NewMapType(types.StringType, types.StringType)}, true
		}
	}

	if model := p.schema.FindModel(structType); model == nil {
		return nil, false
	}

	field := proto.FindField(p.schema.Models, structType, fieldName)
	if field == nil {
		return nil, false
	}

	t := fromKeel(field.Type)

	if field.Optional {
		t = types.NewNullableType(t)
	}

	return &types.FieldType{Type: t}, true
}

func (p *typeProvider) FindFieldType(messageType string, fieldName string) (*types.FieldType, bool) {

	return &types.FieldType{
		Type: types.StringType,
	}, true
}

func (p *typeProvider) NewValue(typeName string, fields map[string]ref.Val) ref.Val {
	return types.NewErr("unknown type '%s'", typeName)
}
