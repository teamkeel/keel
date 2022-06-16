package expression

import (
	"strings"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

type ResolvedValue struct {
	*node.Node

	Type string
}

func ValidateExpressionRule(asts []*parser.AST) []error {
	errs := make([]error, 0)

	for _, model := range query.Models(asts) {
		attrs := query.ModelAttributes(model)

		for _, attr := range attrs {
			for _, arg := range attr.Arguments {
				condition, err := expressions.ToEqualityCondition(arg.Expression)

				if err != nil {
					// it is not an equality expression, so we are not interested
					continue
				}

				// Example: a full condition as a string could be: "a.b.c == c.b.a"

				// Check left hand side (a.b.c) of conditional to try to resolve it
				_, err = checkExpressionConditionSide(asts, model, condition.LHS)
				if err != nil {
					errs = append(errs, err)
				}

				// Check right hand side (c.b.a) of conditional to try to resolve it
				_, err = checkExpressionConditionSide(asts, model, condition.RHS)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	return errs
}

func checkExpressionConditionSide(ast *parser.AST, contextModel *parser.ModelNode, value *expressions.Value) (*ResolvedValue, error) {
	if value.Ident != nil {
		fragments := strings.Split(value.ToString(), ".")

		// Handle special case where an ident refers to the ctx object, which is not a model.
		if fragments[0] == "ctx" {
			return &ResolvedValue{
				Type: "ctx",
			}, nil
		}

		// todo: check levenstein distance for ctx (e.g user writes context) and return suggestion hint

		// Try to resolve the association based on the contextModel
		// e.g contextModel will be "modelName" in the path fragment modelName.associationA.associationB
		v, err := tryAssociation(ast, contextModel, fragments)

		if err != nil {
			return v, nil
		}
	}

	return &ResolvedValue{
		Type: value.Type(),
	}, nil
}

func tryAssociation(ast *parser.AST, contextModel *parser.ModelNode, fragments []string) (*ResolvedValue, error) {
	// n, err := query.ResolveAssociation(contextModel, fragments)

	// if err == nil {
	// 	return &ResolvedValue{
	// 		Node: n,
	// 		Type: "association",
	// 	}, nil
	// }

	// return nil, err

	return nil, nil
}
