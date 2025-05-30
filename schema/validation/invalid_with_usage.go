package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	ValidActionTypes = []string{parser.ActionTypeCreate, parser.ActionTypeUpdate}
)

// InvalidWithUsage checks that the 'with' keyword is only used for actions that receive write values.
func InvalidWithUsage(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterAction: func(action *parser.ActionNode) {
			if len(action.With) > 0 && !lo.Contains(ValidActionTypes, action.Type.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.ActionInputError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("The 'with' keyword cannot be used with the '%s' action type", action.Type.Value),
						Hint:    fmt.Sprintf("'with' can only be used with %s", formatting.HumanizeList(ValidActionTypes, formatting.DelimiterOr)),
					},
					action,
				))
			}
		},
	}
}
