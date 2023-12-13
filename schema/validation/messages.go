package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func MessagesRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var message *parser.MessageNode
	fieldNames := map[string]bool{}

	return Visitor{
		EnterMessage: func(m *parser.MessageNode) {
			message = m
		},
		LeaveMessage: func(m *parser.MessageNode) {
			message = nil
			fieldNames = map[string]bool{}
		},
		EnterField: func(f *parser.FieldNode) {
			if message == nil {
				return
			}

			// Duplicate field names check
			if _, ok := fieldNames[f.Name.Value]; ok {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.DuplicateDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("field '%s' already defined in message '%s'", f.Name.Value, message.Name.Value),
					},
					f.Name,
				))
			}
			fieldNames[f.Name.Value] = true

			// Type check
			if !parser.IsBuiltInFieldType(f.Type.Value) &&
				!query.IsUserDefinedType(asts, f.Type.Value) &&
				query.Message(asts, f.Type.Value) == nil &&
				f.Type.Value != parser.MessageFieldTypeAny {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.TypeError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("invalid type '%s' - must be a built-in type, model, enum, or message", f.Type.Value),
					},
					f.Type,
				))
			}

		},
		EnterAttribute: func(a *parser.AttributeNode) {
			if message == nil {
				return
			}

			errs.AppendError(errorhandling.NewValidationErrorWithDetails(
				errorhandling.AttributeNotAllowedError,
				errorhandling.ErrorDetails{
					Message: "message fields do not support attributes",
				},
				a.Name,
			))
		},
	}
}
