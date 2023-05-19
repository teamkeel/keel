package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// FieldOfSameModelType ensures that a model cannot have a field of the same type,
// as this results in an infinite recursion.
func FieldOfSameMessageType(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var message *parser.MessageNode
	message = nil

	return Visitor{
		EnterMessage: func(m *parser.MessageNode) {
			message = m
		},
		LeaveMessage: func(m *parser.MessageNode) {
			message = nil
		},
		EnterField: func(f *parser.FieldNode) {
			if message != nil && message.Name.Value == f.Type.Value {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("The message '%s' cannot have a field of its own type.", f.Type.Value),
						},
						f.Type,
					),
				)
			}
		},
	}
}
