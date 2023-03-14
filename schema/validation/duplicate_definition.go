package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// DuplicateDefinitionRule checks the uniqueness of model, enum and message names.
//
// Model, enum, and message names need to be globally unique.
func DuplicateDefinitionRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	names := map[string]string{}

	return Visitor{
		EnterModel: func(n *parser.ModelNode) {
			if entity, ok := names[n.Name.Value]; ok {
				errs.AppendError(
					duplicateDefinitionError(n.Name, entity),
				)
				return
			}
			names[n.Name.Value] = "model"
		},
		EnterEnum: func(n *parser.EnumNode) {
			if entity, ok := names[n.Name.Value]; ok {
				errs.AppendError(
					duplicateDefinitionError(n.Name, entity),
				)
				return
			}
			names[n.Name.Value] = "enum"
		},
		EnterMessage: func(n *parser.MessageNode) {
			if entity, ok := names[n.Name.Value]; ok {
				errs.AppendError(
					duplicateDefinitionError(n.Name, entity),
				)
				return
			}
			names[n.Name.Value] = "message"
		},
	}
}

func duplicateDefinitionError(n parser.NameNode, existingEntity string) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.NamingError,
		errorhandling.ErrorDetails{
			Message: fmt.Sprintf("There is already a %s with the name %s", existingEntity, n.Value),
			Hint:    "Use unique names between models, enums and messages",
		},
		n,
	)
}
