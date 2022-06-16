package expression

import (
	"strings"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
)

func ValidateExpressionRule(ast *parser.AST) []error {
	errs := make([]error, 0)

	for _, model := range ast.Models() {
		attrs := model.Attributes()

		if attrs != nil {
			for _, attr := range attrs {
				for _, arg := range attr.Arguments {
					condition, err := expressions.ToEqualityCondition(arg.Expression)

					if err != nil {
						// this is not an equality expression
						continue
					}

					if condition.LHS.Ident != nil {
						_, err := checkResolution(ast, model, condition.LHS)

						if err != nil {
							errs = append(errs, err)
						}

					}
					if condition.RHS.Ident != nil {
						_, err := checkResolution(ast, model, condition.RHS)

						if err != nil {
							errs = append(errs, err)
						}
					}
				}
			}
		}
	}

	return errs
}

func checkResolution(ast *parser.AST, contextModel *parser.ModelNode, value *expressions.Value) (*node.Node, error) {
	if value.Ident != nil {
		fragments := strings.Split(value.ToString(), ".")

		n, err := ast.ResolveAssociation(contextModel, fragments)

		if err != nil {
			return nil, err
		}

		return n, nil
	}

	return nil, nil
}
