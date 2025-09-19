package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func OrderByTaskAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentTask *parser.TaskNode
	var currentAttribute *parser.AttributeNode
	var argumentLabels []string
	var orderByAttributeDefined bool

	return Visitor{
		EnterTask: func(task *parser.TaskNode) {
			currentTask = task
			orderByAttributeDefined = false
		},
		LeaveModel: func(_ *parser.ModelNode) {
			currentTask = nil
		},

		EnterAttribute: func(attribute *parser.AttributeNode) {
			if currentTask == nil {
				return
			}

			currentAttribute = attribute
			argumentLabels = []string{}

			if attribute.Name.Value != parser.AttributeOrderBy {
				return
			}

			if orderByAttributeDefined {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeNotAllowedError,
					errorhandling.ErrorDetails{
						Message: "@orderBy can only be defined once per task",
					},
					attribute.Name,
				))
			}

			orderByAttributeDefined = true

			if len(attribute.Arguments) == 0 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy requires at least one argument",
					},
					attribute,
				))
			}
		},
		EnterAttributeArgument: func(arg *parser.AttributeArgumentNode) {
			if currentTask == nil {
				return
			}

			if currentAttribute.Name.Value != parser.AttributeOrderBy {
				return
			}

			if arg.Label == nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy arguments must be specified with a label corresponding with a field on this task",
						Hint:    "For example, @orderBy(shipByDate: desc, orderDate: desc)",
					},
					arg,
				))
				return
			}

			field := currentTask.Field(arg.Label.Value)

			if field == nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("@orderBy argument label '%s' must correspond to a field on this model", arg.Label.Value),
					},
					arg.Label,
				))
				return
			}

			if query.IsHasOneModelField(asts, field) || query.IsHasManyModelField(asts, field) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy does not support ordering of relationships fields",
					},
					arg.Label,
				))
				return
			}

			if field.Repeated {
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

			ident, err := resolve.AsIdent(arg.Expression)
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

			if ident == nil || (ident.Fragments[0] != parser.OrderByAscending && ident.Fragments[0] != parser.OrderByDescending) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@orderBy argument value must either be asc or desc",
						Hint:    "For example, @orderBy(surname: asc, firstName: asc)",
					},
					ident,
				))
				return
			}
		},
	}
}
