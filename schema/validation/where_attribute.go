package validation

import (
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func WhereAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var action *parser.ActionNode
	var attribute *parser.AttributeNode

	return Visitor{
		EnterAction: func(a *parser.ActionNode) {
			action = a
		},
		LeaveAction: func(*parser.ActionNode) {
			action = nil
		},
		EnterAttribute: func(a *parser.AttributeNode) {
			attribute = a
		},
		LeaveAttribute: func(*parser.AttributeNode) {
			attribute = nil
		},
		EnterExpression: func(e *parser.Expression) {
			if attribute.Name.Value != parser.AttributeWhere {
				return
			}

			issues, err := attributes.ValidateWhereExpression(asts, action, e)
			if err != nil {
				panic(err.Error())
			}

			if len(issues) > 0 {
				for _, issue := range issues {
					errs.AppendError(issue)
				}
			}
		},
	}
}
