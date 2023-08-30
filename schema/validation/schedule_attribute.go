package validation

import (
	"strings"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ScheduleAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterAttribute: func(attribute *parser.AttributeNode) {
			if attribute.Name.Value != parser.AttributeSchedule {
				return
			}

			if len(attribute.Arguments) != 1 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@schedule must have exactly one argument",
					},
					attribute.Name,
				))
			}

			if attribute.Arguments[0].Label != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@schedule must not have a label",
					},
					attribute.Name,
				))
			}

			operand := attribute.Arguments[0].Expression.Tokens[0].Value
			removed := strings.ReplaceAll(operand, "\"", "")
			if removed == "" {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@schedule argument is not correctly formatted",
						Hint:    "schedule should be in the following format @schedule(\"0 6 * * * *\")",
					},
					attribute.Name,
				))
			}
		},
	}

}
