package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// DuplicateDefinitionRule checks the uniqueness of model, enum, message, and job names.
//
// Model, enum, and message names need to be globally unique.
// There cannot be two jobs with the same name
func DuplicateDefinitionRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	names := map[string]string{}
	jobs := map[string]bool{}

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
		EnterJob: func(n *parser.JobNode) {
			if _, ok := jobs[n.Name.Value]; ok {
				errs.AppendError(
					duplicateDefinitionError(n.Name, "job"),
				)
				return
			}
			jobs[n.Name.Value] = true
		},
	}
}

func duplicateDefinitionError(n parser.NameNode, existingEntity string) *errorhandling.ValidationError {
	hint := "Use unique names between models, enums and messages"
	if existingEntity == "job" {
		hint = "Job names must be unique"
	}

	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.NamingError,
		errorhandling.ErrorDetails{
			Message: fmt.Sprintf("There is already a %s with the name %s", existingEntity, n.Value),
			Hint:    hint,
		},
		n,
	)
}
