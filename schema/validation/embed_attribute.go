package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
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
						Message: "@embed requires at least one argument",
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

			ident, err := resolve.AsIdent(arg.Expression)
			if err != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "The @embed attribute can only be used with valid model fields",
						Hint:    "For example, use @embed(fieldName)",
					},
					arg,
				))
				return
			}

			// now we go through the identifier fragments and ensure that they are relationships
			model := currentModel
			for _, fragment := range ident.Fragments {
				// get the field in the relationship fragments
				currentField := query.ModelField(model, fragment)
				if currentField == nil {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%s is not a field in the %s model", fragment, model.Name.Value),
							Hint:    "The @embed attribute must reference an existing model field",
						},
						ident,
					))

					return
				}

				// model will be null if this is not a model field
				model = query.Model(asts, currentField.Type.Value)
				if model == nil {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%s is not a model field", currentField.Name.Value),
							Hint:    "The @embed attribute must reference a related model field",
						},
						ident,
					))

					return
				}
			}

			if lo.SomeBy(arguments, func(a string) bool { return a == ident.ToString() }) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("@embed argument '%s' already defined within this action", ident.ToString()),
					},
					ident,
				))
				return
			}

			arguments = append(arguments, ident.ToString())
		},
	}
}
