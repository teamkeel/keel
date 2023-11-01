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

			// This is a hacky way of checking if this is a message, which we'll skip from the validation.
			// Otherwise we could run into a duplicate input validation error on this: write writeFn(Any) returns (Any)
			// I would really like to differentiate between Input and Output nodes on the AST, as it would make validation much easier (and the rules are different).
			// My original proposed PR on the matter: https://github.com/teamkeel/keel/pull/1016/files#diff-f880f21e2ba759b058ddc06776b9962736266b247d4b38e5d703a0802ca08d6d
			// if strcase.ToCamel(input) == input {
			// 	return
			// }

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
