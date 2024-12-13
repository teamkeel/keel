package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func OrderByAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	var currentOperation *parser.ActionNode
	var currentAttribute *parser.AttributeNode
	var orderByAttributeDefined bool
	var argumentLabels []string

	return Visitor{
		EnterModel: func(model *parser.ModelNode) {
			currentModel = model
		},
		LeaveModel: func(_ *parser.ModelNode) {
			currentModel = nil
		},
		EnterAction: func(action *parser.ActionNode) {
			currentOperation = action
			orderByAttributeDefined = false
		},
		LeaveAction: func(_ *parser.ActionNode) {
			currentOperation = nil
			orderByAttributeDefined = false
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			currentAttribute = attribute
			argumentLabels = []string{}

			if attribute.Name.Value != parser.AttributeOrderBy {
				return
			}

			if currentOperation == nil {
				return
			}

			if currentOperation.Type.Value != parser.ActionTypeList {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@orderBy can only be used on list actions",
					},
					attribute.Name,
				))
			}

			if orderByAttributeDefined {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@orderBy can only be defined once per action",
					},
					attribute.Name,
				))
			}

			orderByAttributeDefined = true

			if len(attribute.Arguments) == 0 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy requires at least once argument",
					},
					attribute,
				))
			}
		},
		EnterAttributeArgument: func(arg *parser.AttributeArgumentNode) {
			if currentAttribute.Name.Value != parser.AttributeOrderBy {
				return
			}

			if arg.Label == nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy arguments must be specified with a label corresponding with a field on this model",
						Hint:    "For example, @orderBy(surname: asc, firstName: asc)",
					},
					arg,
				))
				return
			}

			modelField := query.ModelField(currentModel, arg.Label.Value)

			if modelField == nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("@orderBy argument label '%s' must correspond to a field on this model", arg.Label.Value),
					},
					arg.Label,
				))
				return
			}

			if query.IsHasOneModelField(asts, modelField) || query.IsHasManyModelField(asts, modelField) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy does not support ordering of relationships fields",
					},
					arg.Label,
				))
				return
			}

			if modelField.Repeated {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy does not support ordering of array fields",
					},
					arg.Label,
				))
				return
			}

			if lo.SomeBy(argumentLabels, func(a string) bool { return a == arg.Label.Value }) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("@orderBy argument name '%s' already defined", arg.Label.Value),
					},
					arg.Label,
				))
				return
			}

			argumentLabels = append(argumentLabels, arg.Label.Value)

			ident, err := resolve.AsIdent(arg.Expression.String())
			if err != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy argument value must either be asc or desc",
						Hint:    "For example, @orderBy(surname: asc, firstName: asc)",
					},
					arg.Expression,
				))
				return
			}

			if ident == nil || (ident[0] != parser.OrderByAscending && ident[0] != parser.OrderByDescending) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy argument value must either be asc or desc",
						Hint:    "For example, @orderBy(surname: asc, firstName: asc)",
					},
					arg.Expression,
				))
				return
			}
		},
	}
}
