package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func StudioFeatures(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterJob: func(n *parser.JobNode) {
			errs.AppendWarning(
				errorhandling.NewValidationErrorWithDetails(
					errorhandling.UnsupportedFeatureError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job definitions are not supported in Keel Studio: '%s'", n.Name.Value),
					},
					n.Name,
				),
			)
		},
		EnterAction: func(n *parser.ActionNode) {
			if n.IsFunction() && !n.BuiltIn {
				errs.AppendWarning(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.UnsupportedFeatureError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("Custom functions are not supported in Keel Studio: '%s'", n.Name.Value),
						},
						n.Name,
					),
				)
			}
		},
		EnterAttribute: func(n *parser.AttributeNode) {
			if n.Name.Value == parser.AttributeOn {
				errs.AppendWarning(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.UnsupportedFeatureError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("Event subscribers are not supported in Keel Studio"),
						},
						n.Name,
					),
				)
			}
		},
	}
}
