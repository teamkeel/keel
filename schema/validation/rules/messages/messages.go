package messages

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func MessageNamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, message := range query.Messages(asts) {
		if message.Name.Value != strcase.ToCamel(message.Name.Value) {
			errs.AppendError(
				errorhandling.NewValidationErrorWithDetails(
					errorhandling.NamingError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("'%s' is not in upper camel case", message.Name.Value),
						Hint:    fmt.Sprintf("Use '%s' instead", strcase.ToCamel(message.Name.Value)),
					},
					message.Name,
				),
			)
		}
	}

	return
}
