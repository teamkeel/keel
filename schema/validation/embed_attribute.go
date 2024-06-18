package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func EmbedAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	var currentOperation *parser.ActionNode
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
			currentOperation = action
		},
		LeaveAction: func(_ *parser.ActionNode) {
			currentOperation = nil
			arguments = []string{}
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			currentAttribute = attribute

			if attribute.Name.Value != parser.AttributeEmbed {
				return
			}

			if currentOperation == nil {
				return
			}

			if currentOperation.Type.Value != parser.ActionTypeList && currentOperation.Type.Value != parser.ActionTypeGet {
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

			if currentOperation == nil {
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
						Hint:    "For example, use @embed(firstName)",
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
						Hint:    "For example, use @embed(firstName)",
					},
					arg,
				))
				return
			}
			var field *parser.FieldNode
			var model = currentModel

			// Iterate through each fragment in the LHS operand, and ensure:
			// - it is a field which is part of a model being created or updated (including nested creates)
			for _, fragment := range operand.Ident.Fragments {
				// get the next field in the relationship fragments
				field = query.ModelField(model, fragment.Fragment)
				if field == nil {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("The @embed attribute (%s) does not exist in model %s", fragment.Fragment, model.Name.Value),
							Hint:    "For example, use @embed(firstName)",
						},
						arg,
					))
					return
				}
				// will be null if this is not a model field
				model = query.Model(asts, field.Type.Value)
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
