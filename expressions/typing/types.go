package typing

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

type Ident []string

var (
	TypeID        = cel.OpaqueType(parser.FieldTypeID)
	TypeText      = cel.OpaqueType(parser.FieldTypeText)
	TypeMarkdown  = cel.OpaqueType(parser.FieldTypeMarkdown)
	TypeNumber    = cel.OpaqueType(parser.FieldTypeNumber)
	TypeDecimal   = cel.OpaqueType(parser.FieldTypeDecimal)
	TypeBoolean   = cel.OpaqueType(parser.FieldTypeBoolean)
	TypeTimestamp = cel.OpaqueType(parser.FieldTypeTimestamp)
	TypeDate      = cel.OpaqueType(parser.FieldTypeDate)
	TypeDuration  = cel.OpaqueType(parser.FieldTypeDuration)

	TypeIDArray        = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeID))
	TypeTextArray      = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeText))
	TypeMarkdownArray  = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeMarkdown))
	TypeNumberArray    = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeNumber))
	TypeDecimalArray   = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeDecimal))
	TypeBooleanArray   = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeBoolean))
	TypeTimestampArray = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeTimestamp))
	TypeDateArray      = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeDate))
	TypeDurationArray  = cel.OpaqueType(fmt.Sprintf("%s[]", parser.FieldTypeDuration))
)

var (
	TypeNameContext = "_Context"
	TypeNameHeaders = "_Headers"
	TypeNameSecrets = "_Secrets"
	TypeNameEnvvars = "_EnvironmentVariables"
)

var (
	TypeContext = types.NewObjectType(TypeNameContext)
	TypeHeaders = types.NewObjectType(TypeNameHeaders)
	TypeSecrets = types.NewObjectType(TypeNameSecrets)
	TypeEnvvars = types.NewObjectType(TypeNameEnvvars)
)

var (
	FunctionSum    = "SUM"
	FunctionCount  = "COUNT"
	FunctionAvg    = "AVG"
	FunctionMedian = "MEDIAN"
	FunctionMin    = "MIN"
	FunctionMax    = "MAX"

	FunctionSumIf    = "SUMIF"
	FunctionCountIf  = "COUNTIF"
	FunctionAvgIf    = "AVGIF"
	FunctionMedianIf = "MEDIANIF"
	FunctionMinIf    = "MINIF"
	FunctionMaxIf    = "MAXIF"
)

var (
	Role = cel.OpaqueType("_Role")
)

func MapType(schema []*parser.AST, typeName string, isRepeated bool) (*types.Type, error) {
	// For single operand conditions
	if typeName == parser.FieldTypeBoolean && !isRepeated {
		return types.BoolType, nil
	}

	switch typeName {
	case parser.FieldTypeID,
		parser.FieldTypeText,
		parser.FieldTypeMarkdown,
		parser.FieldTypeNumber,
		parser.FieldTypeBoolean,
		parser.FieldTypeDecimal,
		parser.FieldTypeTimestamp,
		parser.FieldTypeDate,
		parser.FieldTypeFile,
		parser.FieldTypeVector,
		parser.FieldTypeSecret,
		parser.FieldTypePassword,
		parser.FieldTypeDuration:
		if isRepeated {
			return cel.OpaqueType(fmt.Sprintf("%s[]", typeName)), nil
		} else {
			return cel.OpaqueType(typeName), nil
		}

	case Role.String(), "_ActionType", "_FieldName":
		if isRepeated {
			typeName = typeName + "[]"
		}
		return types.NewOpaqueType(typeName), nil
	}

	switch {
	case query.Enum(schema, typeName) != nil:
		if isRepeated {
			typeName = typeName + "[]"
		}
		return types.NewOpaqueType(typeName), nil
	case query.Entity(schema, typeName) != nil:
		if isRepeated {
			typeName = typeName + "[]"
		}
		return types.NewObjectType(typeName), nil
	}

	return nil, fmt.Errorf("unknown type '%s'", typeName)
}
