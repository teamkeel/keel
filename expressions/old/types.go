package expressions_old

import (
	"github.com/google/cel-go/common/types"
	"github.com/teamkeel/keel/proto"
)

func fromKeel(t *proto.TypeInfo) *types.Type {
	switch t.Type {
	case proto.Type_TYPE_STRING:
		return types.StringType
	case proto.Type_TYPE_INT:
		return types.IntType
	case proto.Type_TYPE_BOOL:
		return types.BoolType
	case proto.Type_TYPE_MODEL:
		return types.NewObjectType(t.ModelName.Value)
	}

	panic("not implemented")
}
