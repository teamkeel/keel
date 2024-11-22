package expressions

import (
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// typeProvider supplies the CEL context with the relevant Keel types and identifiers
type typeProvider struct {
	schema  []*parser.AST
	model   string
	objects map[string]map[string]*types.Type
}

var _ types.Provider = new(typeProvider)

func NewTypeProvider() *typeProvider {
	return &typeProvider{
		objects: map[string]map[string]*types.Type{},
	}
}

func (p *typeProvider) RegisterDescriptor(protoreflect.FileDescriptor) error {
	panic("not implemented")
}

func (p *typeProvider) RegisterType(types ...ref.Type) error {

	//panic("not implemented")
	return nil
}

func (p *typeProvider) EnumValue(enumName string) ref.Val {
	return types.NewErr("unknown enum name '%s'", enumName)
}

func (p *typeProvider) FindIdent(identName string) (ref.Val, bool) {
	return nil, false
}

func (p *typeProvider) FindType(typeName string) (*expr.Type, bool) {
	panic("not implemented")
	return decls.NewTypeType(decls.NewObjectType(strcase.ToCamel(typeName))), true
}

func (p *typeProvider) FindStructType(structType string) (*types.Type, bool) {
	obj := strings.TrimSuffix(structType, "[]")

	switch {
	case query.Model(p.schema, obj) != nil:
		return types.NewObjectType(structType), true
	case strings.Contains(obj, "_EnumDefinition") && query.Enum(p.schema, strings.TrimSuffix(obj, "_EnumDefinition")) != nil:
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
	obj := strings.TrimSuffix(structType, "[]")

	parentIsArray := strings.HasSuffix(structType, "[]")

	if model := query.Model(p.schema, obj); model != nil {
		field := query.Field(model, fieldName)
		if field == nil {
			return nil, false
		}

		t, err := mapType(p.schema, field.Type.Value)
		if err != nil {
			return nil, false
		}

		if field.Optional {
			// only works with primitives
			t = types.NewNullableType(t)
		}

		if field.Repeated || parentIsArray {
			if query.Model(p.schema, field.Type.Value) != nil {
				t = cel.ObjectType(field.Type.Value + "[]")
			} else {
				t = cel.ListType(t)
			}
		}

		return &types.FieldType{Type: t}, true
	}

	if strings.Contains(structType, "_EnumDefinition") {
		e := strings.TrimSuffix(structType, "_EnumDefinition")
		if enum := query.Enum(p.schema, e); enum != nil {
			for _, v := range enum.Values {
				if v.Name.Value == fieldName {
					return &types.FieldType{Type: types.NewOpaqueType(e)}, true
				}
			}
		}
	}

	if field, has := p.objects[structType][fieldName]; has {
		return &types.FieldType{Type: field}, true
	}

	return nil, false
}

func (p *typeProvider) FindFieldType(messageType string, fieldName string) (*types.FieldType, bool) {
	panic("not implemented")
	return &types.FieldType{
		Type: types.StringType,
	}, true
}

func (p *typeProvider) NewValue(typeName string, fields map[string]ref.Val) ref.Val {
	panic("not implemented")
	return types.NewErr("unknown type '%s'", typeName)
}
