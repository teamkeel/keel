package validation

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var supportedActionTypes = []string{
	parser.ActionTypeCreate,
	parser.ActionTypeDelete,
	parser.ActionTypeUpdate,
}

func OnAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentAttribute *parser.AttributeNode
	var arguments []*parser.AttributeArgumentNode

	return Visitor{
		EnterAttribute: func(attribute *parser.AttributeNode) {
			if attribute.Name.Value != parser.AttributeOn {
				return
			}

			currentAttribute = attribute
			arguments = []*parser.AttributeArgumentNode{}

			if len(attribute.Arguments) < 2 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@on requires two arguments - an array of action types and a subscriber name",
						Hint:    "For example, @on([create, update], verifyEmailAddress)",
					},
					attribute.Name,
				))
				return
			}
		},
		LeaveAttribute: func(n *parser.AttributeNode) {
			currentAttribute = nil
		},
		EnterAttributeArgument: func(arg *parser.AttributeArgumentNode) {
			if currentAttribute == nil {
				return
			}

			arguments = append(arguments, arg)

			if arg.Label != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@on does not support or require named arguments",
						Hint:    "For example, @on([create, update], verifyEmailAddress)",
					},
					arg,
				))
				return
			}

			// Rules for the first argument (the action types array)
			if len(arguments) == 1 {
				operands, err := resolve.AsIdentArray(arg.Expression)
				if err != nil {
					errs.AppendError(actionTypesNonArrayError(arg))
					return
				}

				for _, element := range operands {
					if len(element.Fragments) != 1 {
						errs.AppendError(errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("@on only supports the following action types: %s", strings.Join(supportedActionTypes, ", ")),
								Hint:    "For example, @on([create, update], verifyEmailAddress)",
							},
							element,
						))
						continue
					}

					if !lo.Contains(supportedActionTypes, element.Fragments[0]) {
						errs.AppendError(errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("@on only supports the following action types: %s", strings.Join(supportedActionTypes, ", ")),
								Hint:    "For example, @on([create, update], verifyEmailAddress)",
							},
							element,
						))
					}
				}
			}

			// Rules for the second argument (the subscriber name)
			if len(arguments) == 2 {
				ident, err := resolve.AsIdent(arg.Expression)
				if err != nil {
					errs.AppendError(subscriberNameInvalidError(arg))
					return
				}

				if len(ident.Fragments) != 1 {
					errs.AppendError(subscriberNameInvalidError(ident))
					return
				}

				name := ident.ToString()
				if name != strcase.ToLowerCamel(name) {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: "a valid function name must be in lower camel case",
							Hint:    fmt.Sprintf("Try use '%s'", strcase.ToLowerCamel(name)),
						},
						ident,
					))
					return
				}
			}

			if len(arguments) > 2 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@on only takes two arguments",
						Hint:    "For example, @on([create, update], verifyEmailAddress)",
					},
					arg,
				))
			}
		},
	}
}

func actionTypesNonArrayError(position node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.AttributeArgumentError,
		errorhandling.ErrorDetails{
			Message: "@on argument must be an array of action types",
			Hint:    "For example, @on([create, update], verifyEmailAddress)",
		},
		position)
}

func subscriberNameInvalidError(position node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.AttributeArgumentError,
		errorhandling.ErrorDetails{
			Message: "@on subscriber argument must be a valid function name",
			Hint:    "For example, @on([create, update], verifyEmailAddress)",
		},
		position)
}
