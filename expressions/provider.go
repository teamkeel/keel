package orderby_expression

import (
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// typeProvider supplies the CEL context with the relevant Keel types and identifiers
type typeProvider struct {
	asts    []*parser.AST
	model   string
	context map[string]map[string]*types.Type
	// Person -> name -> Text
	//        -> age  -> Number

}

var _ types.Provider = new(typeProvider)

func NewTypeProvider() *typeProvider {
	return &typeProvider{
		context: map[string]map[string]*types.Type{},
	}
}

func (p *typeProvider) EnumValue(enumName string) ref.Val {
	return types.NewErr("unknown enum name '%s'", enumName)
}

func (p *typeProvider) FindIdent(identName string) (ref.Val, bool) {
	return nil, false
}

func (p *typeProvider) FindType(typeName string) (*expr.Type, bool) {
	return decls.NewTypeType(decls.NewObjectType(strcase.ToCamel(typeName))), true
}

func (p *typeProvider) FindStructType(structType string) (*types.Type, bool) {

	switch {

	case query.Model(p.asts, strcase.ToCamel(structType)) != nil:
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

	if model := query.Model(p.asts, strcase.ToCamel(structType)); model != nil {
		field := query.Field(model, fieldName)
		if field == nil {
			return nil, false
		}

		t := mapType(field)

		if field.Optional {
			t = types.NewNullableType(t)
		}

		return &types.FieldType{Type: t}, true
	}

	return nil, false

}

func (p *typeProvider) FindFieldType(messageType string, fieldName string) (*types.FieldType, bool) {

	return &types.FieldType{
		Type: types.StringType,
	}, true
}

func (p *typeProvider) NewValue(typeName string, fields map[string]ref.Val) ref.Val {
	return types.NewErr("unknown type '%s'", typeName)
}
