package expressions

import (
	"fmt"

	"github.com/google/cel-go/common/types"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func mapType(schema []*parser.AST, typeName string) (*types.Type, error) {
	// TODO: should we define our own types?  i.e. types.NewOpaqueType("ID")
	switch typeName {
	case parser.FieldTypeID:
		return types.StringType, nil
	case parser.FieldTypeText:
		return types.StringType, nil
	case parser.FieldTypeNumber:
		return types.IntType, nil
	case parser.FieldTypeDecimal:
		return types.DoubleType, nil
	case parser.FieldTypeBoolean:
		return types.BoolType, nil
	case parser.FieldTypeDatetime:
		return types.TimestampType, nil
	case parser.FieldTypeDate:
		return types.TimestampType, nil
	case parser.FieldTypeMarkdown:
		return types.StringType, nil
	default:

	}

	switch {
	case query.Enum(schema, typeName) != nil:
		return types.NewOpaqueType(typeName), nil
	case query.Model(schema, typeName) != nil:
		return types.NewObjectType(typeName), nil
	}

	return nil, fmt.Errorf("cannot map from type '%s'", typeName)
}
