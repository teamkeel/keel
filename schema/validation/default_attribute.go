package validation

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func DefaultAttributeExpressionRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var field *parser.FieldNode
	var attribute *parser.AttributeNode

	return Visitor{
		EnterField: func(f *parser.FieldNode) {
			field = f
		},
		LeaveField: func(_ *parser.FieldNode) {
			field = nil
		},
		EnterAttribute: func(a *parser.AttributeNode) {
			attribute = a

			if a == nil || a.Name.Value != parser.AttributeDefault {
				return
			}

			for _, attr := range field.Attributes {
				if attr.Name.Value == parser.AttributeComputed {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeExpressionError,
						errorhandling.ErrorDetails{
							Message: "@default cannot be used with computed fields",
							Hint:    "Either remove the @default attribute or remove the @computed attribute",
						},
						a,
					))
				}
			}

			typesWithZeroValue := []string{"Text", "Number", "Boolean", "ID", "Timestamp"}
			if len(a.Arguments) == 0 && !lo.Contains(typesWithZeroValue, field.Type.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@default requires an expression",
						Hint:    "Try @default(MyDefaultValue) instead",
					},
					a,
				))
			}
		},
		LeaveAttribute: func(*parser.AttributeNode) {
			attribute = nil
		},
		EnterExpression: func(expression *parser.Expression) {
			if attribute.Name.Value != parser.AttributeDefault {
				return
			}

			issues, err := attributes.ValidateDefaultExpression(asts, field, expression)
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
