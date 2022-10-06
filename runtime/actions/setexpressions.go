package actions

import (
	"context"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
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

		modelMap[strcase.ToSnake(fieldName)], err = evaluateOperandValue(ctx, assignment.RHS, operation, schema, values, lhsOperandType)
		if err != nil {
			return nil, err
		}
	}

	return modelMap, nil
}
