package validation

import (
	"fmt"

	"github.com/samber/lo"
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
	modelInputs := map[*parser.ActionInputNode]bool{}

	return Visitor{
		EnterModelSection: func(n *parser.ModelSectionNode) {
			isAction = len(n.Actions) > 0
		},
		LeaveModelSection: func(n *parser.ModelSectionNode) {
			isAction = false
		},
		EnterAction: func(n *parser.ActionNode) {
			modelInputs = map[*parser.ActionInputNode]bool{}
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			if n.Label == nil && isAction {
				modelInputs[n] = true
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

			for _, cond := range expr.Conditions() {
				operands := []*parser.Operand{cond.LHS}
				if n.Name.Value == parser.AttributeWhere {
					operands = append(operands, cond.RHS)
				}

				for _, operand := range operands {
					if operand == nil || operand.Ident == nil {
						continue
					}
					for in := range modelInputs {
						// in an expression the first ident fragment will be the model name
						// we create an indent without the first fragment
						identWithoutModelName := &parser.Ident{
							Fragments: operand.Ident.Fragments[1:],
						}

						if in.Type.ToString() != identWithoutModelName.ToString() {
							continue
						}

						errs.AppendError(
							errorhandling.NewValidationErrorWithDetails(
								errorhandling.ActionInputError,
								errorhandling.ErrorDetails{
									Message: fmt.Sprintf("%s is already being used as an input so cannot also be used in an expression", in.Type.ToString()),
								},
								operand.Ident,
							),
						)
					}
				}
			}
		},
	}
}
