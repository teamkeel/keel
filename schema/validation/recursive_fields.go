package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func RecursiveFieldsRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var message *parser.MessageNode

	return Visitor{
		EnterMessage: func(m *parser.MessageNode) {
			message = m
		},
		LeaveMessage: func(m *parser.MessageNode) {
			message = nil
		},
		EnterField: func(f *parser.FieldNode) {
			var entity string
			path := []string{}

			switch {
			case message != nil:
				entity = "message"
				path = append(path, message.Name.Value)
			default:
				return
			}

			recursivePath := fieldIsRecursive(asts, path, f)
			if len(recursivePath) == 0 {
				return
			}

			message := fmt.Sprintf("a %s cannot refer to itself unless the field is optional", entity)
			if len(recursivePath) > 2 {
				message += " - "
				for i := 1; i < len(recursivePath)-1; i++ {
					if i > 1 {
						message += ", "
					}
					message += fmt.Sprintf("'%s' refers to '%s'", recursivePath[i], recursivePath[i+1])
				}
			}

			errs.AppendError(errorhandling.NewValidationErrorWithDetails(
				errorhandling.TypeError,
				errorhandling.ErrorDetails{
					Message: message,
				},
				f.Name,
			))
		},
	}
}

func fieldIsRecursive(asts []*parser.AST, path []string, field *parser.FieldNode) []string {
	// If a field is optional then it's ok to be recursive
	if field.Optional {
		return nil
	}
	if field.Repeated {
		return nil
	}

	path = append(path[:], field.Type.Value)

	// If this fields type matches the first type in the path, then there is recursion
	if field.Type.Value == path[0] {
		return path
	}

	var fields []*parser.FieldNode

	model := query.Model(asts, field.Type.Value)
	if model != nil {
		fields = query.ModelFields(model)
	}

	message := query.Message(asts, field.Type.Value)
	if message != nil {
		fields = message.Fields
	}

	for _, other := range fields {
		// We ignore recursion here as it's an error with a different type
		// to the one we're currently validating
		if other.Type.Value == field.Type.Value {
			continue
		}

		recursivePath := fieldIsRecursive(asts, path, other)
		if len(recursivePath) > 0 {
			return recursivePath
		}
	}

	return nil
}
