package validation

import (
	"fmt"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ComputedAttributeRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var field *parser.FieldNode
	var attribute *parser.AttributeNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m
		},
		LeaveModel: func(*parser.ModelNode) {
			model = nil
		},
		EnterField: func(f *parser.FieldNode) {
			field = f
		},
		LeaveField: func(n *parser.FieldNode) {
			field = nil
		},
		EnterAttribute: func(attr *parser.AttributeNode) {
			if field == nil || attr.Name.Value != parser.AttributeComputed {
				return
			}

			switch field.Type.Value {
			case parser.FieldTypeBoolean,
				parser.FieldTypeNumber,
				parser.FieldTypeDecimal:
				attribute = attr
			default:
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("@computed cannot be used on field of type %s", field.Type.Value),
						},
						attr,
					),
				)
			}

			if field.Repeated {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: "@computed cannot be used on repeated fields",
						},
						attr,
					),
				)
			}

			if len(attr.Arguments) != 1 {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%v argument(s) provided to @computed but expected 1", len(attr.Arguments)),
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

			issues, err := attributes.ValidateComputedExpression(asts, model, field, expression)
			if err != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeExpressionError,
					errorhandling.ErrorDetails{
						Message: "expression could not be parsed",
					},
					expression))
				return
			}

			if len(issues) > 0 {
				for _, issue := range issues {
					errs.AppendError(issue)
				}
			}

			operands, err := resolve.IdentOperands(expression)
			if err != nil {
				return
			}

			for _, operand := range operands {
				if len(operand.Fragments) < 2 {
					continue
				}

				f := query.Field(model, operand.Fragments[1])

				if f == field {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: "@computed expressions cannot reference itself",
							},
							operand,
						),
					)
				}
			}
		},
	}
}
