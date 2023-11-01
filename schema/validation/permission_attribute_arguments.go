package validation

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/rules/expression"
)

// Validates the arguments to any attribute expression.
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
		EnterAttribute: func(attribute *parser.AttributeNode) {
			if attribute.Name.Value != parser.AttributePermission {
				return
			}

			errors := validatePermissionAttribute(asts, attribute, model, action, job)
			errs.Concat(errors)
		},
	}
}

func validatePermissionAttribute(asts []*parser.AST, attr *parser.AttributeNode, model *parser.ModelNode, action *parser.ActionNode, job *parser.JobNode) (errs errorhandling.ValidationErrors) {
	hasActions := false
	hasExpression := false
	hasRoles := false
	errs = errorhandling.ValidationErrors{}

	for _, arg := range attr.Arguments {
		if arg.Label == nil || arg.Label.Value == "" {
			// All arguments to @permission should have a label
			errs.Append(errorhandling.ErrorAttributeRequiresNamedArguments,
				map[string]string{
					"AttributeName":      "permission",
					"ValidArgumentNames": "'actions', 'expression', or 'roles'",
				},
				arg,
			)
			continue
		}

		switch arg.Label.Value {
		case "actions":
			// The 'actions' argument should not be provided if the permission attribute
			// is defined inside an actionn as that implicitly means the permission only
			// applies to that action. It is also not valid in a job.
			if model != nil {
				allowedIdents := append([]string{}, validActionKeywords...)
				errs.Concat(validateIdentArray(arg.Expression, allowedIdents))
			} else {
				errs.Append(errorhandling.ErrorInvalidAttributeArgument,
					map[string]string{
						"AttributeName": "permission",
						"ArgumentName":  "actions",
						"Location":      "action",
					},
					arg,
				)
			}
			hasActions = true
		case "expression":
			hasExpression = true

			var context expressions.ExpressionContext
			var rules []expression.Rule
			switch {
			case action != nil || model != nil:
				rules = []expression.Rule{
					expression.OperatorLogicalRule,
				}

				context = expressions.ExpressionContext{
					Model:     model,
					Attribute: attr,
					Action:    action,
				}
			case job != nil:
				rules = []expression.Rule{

					expression.OperatorLogicalRule,
				}

				context = expressions.ExpressionContext{
					Attribute: attr,
				}
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
		case "roles":
			hasRoles = true
			allowedIdents := []string{}
			for _, role := range query.Roles(asts) {
				allowedIdents = append(allowedIdents, role.Name.Value)
			}
			errs.Concat(validateIdentArray(arg.Expression, allowedIdents))
		default:
			// Unknown argument
			errs.Append(errorhandling.ErrorInvalidAttributeArgument,
				map[string]string{
					"AttributeName":      "permission",
					"ArgumentName":       arg.Label.Value,
					"ValidArgumentNames": "'actions', 'expression', or 'roles'",
				},
				arg.Label,
			)
		}
	}

	// Missing actions argument which is required
	if job == nil && action == nil && !hasActions {
		errs.Append(errorhandling.ErrorAttributeMissingRequiredArgument,
			map[string]string{
				"AttributeName": "permission",
				"ArgumentName":  "actions",
			},
			attr.Name,
		)
	}

	// One of expression or roles must be provided
	if !hasExpression && !hasRoles {
		errs.Append(errorhandling.ErrorAttributeMissingRequiredArgument,
			map[string]string{
				"AttributeName": "permission",
				"ArgumentName":  `"expression" or "roles"`,
			},
			attr.Name,
		)
	}

	return
}

func validateIdentArray(expr *parser.Expression, allowedIdents []string) (errs errorhandling.ValidationErrors) {
	value, err := expr.ToValue()
	if err != nil || value.Array == nil {
		expected := ""
		if len(allowedIdents) > 0 {
			expected = "an array containing any of the following identifiers - " + formatting.HumanizeList(allowedIdents, formatting.DelimiterOr)
		}
		// Check expression is an array
		errs.Append(errorhandling.ErrorInvalidValue,
			map[string]string{
				"Expected": expected,
			},
			expr,
		)
		return
	}

	for _, item := range value.Array.Values {
		// Each item should be a singular ident e.g. "foo" and not "foo.baz.bop"
		// String literal idents e.g ["thisisinvalid"] are assumed not to be invalid
		valid := false

		if item.Ident != nil {
			valid = len(item.Ident.Fragments) == 1
		}

		if valid {
			// If it is a single ident check it's an allowed value
			name := item.Ident.Fragments[0].Fragment
			valid = lo.Contains(allowedIdents, name)
		}

		if !valid {
			expected := ""
			if len(allowedIdents) > 0 {
				expected = "any of the following identifiers - " + formatting.HumanizeList(allowedIdents, formatting.DelimiterOr)
			}
			errs.Append(errorhandling.ErrorInvalidValue,

				map[string]string{
					"Expected": expected,
				},

				item,
			)
		}
	}

	return
}

var validActionKeywords = []string{
	parser.ActionTypeGet,
	parser.ActionTypeCreate,
	parser.ActionTypeUpdate,
	parser.ActionTypeList,
	parser.ActionTypeDelete,
}
