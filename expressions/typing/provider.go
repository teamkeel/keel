package typing

import (
	"strings"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// TypeProvider supplies the CEL context with the relevant Keel types and identifiers
type TypeProvider struct {
	Schema  []*parser.AST
	Model   string
	Objects map[string]map[string]*types.Type
}

var _ types.Provider = new(TypeProvider)

func NewTypeProvider() *TypeProvider {
	return &TypeProvider{
		Objects: map[string]map[string]*types.Type{},
	}
}
func (p *TypeProvider) EnumValue(enumName string) ref.Val {
	return types.NewErr("unknown enum name '%s'", enumName)
}

func (p *TypeProvider) FindIdent(identName string) (ref.Val, bool) {
	return nil, false
}

func (p *TypeProvider) FindType(typeName string) (*expr.Type, bool) {
	panic("not implemented")
}

func (p *TypeProvider) FindStructType(structType string) (*types.Type, bool) {
	obj := strings.TrimSuffix(structType, "[]")

	switch {
	case query.Model(p.Schema, obj) != nil:
		return types.NewObjectType(structType), true
	case strings.Contains(obj, "_Enum") && query.Enum(p.Schema, strings.TrimSuffix(obj, "_Enum")) != nil:
		return types.NewObjectType(structType), true
	case structType == "Context":
		return types.NewObjectType(structType), true

	}

	return nil, false
}

func (p *TypeProvider) FindStructFieldNames(structType string) ([]string, bool) {
	panic("not implemented")
}

func (p *TypeProvider) FindStructFieldType(structType, fieldName string) (*types.FieldType, bool) {
	obj := strings.TrimSuffix(structType, "[]")

	parentIsArray := strings.HasSuffix(structType, "[]")

	if model := query.Model(p.Schema, obj); model != nil {
		field := query.Field(model, fieldName)
		if field == nil {
			return nil, false
		}

		t, err := MapType(p.Schema, field.Type.Value, field.Repeated || parentIsArray)
		if err != nil {
			return nil, false
		}

		return &types.FieldType{Type: t}, true
	}

	if strings.Contains(structType, "_Enum") {
		e := strings.TrimSuffix(structType, "_Enum")
		if enum := query.Enum(p.Schema, e); enum != nil {
			for _, v := range enum.Values {
				if v.Name.Value == fieldName {
					return &types.FieldType{Type: types.NewOpaqueType(e)}, true
				}
			}
		}
	}

	if field, has := p.Objects[structType][fieldName]; has {
		return &types.FieldType{Type: field}, true
	}

	return nil, false
}

func (p *TypeProvider) FindFieldType(messageType string, fieldName string) (*types.FieldType, bool) {
	panic("not implemented")
}

func (p *TypeProvider) NewValue(typeName string, fields map[string]ref.Val) ref.Val {
	panic("not implemented")
}
