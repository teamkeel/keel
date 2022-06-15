package expression

import (
	"errors"
	"strings"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ValidateExpressionRule(ast *parser.AST) []error {
	//todo replace with loop
	modelName := "Profile"
	attributes := query.AttributesInModel(ast, modelName)

	for _, attr := range attributes {
		for _, arg := range attr.Arguments {
			condition, err := expressions.ToEqualityCondition(arg.Expression)

			if err != nil {
				// this is not an equality expression
				continue
			}

			if condition.LHS.Ident != nil {
				lhs, err := checkResolution(ast, modelName, condition.LHS)

			}
			if condition.RHS.Ident != nil {
				rhs, err := checkResolution(ast, modelName, condition.RHS)

			}

		}
	}

	return make([]error, 0)
}

func checkResolution(ast *parser.AST, contextModel string, value *expressions.Value) (*node.Node, error) {
	fragments := strings.Split(value.ToString(), ".")

	if fragments[0] != strings.ToLower(contextModel) {

		// e.g model is Profile
		// but expression is something.else == 123 where something should be profile (lowercased)
		return nil, errors.New("Does not match model context")
	}

	if value.Ident != nil {
		n, err := ast.ResolveAssociation(contextModel, fragments)

		if err != nil {
			return nil, err
		}

		return n, nil
	}

	return nil, nil
}
