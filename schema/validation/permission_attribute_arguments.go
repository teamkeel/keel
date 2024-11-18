package validation

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/rules/expression"
)

func PermissionsAttributeArguments(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var action *parser.ActionNode
	var job *parser.JobNode

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
		EnterJob: func(j *parser.JobNode) {
			job = j
		},
		LeaveJob: func(_ *parser.JobNode) {
			job = nil
		},
		EnterAttribute: func(attr *parser.AttributeNode) {
			if attr.Name.Value != parser.AttributePermission {
				return
			}

			hasActions := false
			hasExpression := false
			hasRoles := false

			for _, arg := range attr.Arguments {
				if arg.Label == nil {
					// Argument validation happens elsewhere
					continue
				}

				switch arg.Label.Value {
				case "actions":
					hasActions = true

					errs.Concat(validateIdentArray(arg.Expression, []string{
						parser.ActionTypeGet,
						parser.ActionTypeCreate,
						parser.ActionTypeUpdate,
						parser.ActionTypeList,
						parser.ActionTypeDelete,
					}, "valid action type"))
				case "expression":
					hasExpression = true

					context := expressions.ExpressionContext{
						Model:     model,
						Attribute: attr,
						Action:    action,
					}
					rules := []expression.Rule{
						expression.OperatorLogicalRule,
					}

					expressionErrors := expression.ValidateExpression(
						asts,
						arg.Expression,
						rules,
						context,
					)
					for _, err := range expressionErrors {
						// TODO: remove cast when expression.ValidateExpression returns correct type
						errs.AppendError(err.(*errorhandling.ValidationError))
					}

					// Extra check for using row-based expression in a read/write function
					// Ideally this would be done as part of the expression validation, but
					// if we don't provide the model as context the error is not very helpful.
					if action != nil && (action.Type.Value == "read" || action.Type.Value == "write") {
						for _, cond := range arg.Expression.Conditions() {
							for _, op := range []*parser.Operand{cond.LHS, cond.RHS} {
								if op == nil || op.Ident == nil {
									continue
								}
								// An ident must have at least one fragment - we only care about the first one
								fragment := op.Ident.Fragments[0]
								if fragment.Fragment == casing.ToLowerCamel(model.Name.Value) {
									errs.AppendError(errorhandling.NewValidationErrorWithDetails(
										errorhandling.AttributeArgumentError,
										errorhandling.ErrorDetails{
											Message: fmt.Sprintf(
												"cannot use row-based permissions in a %s action",
												action.Type.Value,
											),
											Hint: "implement your permissions logic in your function code using the permissions API - https://docs.keel.so/functions#permissions",
										},
										fragment,
									))
								}
							}
						}
					}

				case "roles":
					hasRoles = true

					roles := []string{}
					for _, role := range query.Roles(asts) {
						roles = append(roles, role.Name.Value)
					}

					errs.Concat(validateIdentArray(arg.Expression, roles, "role defined in your schema"))
				}
			}

			// Missing actions argument which is required
			if job == nil && action == nil && !hasActions {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "required argument 'actions' missing",
					},
					attr.Name,
				))
			}

			// One of expression or roles must be provided
			if !hasExpression && !hasRoles {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@permission requires either the 'expressions' or 'roles' argument to be provided",
					},
					attr.Name,
				))
			}
		},
	}
}

func validateIdentArray(expr *parser.Expression, allowed []string, identType string) (errs errorhandling.ValidationErrors) {
	value, err := expr.ToValue()
	if err != nil || value.Array == nil {
		example := ""
		if len(allowed) > 0 {
			example = allowed[0]
		}
		errs.AppendError(errorhandling.NewValidationErrorWithDetails(
			errorhandling.AttributeArgumentError,
			errorhandling.ErrorDetails{
				Message: fmt.Sprintf("value should be a list e.g. [%s]", example),
			},
			expr,
		))
		return
	}

	for _, item := range value.Array.Values {
		valid := item.Ident != nil && lo.Contains(allowed, item.ToString())

		if !valid {
			hint := ""
			if len(allowed) > 0 {
				hint = fmt.Sprintf("valid values are: %s", strings.Join(allowed, ", "))
			}
			errs.AppendError(errorhandling.NewValidationErrorWithDetails(
				errorhandling.AttributeArgumentError,
				errorhandling.ErrorDetails{
					Message: fmt.Sprintf("%s is not a %s", item.ToString(), identType),
					Hint:    hint,
				},
				item,
			))
		}
	}

	return
}
