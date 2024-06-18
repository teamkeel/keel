package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func EmbedAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	var currentAction *parser.ActionNode
	var currentAttribute *parser.AttributeNode
	var arguments []string

	return Visitor{
		EnterModel: func(model *parser.ModelNode) {
			currentModel = model
		},
		LeaveModel: func(_ *parser.ModelNode) {
			currentModel = nil
		},
		EnterAction: func(action *parser.ActionNode) {
			currentAction = action
		},
		LeaveAction: func(_ *parser.ActionNode) {
			currentAction = nil
			arguments = []string{}
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			currentAttribute = attribute

			if attribute.Name.Value != parser.AttributeEmbed {
				return
			}

			if currentAction == nil {
				return
			}

			if currentAction.Type.Value != parser.ActionTypeList && currentAction.Type.Value != parser.ActionTypeGet {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@embed can only be used on list or get actions",
					},
					attribute.Name,
				))
			}

			if len(attribute.Arguments) == 0 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@embed requires at least once argument",
					},
					attribute,
				))
			}
		},
		LeaveAttribute: func(attribute *parser.AttributeNode) {
			currentAttribute = nil
		},
		EnterAttributeArgument: func(arg *parser.AttributeArgumentNode) {
			if currentAttribute.Name.Value != parser.AttributeEmbed {
				return
			}

			if currentAction == nil {
				return
			}

			if arg.Label != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@embed arguments should not be labelled",
						Hint:    "For example, use @embed(firstName, surname)",
					},
					arg,
				))
				return
			}

			if !arg.Expression.IsValue() {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@embed argument is not correctly formatted",
						Hint:    "For example, use @embed(user.firstName)",
					},
					arg,
				))
				return
			}

			conditions := arg.Expression.Conditions()
			if len(conditions) > 1 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "An @embed attribute can only consist of model fields references",
						Hint:    "For example, use @embed(user.firstName)",
					},
					arg,
				))
				return
			}

			if conditions[0].Type() != parser.ValueCondition {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "An @embed attribute must be a value condition",
						Hint:    "For example, use @embed(user.surname)",
					},
					arg,
				))
				return
			}

			if conditions[0].Type() == parser.LogicalCondition {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "An @embed attribute cannot be a logical condition",
						Hint:    "For example, use @embed(user.surname)",
					},
					arg,
				))
				return
			}

			operand, err := arg.Expression.ToValue()
			if err != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "Ab @embed argument must reference a field",
						Hint:    "For example, use @embed(author.firstName)",
					},
					arg,
				))
				return
			}

			if operand.Ident == nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "The @embed attribute can only be used with valid model fields",
						Hint:    "For example, use @embed(author.firstName)",
					},
					arg,
				))
				return
			}

			expressionContext := expressions.ExpressionContext{
				Model:     currentModel,
				Attribute: currentAttribute,
				Action:    currentAction,
			}

			// We resolve whether the actual fragments are valid idents in other validations,
			// but we need to exit early here if they dont resolve.
			resolver := expressions.NewOperandResolver(operand, asts, &expressionContext, expressions.OperandPositionLhs)
			_, rerr := resolver.Resolve()
			if rerr != nil {
				errs.AppendError(rerr.ToValidationError())
			}

			if lo.SomeBy(arguments, func(a string) bool { return a == operand.Ident.ToString() }) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("@embed argument name '%s' already defined", operand.Ident.ToString()),
					},
					arg.Expression,
				))
				return
			}

			arguments = append(arguments, operand.Ident.ToString())
		},
	}
}
