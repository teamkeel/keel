package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// NotMutableInputs checks that the write action inputs for update and create aren't
// setting the id of the root model and aren't setting the createdAt and updatedAt fields on any models.
func NotMutableInputs(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var action *parser.ActionNode

	return Visitor{
		EnterAction: func(n *parser.ActionNode) {
			action = n
		},
		LeaveAction: func(n *parser.ActionNode) {
			action = nil
		},
		EnterActionInput: func(input *parser.ActionInputNode) {
			if action == nil {
				return
			}

			if action.Type.Value != parser.ActionTypeCreate && action.Type.Value != parser.ActionTypeUpdate {
				return
			}

			if input.Label != nil {
				return
			}

			if !lo.Contains(action.With, input) {
				return
			}

			fragments := input.Type.Fragments

			field := fragments[len(fragments)-1].Fragment

			if lo.Contains(fieldsNotMutable, field) {
				if len(fragments) > 1 && field == parser.FieldNameId {
					return
				}

				errs.AppendError(makeNotMutableInputError(
					fmt.Sprintf("Cannot set the field '%s' as it is a built-in field and can only be mutated internally", field),
					"Target another field on the model or remove the input entirely",
					input,
				))
				return
			}
		},
	}
}

func makeNotMutableInputError(message string, hint string, node node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.ActionInputError,
		errorhandling.ErrorDetails{
			Message: message,
			Hint:    hint,
		},
		node,
	)
}
