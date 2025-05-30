package validation

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// ConflictingValueInputsRule checks for model value inputs that are also used in @set
// In this case one usage would cancel out the other,
// so it doesn't make sense to have both.
//
// For example, createPost() with (title) { @set(post.title = title) }.
func ConflictingValueInputsRule(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	isAction := false
	var action *parser.ActionNode
	writeInputs := map[*parser.ActionInputNode]bool{}

	return Visitor{
		EnterModelSection: func(n *parser.ModelSectionNode) {
			isAction = len(n.Actions) > 0
		},
		LeaveModelSection: func(n *parser.ModelSectionNode) {
			isAction = false
		},
		EnterAction: func(n *parser.ActionNode) {
			writeInputs = map[*parser.ActionInputNode]bool{}
			action = n
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			if n.Label == nil && isAction {
				if lo.Contains(action.With, n) {
					writeInputs[n] = true
				}
			}
		},
		EnterAttribute: func(n *parser.AttributeNode) {
			isRelevantAttr := lo.Contains([]string{parser.AttributeSet}, n.Name.Value)
			if !isAction || !isRelevantAttr || len(n.Arguments) == 0 {
				return
			}

			expr := n.Arguments[0].Expression
			if expr == nil {
				return
			}

			lhs, _, err := n.Arguments[0].Expression.ToAssignmentExpression()
			if err != nil {
				return
			}

			idents, err := resolve.IdentOperands(lhs)
			if err != nil {
				return
			}

			for _, operand := range idents {
				for in := range writeInputs {
					// in an expression the first ident fragment will be the model name
					// we create an indent without the first fragment
					identWithoutModelName := operand.Fragments[1:]

					if in.Type.ToString() != strings.Join(identWithoutModelName, ".") {
						continue
					}

					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.ActionInputError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("%s is already being used as a value input so cannot also be used in @set", in.Type.ToString()),
							},
							operand,
						),
					)
				}
			}
		},
	}
}
