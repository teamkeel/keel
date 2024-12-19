package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func WhereAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var action *parser.ActionNode
	var attribute *parser.AttributeNode

	return Visitor{
		EnterAction: func(a *parser.ActionNode) {
			action = a
		},
		LeaveAction: func(*parser.ActionNode) {
			action = nil
		},
		EnterAttribute: func(attr *parser.AttributeNode) {
			if attr.Name.Value != parser.AttributeWhere {
				return
			}

			attribute = attr

			if len(attr.Arguments) != 1 {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%v argument(s) provided to @unique but expected 1", len(attr.Arguments)),
						},
						attr,
					),
				)
			}
		},
		LeaveAttribute: func(*parser.AttributeNode) {
			attribute = nil
		},
		EnterExpression: func(expression *parser.Expression) {
			if attribute == nil {
				return
			}

			issues, err := attributes.ValidateWhereExpression(asts, action, expression)
			if err != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeExpressionError,
					errorhandling.ErrorDetails{
						Message: "expression could not be parsed",
					},
					expression))
			}

			if len(issues) > 0 {
				for _, issue := range issues {
					errs.AppendError(issue)
				}
			}
		},
	}
}
