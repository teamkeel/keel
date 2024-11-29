package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
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
				if arg.Label == nil || arg.Label.Value == "" {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: "@permission requires all arguments to be named, for example @permission(roles: [MyRole])",
						},
						arg,
					))
					continue
				}

				switch arg.Label.Value {
				case "actions":
					hasActions = true

					if action != nil || job != nil {
						errs.AppendError(errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf(
									"cannot provide 'actions' arguments when using @permission in %s",
									lo.Ternary(action != nil, "an action", "a job"),
								),
							},
							arg.Label,
						))
						continue
					}

					issues, err := attributes.ValidatePermissionActions(arg.Expression.String())
					if err != nil {
						panic(err.Error())
					}

					if len(issues) > 0 {
						for _, issue := range issues {
							errs.AppendError(errorhandling.NewValidationErrorWithDetails(
								errorhandling.AttributeNotAllowedError,
								errorhandling.ErrorDetails{
									Message: issue,
									Hint:    "",
								},
								arg.Expression,
							))
						}
						return
					}
				case "expression":
					hasExpression = true

					issues, err := attributes.ValidatePermissionExpression(asts, action, arg.Expression.String())
					if err != nil {
						panic(err.Error())
					}

					if len(issues) > 0 {
						for _, issue := range issues {
							errs.AppendError(errorhandling.NewValidationErrorWithDetails(
								errorhandling.AttributeNotAllowedError,
								errorhandling.ErrorDetails{
									Message: issue,
									Hint:    "",
								},
								arg.Expression,
							))
						}
						return
					}

					// Extra check for using row-based expression in a read/write function
					// Ideally this would be done as part of the expression validation, but
					// if we don't provide the model as context the error is not very helpful.
					if action != nil && (action.Type.Value == "read" || action.Type.Value == "write") {

						operands, err := resolve.IdentOperands(arg.Expression.String())
						if err != nil {
							return
						}

						for _, op := range operands {
							// An ident must have at least one fragment - we only care about the first one
							fragment := op[0]
							if fragment == casing.ToLowerCamel(model.Name.Value) {
								errs.AppendError(errorhandling.NewValidationErrorWithDetails(
									errorhandling.AttributeArgumentError,
									errorhandling.ErrorDetails{
										Message: fmt.Sprintf(
											"cannot use row-based permissions in a %s action",
											action.Type.Value,
										),
										Hint: "implement your permissions logic in your function code using the permissions API - https://docs.keel.so/functions#permissions",
									},
									arg.Expression,
								))
							}
						}
					}

				case "roles":
					hasRoles = true

					issues, err := attributes.ValidatePermissionRole(asts, arg.Expression.String())
					if err != nil {
						panic(err.Error())
					}

					if len(issues) > 0 {
						for _, issue := range issues {
							errs.AppendError(errorhandling.NewValidationErrorWithDetails(
								errorhandling.AttributeNotAllowedError,
								errorhandling.ErrorDetails{
									Message: issue,
									Hint:    "",
								},
								arg.Expression,
							))
						}
						return
					}
				default:
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf(
								"'%s' is not a valid argument for @permission",
								arg.Label.Value,
							),
							Hint: "Did you mean one of 'actions', 'expression', or 'roles'?",
						},
						arg.Label,
					))
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
