package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func DuplicateModelNames(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterModel: func(n *parser.ModelNode) {
			if n.BuiltIn {
				return
			}

			for _, model := range query.Models(asts) {
				if n == model {
					continue
				}

				if model.Name.Value == n.Name.Value {
					var message string
					if model.BuiltIn {
						message = fmt.Sprintf("There already exists a reserved model with the name '%s'", n.Name.Value)
					} else {
						message = fmt.Sprintf("There already exists a model with the name '%s'", n.Name.Value)
					}

					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.NamingError,
							errorhandling.ErrorDetails{
								Message: message,
								Hint:    "Use unique names between models, enums and messages",
							},
							n.Name,
						),
					)
					break
				}
			}

			for _, enum := range query.Enums(asts) {
				if enum.Name.Value == n.Name.Value {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.NamingError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("There already exists an enum with the name '%s'", n.Name.Value),
								Hint:    "Use unique names between models, enums and messages",
							},
							n.Name,
						),
					)
					break
				}
			}

			for _, message := range query.Messages(asts) {
				if message.Name.Value == n.Name.Value {
					var m string
					if message.BuiltIn {
						m = fmt.Sprintf("There already exists a reserved message with the name '%s'", n.Name.Value)
					} else {
						m = fmt.Sprintf("There already exists a message with the name '%s'", n.Name.Value)
					}

					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.NamingError,
							errorhandling.ErrorDetails{
								Message: m,
								Hint:    "Use unique names between models, enums and messages",
							},
							n.Name,
						),
					)
					break
				}
			}
		},
	}
}
