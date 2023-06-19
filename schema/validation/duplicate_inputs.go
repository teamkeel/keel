package validation

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// DuplicateInputsRule checks that inputs are not duplicated as inputs.
func DuplicateInputsRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	names := []string{}

	return Visitor{
		EnterAction: func(n *parser.ActionNode) {
			names = []string{}
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			var input string
			if n.Label != nil {
				input = n.Label.Value
			} else {
				frags := lo.Map(n.Type.Fragments, func(f *parser.IdentFragment, _ int) string { return f.Fragment })
				input = strings.Join(frags, ".")
			}

			if lo.Contains(names, input) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.NamingError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("This input '%s' has already been defined as an input", input),
					},
					n))
			} else {
				names = append(names, input)
			}
		},
	}
}
