package expression

import (
	"github.com/teamkeel/keel/schema/associations"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
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

func ValidateConditionSide(asts []*parser.AST, operand expressions.Operand) error {
	if operand.Ident != nil {
		tree, err := associations.TryResolveIdent(asts, operand.Ident)

		if err != nil {
			unresolved := tree.ErrorFragment()

			if len(tree.Fragments) == 1 {
				return errorhandling.NewValidationError(errorhandling.ErrorUnresolvedRootModel,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Type":       "association",
							"Root":       unresolved.Ident,
							"Suggestion": "",
						},
					},
					unresolved.Node,
				)
			}

			return errorhandling.NewValidationError(errorhandling.ErrorUnresolvableExpression,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Type":        "association",
						"Fragment":    unresolved.Ident,
						"Parent":      tree.Fragments[len(tree.Fragments)-2].ToString(),
						"Suggestions": "",
					},
				},
				unresolved.Node,
			)
		}
	}

	return nil
}
