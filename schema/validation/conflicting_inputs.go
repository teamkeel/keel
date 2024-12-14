package validation

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// ConflictingInputsRule checks for model inputs that are also used in @set
// or @where attributes. In this case one usage would cancel out the other,
// so it doesn't make sense to have both.
//
// For example in the getThing operation `id` is listed as a model input but
// is also used in a @where expression.
func ConflictingInputsRule(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	isAction := false
	var action *parser.ActionNode
	filterInputs := map[*parser.ActionInputNode]bool{}
	writeInputs := map[*parser.ActionInputNode]bool{}

	return Visitor{
		EnterModelSection: func(n *parser.ModelSectionNode) {
			isAction = len(n.Actions) > 0
		},
		LeaveModelSection: func(n *parser.ModelSectionNode) {
			isAction = false
		},
		EnterAction: func(n *parser.ActionNode) {
			filterInputs = map[*parser.ActionInputNode]bool{}
			writeInputs = map[*parser.ActionInputNode]bool{}
			action = n
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			if n.Label == nil && isAction {
				if lo.Contains(action.Inputs, n) {
					filterInputs[n] = true
				}
				if lo.Contains(action.With, n) {
					writeInputs[n] = true
				}
			}
		},
		EnterAttribute: func(n *parser.AttributeNode) {
			isRelevantAttr := lo.Contains([]string{parser.AttributeWhere, parser.AttributeSet}, n.Name.Value)
			if !isAction || !isRelevantAttr || len(n.Arguments) == 0 {
				return
			}

			expr := n.Arguments[0].Expression
			if expr == nil {
				return
			}

			var inputs map[*parser.ActionInputNode]bool
			if n.Name.Value == parser.AttributeWhere {
				inputs = filterInputs
			} else if n.Name.Value == parser.AttributeSet {
				inputs = writeInputs
			}

			idents := []*parser.ExpressionIdent{}
			var err error
			switch n.Name.Value {
			case parser.AttributeWhere:
				idents, err = resolve.IdentOperands(n.Arguments[0].Expression)
			case parser.AttributeSet:
				lhs, _, err := n.Arguments[0].Expression.ToAssignmentExpression()
				if err != nil {
					return
				} else {
					idents, err = resolve.IdentOperands(lhs)
				}
			}
			if err != nil {
				return
			}

			for _, operand := range idents {
				for in := range inputs {
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
								Message: fmt.Sprintf("%s is already being used as an input so cannot also be used in an expression", in.Type.ToString()),
							},
							n.Arguments[0].Expression,
						),
					)
				}
			}
		},
	}
}
