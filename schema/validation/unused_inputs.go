package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// UnusedInputRule checks that all named action inputs are
// used in either @set or @where expressions in the action.
func UnusedInputRule(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	isAction := false
	unused := map[string]*parser.NameNode{}

	return Visitor{
		EnterAction: func(n *parser.ActionNode) {
			unused = map[string]*parser.NameNode{}
			isAction = true
		},
		LeaveAction: func(n *parser.ActionNode) {
			isAction = false

			// if the action is implemented as a function, then we don't know how the function
			// uses the inputs (if at all), so not a validation error.
			if n.IsFunction() {
				return
			}

			for _, name := range unused {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.ActionInputError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%s is not used. Labelled inputs must be used in the action, for example in a @set or @where attribute", name.Value),
						},
						name,
					),
				)
			}
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			if n.Label == nil {
				return
			}
			unused[n.Label.Value] = n.Label
		},
		EnterAttribute: func(n *parser.AttributeNode) {
			// TODO: @permission should not be in this list as we don't want to
			// allow inputs to be used in permission expressions, however we have
			// some test schemas that do this so they all need to be updated before
			// we can remove it here
			relevantAttributes := []string{parser.AttributeWhere, parser.AttributeSet, parser.AttributePermission}

			isRelevantAttr := lo.Contains(relevantAttributes, n.Name.Value)
			if !isAction || !isRelevantAttr || len(n.Arguments) == 0 {
				return
			}

			var expression *parser.Expression
			if n.Name.Value == parser.AttributeSet {
				_, rhs, err := n.Arguments[0].Expression.ToAssignmentExpression()
				if err != nil {
					return
				}
				expression = rhs
			} else {
				expression = n.Arguments[0].Expression
				if expression == nil {
					return
				}
			}

			operands, err := resolve.IdentOperands(expression)
			if err != nil {
				return
			}

			for _, operand := range operands {
				for k, u := range unused {
					if u.Value == operand.ToString() {
						delete(unused, k)
					}
				}
			}
		},
	}
}
