package expression

import (
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
)

type ResolvedValue struct {
	*node.Node

	Type string
}

func ValidateExpressionConditions(asts []*parser.AST, arg *parser.AttributeArgumentNode, validateFunc func(lhs *expressions.Operand, operator *expressions.Operator, rhs *expressions.Operand) []error) (errs []error) {
	// get all of the nested conditions in the expression
	conditions := arg.Expression.Conditions()

	for _, condition := range conditions {

		// conditionType := condition.Type()
		lhs, operator, rhs := condition.ToFragments()

		result := validateFunc(lhs, operator, rhs)

		if len(errs) > 0 {
			errs = append(errs, result...)
		}
	}

	return errs
}
