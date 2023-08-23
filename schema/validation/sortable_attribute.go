package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func SortableAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	var currentOperation *parser.ActionNode
	var currentAttribute *parser.AttributeNode
	var sortableAttributeDefined bool
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
			sortableAttributeDefined = false
		},
		LeaveAction: func(_ *parser.ActionNode) {
			currentOperation = nil
			sortableAttributeDefined = false
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			currentAttribute = attribute
			arguments = []string{}

			if attribute.Name.Value != parser.AttributeSortable {
				return
			}

			if currentOperation == nil {
				return
			}

			if currentOperation.Type.Value != parser.ActionTypeList {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@sortable can only be used on list actions",
					},
					attribute.Name,
				))
			}

			if sortableAttributeDefined {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@sortable can only be defined once per action",
					},
					attribute.Name,
				))
			}

			sortableAttributeDefined = true

			if len(attribute.Arguments) == 0 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@sortable requires at least once argument",
					},
					attribute,
				))
			}
		},
		LeaveAttribute: func(attribute *parser.AttributeNode) {
			currentAttribute = nil
			arguments = []string{}
		},
		EnterAttributeArgument: func(arg *parser.AttributeArgumentNode) {
			if currentAttribute.Name.Value != parser.AttributeSortable {
				return
			}

			if currentOperation == nil {
				return
			}

			if arg.Label != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@sortable arguments should not be labelled",
						Hint:    "For example, use @sortable(firstName, surname)",
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
						Message: "@sortable argument is not correctly formatted",
						Hint:    "For example, use @sortable(firstName, surname)",
					},
					arg,
				))
				return
			}

			if operand.Ident == nil || len(operand.Ident.Fragments) != 1 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@sortable argument is not correct formatted",
						Hint:    "For example, use @sortable(firstName, surname)",
					},
					arg.Expression,
				))
				return
			}

			argumentValue := operand.Ident.Fragments[0].Fragment
			modelField := query.ModelField(currentModel, argumentValue)

			if modelField == nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("@sortable argument '%s' must correspond to a field on this model", argumentValue),
					},
					arg.Expression,
				))
				return
			}

			if query.IsHasOneModelField(asts, modelField) || query.IsHasManyModelField(asts, modelField) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@sortable does not support ordering of relationships fields",
					},
					arg.Expression,
				))
				return
			}

			if lo.SomeBy(arguments, func(a string) bool { return a == argumentValue }) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("@sortable argument name '%s' already defined", argumentValue),
					},
					arg.Expression,
				))
				return
			}

			arguments = append(arguments, argumentValue)

		},
	}
}
