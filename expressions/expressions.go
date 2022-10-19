package expressions

import (
	"github.com/teamkeel/keel/schema/parser"
)

func ResolveCondition(asts []*parser.AST, c *parser.Condition, context ExpressionContext) (resolvedLhs *ExpressionScopeEntity, resolvedRhs *ExpressionScopeEntity, errors []error) {
	lhs := c.LHS
	rhs := c.RHS

	scope := BuildRootExpressionScope(asts, context)

	resolvedLhs, lhsErr := ResolveOperand(asts, lhs, scope, context, OperandPositionLhs)

	if lhsErr != nil {
		errors = append(errors, lhsErr.ToValidationError())
	}

	if rhs != nil {
		resolvedRhs, rhsErr := ResolveOperand(asts, rhs, scope, context, OperandPositionRhs)

		if rhsErr != nil {
			errors = append(errors, rhsErr.ToValidationError())
		}

		return resolvedLhs, resolvedRhs, errors
	}

	return resolvedLhs, nil, errors
}
