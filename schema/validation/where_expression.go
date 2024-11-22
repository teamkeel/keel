package validation

import (
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func WhereAttributeExpressionRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	//var model *parser.ModelNode
	var action *parser.ActionNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			//model = m
		},
		LeaveModel: func(_ *parser.ModelNode) {
			//model = nil
		},
		EnterAction: func(a *parser.ActionNode) {
			action = a
		},
		LeaveAction: func(_ *parser.ActionNode) {
			action = nil
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			if attribute == nil || attribute.Name.Value != parser.AttributeWhere {
				return
			}

			expr := attribute.Arguments[0].Expression

			p, err := attributes.NewWhereExpressionParser(asts, action)
			if err != nil {
				panic(err.Error())
			}
			issues, err := p.Validate(expr.String())
			if err != nil {
				panic(err.Error())
			}

			if len(issues) > 0 {
				for _, issue := range issues {
					errs.AppendError(makeWhereExpressionError(
						errorhandling.AttributeExpressionError,
						issue,
						"TODO", // TODO: hints
						expr.AstNode(),
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
