package actions

import (
	"github.com/teamkeel/keel/schema/parser"
)

// DRYCaptureSetValues updates the writeValues field in the given scope, with
// field/values that should be set for each of the given operation's Set expressions.
func DRYCaptureSetValues(scope *Scope, args RequestArguments) error {

	ctx := scope.context

	for _, setExpression := range scope.operation.SetExpressions {
		expression, err := parser.ParseExpression(setExpression.Source)
		if err != nil {
			return err
		}

		assignment, err := expression.ToAssignmentCondition()
		if err != nil {
			return err
		}

		lhsOperandType, err := GetOperandType(assignment.LHS, scope.operation, scope.schema)
		if err != nil {
			return err
		}

		fieldName := assignment.LHS.Ident.Fragments[1].Fragment

		scope.writeValues[fieldName], err = evaluateOperandValue(ctx, assignment.RHS, scope.operation, scope.schema, args, lhsOperandType)
		if err != nil {
			return err
		}
	}
	return nil
}
