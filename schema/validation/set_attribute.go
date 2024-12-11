package validation

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	fieldsNotMutable = []string{
		parser.FieldNameCreatedAt,
		parser.FieldNameUpdatedAt,
	}
)

func SetAttributeExpressionRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var action *parser.ActionNode
	var attribute *parser.AttributeNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m
		},
		LeaveModel: func(_ *parser.ModelNode) {
			model = nil
		},
		EnterAction: func(a *parser.ActionNode) {
			action = a
		},
		LeaveAction: func(_ *parser.ActionNode) {
			action = nil
		},
		EnterAttribute: func(a *parser.AttributeNode) {
			attribute = a
		},
		LeaveAttribute: func(*parser.AttributeNode) {
			attribute = nil
		},
		EnterExpression: func(expression *parser.Expression) {
			if attribute.Name.Value != parser.AttributeSet {
				return
			}

			l, r, err := expression.ToAssignmentExpression()
			if err != nil {
				errs.AppendError(makeSetExpressionError(
					errorhandling.AttributeExpressionError,
					"the @set attribute must be an assignment expression",
					fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
					expression,
				))
				return
			}

			issues, err := attributes.ValidateSetExpression(asts, action, l, r)
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
				return
			}
		},
	}
}

func makeSetExpressionError(t errorhandling.ErrorType, message string, hint string, node node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		t,
		errorhandling.ErrorDetails{
			Message: message,
			Hint:    hint,
		},
		node,
	)
}
