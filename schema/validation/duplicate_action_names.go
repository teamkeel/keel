package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func DuplicateActionNames(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterAction: func(n *parser.ActionNode) {
			if n.BuiltIn {
				return
			}

			for _, model := range query.Models(asts) {
				for _, action := range query.ModelActions(model) {
					if n == action {
						continue
					}

					if action.Name.Value == n.Name.Value {
						var message string
						if action.BuiltIn {
							message = fmt.Sprintf("There already exists a reserved action with the name '%s'", n.Name.Value)
						} else {
							message = fmt.Sprintf("There already exists an action with the name '%s'", n.Name.Value)
						}

						errs.AppendError(
							errorhandling.NewValidationErrorWithDetails(
								errorhandling.NamingError,
								errorhandling.ErrorDetails{
									Message: message,
								},
								n.Name,
							),
						)
						break
					}
				}
			}
		},
	}
}
