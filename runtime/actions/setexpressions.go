package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/expressions"
)

// Given an operation with set expressions, will return a model map with any explicit
// args into a map[string]any
func SetExpressionInputsToModelMap(operation *proto.Operation, values map[string]any, schema *proto.Schema, ctx context.Context) (map[string]any, error) {
	modelMap := map[string]any{}

	for _, setExpression := range operation.SetExpressions {
		expression, err := expressions.Parse(setExpression.Source)
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
			} else if proto.EnumExists(schema.Enums, assignment.RHS.Ident.Fragments[0].Fragment) {
				modelMap[strcase.ToSnake(fieldName)] = assignment.RHS.Ident.Fragments[1].Fragment
			} else {
				// check if there is a match for the set expression in explicit inputs

				rhsIdent := assignment.RHS.Ident

				if match, ok := values[rhsIdent.ToString()]; ok {
					modelMap[strcase.ToSnake(fieldName)] = match

					continue
				}

				return nil, fmt.Errorf("operand type not yet supported: %s", fieldName)
			}
		default:
			return nil, fmt.Errorf("operand type not yet supported")
		}
	}

	return modelMap, nil
}
