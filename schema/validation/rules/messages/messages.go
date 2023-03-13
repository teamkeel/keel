package messages

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueMessageNamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	seenMessageNames := map[string]bool{}

	for _, message := range query.Messages(asts) {
		if _, ok := seenMessageNames[message.Name.Value]; ok {
			errs.AppendError(
				errorhandling.NewValidationErrorWithDetails(
					errorhandling.DuplicateDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("message type '%s' is already defined", message.Name.Value),
						Hint:    "Please use a different name",
					},
					message,
				),
			)

			continue
		}

		seenMessageNames[message.Name.Value] = true
	}

	return
}
