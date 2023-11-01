package validation

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// DuplicateInputsRule checks that input names are not duplicated for an action.
func DuplicateInputsRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var queryInputs []string
	var writeInputs []string
	var action *parser.ActionNode

	return Visitor{
		EnterAction: func(n *parser.ActionNode) {
			queryInputs = []string{}
			writeInputs = []string{}
			action = n
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			var input string

			if n.Label != nil {
				input = n.Label.Value
			} else {
				fragments := []string{}
				for _, frag := range n.Type.Fragments {
					fragments = append(fragments, frag.Fragment)
				}

				input = strings.Join(fragments, ".")
			}

			switch {
			case lo.Contains(action.Inputs, n):
				if lo.Contains(queryInputs, input) {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.NamingError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("'%s' has already been defined as a query input on this action", input),
						},
						n))
				} else {
					queryInputs = append(queryInputs, input)
				}
			case lo.Contains(action.With, n):
				if lo.Contains(writeInputs, input) {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.NamingError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("'%s' has already been defined as a write input on this action", input),
						},
						n))
				} else {
					writeInputs = append(writeInputs, input)
				}
			}
		},
	}
}
