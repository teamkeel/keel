package orderby_expression

import (
	"github.com/google/cel-go/common/types"
	"github.com/teamkeel/keel/schema/parser"
)

func mapType(t *parser.FieldNode) *types.Type {
	switch t.Type.Value {
	case parser.FieldTypeText:
		return types.StringType
	case parser.FieldTypeNumber:
		return types.IntType
	case parser.FieldTypeBoolean:
		return types.BoolType
	default:
		// Model type
		return types.NewObjectType(t.Type.Value)
	}
}
