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
			} else {
				// check if there is a match for the set expression in explicit inputs
				rhsIdent := assignment.RHS.Ident

				if match, ok := values[rhsIdent.ToString()]; ok {
					typeForExplicitInput := getExplicitInputType(operation, rhsIdent.ToString())

					if typeForExplicitInput != nil {
						dbObject, err := toMap(match, *typeForExplicitInput)

						if err != nil {
							return nil, err
						}

						modelMap[strcase.ToSnake(fieldName)] = dbObject
					}

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

func getExplicitInputType(operation *proto.Operation, name string) *proto.Type {
	for _, input := range operation.Inputs {
		if input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		if input.Name == name {
			return &input.Type.Type
		}
	}

	return nil
}
