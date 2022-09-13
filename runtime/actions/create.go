package actions

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/expressions"

	"github.com/iancoleman/strcase"
)

func Create(ctx context.Context, operation *proto.Operation, schema *proto.Schema, args map[string]any) (map[string]any, error) {
	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}
	model := proto.FindModel(schema.Models, operation.ModelName)
	modelMap, err := initialValueForModel(model, schema)
	if err != nil {
		return nil, err
	}

	// Now overwrite the fields for which Inputs have been given accordingly.
	for _, input := range operation.Inputs {
		switch input.Behaviour {
		case proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT:
			modelFieldName := input.Target[0]

			// If this argument is missing it must be optional.
			v, ok := args[input.Name]
			if !ok {
				continue
			}
			v, err := toMap(v, input.Type.Type)
			if err != nil {
				return nil, err
			}
			modelMap[strcase.ToSnake(modelFieldName)] = v
		default:
			return nil, fmt.Errorf("input behaviour %s is not yet supported for Create", input.Behaviour)
		}
	}

	for _, setExpressions := range operation.SetExpressions {
		expression, err := expressions.Parse(setExpressions.Source)
		if err != nil {
			return nil, err
		}

		assignment, err := expressions.ToAssignmentCondition(expression)
		if err != nil {
			return nil, err
		}

		lhsOperandType, err := GetOperandType(assignment.LHS, operation, schema)
		if err != nil {
			return nil, err
		}

		fieldName := assignment.LHS.Ident.Fragments[1].Fragment
		isLiteral, _ := assignment.RHS.IsLiteralType()

		switch {
		case isLiteral:
			modelMap[strcase.ToSnake(fieldName)], err = toNative(assignment.RHS, lhsOperandType)
			if err != nil {
				return nil, err
			}
		case assignment.RHS.Type() == expressions.TypeIdent:
			if assignment.RHS.Ident.IsContextIdentityField() {
				modelMap[strcase.ToSnake(fieldName)], err = runtimectx.GetIdentity(ctx)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("operand type not yet supported: %s", fieldName)
			}
		default:
			return nil, fmt.Errorf("operand type not yet supported")
		}
	}

	// Write a row to the database.
	if err := db.Table(strcase.ToSnake(model.Name)).Create(modelMap).Error; err != nil {
		return nil, err
	}
	return toLowerCamelMap(modelMap), nil
}

// toLowerCamelMap returns a copy of the given map, in which all
// of the key strings are converted to LowerCamelCase.
// It is good for converting identifiers typically used as database
// table or column names, to the case requirements stipulated by the Keel schema.
func toLowerCamelMap(m map[string]any) map[string]any {
	res := map[string]any{}
	for key, value := range m {
		res[strcase.ToLowerCamel(key)] = value
	}
	return res
}

// toLowerCamelMaps is a convenience wrapper around toLowerCamelMap
// that operates on a list of input maps - rather than just a single map.
func toLowerCamelMaps(maps []map[string]any) []map[string]any {
	res := []map[string]any{}
	for _, m := range maps {
		res = append(res, toLowerCamelMap(m))
	}
	return res
}
