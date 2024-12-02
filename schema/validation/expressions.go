package validation

import (
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ExpressionRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var action *parser.ActionNode
	var field *parser.FieldNode
	var attribute *parser.AttributeNode

	return Visitor{
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
		EnterAttributeArgument: func(arg *parser.AttributeArgumentNode) {
			var err error
			issues := []expressions.ValidationError{}

			switch attribute.Name.Value {
			case parser.AttributeWhere:
				issues, err = attributes.ValidateWhereExpression(asts, action, arg.Expression)
			case parser.AttributePermission:
				switch arg.Label.Value {
				case "expression":
					issues, err = attributes.ValidatePermissionExpression(asts, action, arg.Expression.String())
				case "roles":
					issues, err = attributes.ValidatePermissionRoles(asts, arg.Expression.String())
				case "actions":
					issues, err = attributes.ValidatePermissionActions(arg.Expression.String())
				}
			case parser.AttributeDefault:
				issues, err = attributes.ValidateDefaultExpression(asts, field, arg.Expression.String())
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
