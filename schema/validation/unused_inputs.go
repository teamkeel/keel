package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// UnusedInputRule checks that all named operation inputs are
// used in either @set or @where expressions in the action.
func UnusedInputRule(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	isOperation := false
	unused := map[string]*parser.NameNode{}

	return Visitor{
		EnterModelSection: func(n *parser.ModelSectionNode) {
			isOperation = len(n.Operations) > 0
		},
		LeaveModelSection: func(n *parser.ModelSectionNode) {
			isOperation = false
		},
		LeaveAction: func(n *parser.ActionNode) {
			if !isOperation {
				return
			}

			for _, name := range unused {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.ActionInputError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%s is not used. Labelled inputs must be used in the operation, for example in a @set or @where attribute", name.Value),
						},
						name,
					),
				)
			}

			unused = map[string]*parser.NameNode{}
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			if n.Label == nil || !isOperation {
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
			if !isOperation || !isRelevantAttr || len(n.Arguments) == 0 {
				return
			}

			expr := n.Arguments[0].Expression
			if expr == nil {
				return
			}

			for _, cond := range expr.Conditions() {
				for _, operand := range []*parser.Operand{cond.LHS, cond.RHS} {
					if operand == nil || operand.Ident == nil {
						continue
					}

					delete(unused, operand.Ident.ToString())
				}
			}
		},
	}
}
