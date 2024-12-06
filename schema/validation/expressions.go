package validation

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ExpressionRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var action *parser.ActionNode
	var field *parser.FieldNode
	var attribute *parser.AttributeNode
	var job *parser.JobNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m
		},
		LeaveModel: func(*parser.ModelNode) {
			model = nil
		},
		EnterAction: func(a *parser.ActionNode) {
			action = a
		},
		LeaveAction: func(*parser.ActionNode) {
			action = nil
		},
		EnterField: func(f *parser.FieldNode) {
			field = f
		},
		LeaveField: func(*parser.FieldNode) {
			field = nil
		},
		EnterAttribute: func(a *parser.AttributeNode) {
			attribute = a
		},
		LeaveAttribute: func(n *parser.AttributeNode) {
			attribute = nil
		},
		EnterJob: func(j *parser.JobNode) {
			job = j
		},
		LeaveJob: func(*parser.JobNode) {
			job = nil
		},
		EnterAttributeArgument: func(arg *parser.AttributeArgumentNode) {
			var err error
			issues := []expressions.ValidationError{}

			switch attribute.Name.Value {
			case parser.AttributeWhere:
				issues, err = attributes.ValidateWhereExpression(asts, action, arg.Expression)
			case parser.AttributePermission:
				switch arg.Label.Value {
				case "expression":
					issues, err = attributes.ValidatePermissionExpression(asts, model, action, job, arg.Expression)
				case "roles":
					issues, err = attributes.ValidatePermissionRoles(asts, arg.Expression)
				case "actions":
					issues, err = attributes.ValidatePermissionActions(arg.Expression)
				}
			case parser.AttributeDefault:
				issues, err = attributes.ValidateDefaultExpression(asts, field, arg.Expression)
			case parser.AttributeSet:
				l, r, err := arg.Expression.ToAssignmentExpression()
				if err != nil {
					errs.AppendError(makeSetExpressionError(
						errorhandling.AttributeExpressionError,
						"the @set attribute must be an assignment expression",
						fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
						arg.Expression,
					))
					return
				}

				issues, err = attributes.ValidateSetExpression(asts, action, l, r)
			}

			if err != nil {
				panic(err.Error())
			}

			if len(issues) > 0 {
				for _, issue := range issues {
					errs.AppendError(makeWhereExpressionError(
						errorhandling.AttributeExpressionError,
						issue.Message,
						"TODO", // TODO: hints5
						issue.Node,
					))
				}
				return
			}

		},
	}
}

func makeWhereExpressionError(t errorhandling.ErrorType, message string, hint string, node node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		t,
		errorhandling.ErrorDetails{
			Message: message,
			Hint:    hint,
		},
		node,
	)
}
