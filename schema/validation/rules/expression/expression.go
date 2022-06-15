package expression

import (
	"errors"
	"strings"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ValidateExpressionRule(asts []*parser.AST) []error {
	modelName := "Profile"
	attributes := query.AttributesInModel(asts, modelName)

	for _, attr := range attributes {
		for _, arg := range attr.Arguments {
			condition, err := expressions.ToEqualityCondition(arg.Expression)

			if err != nil {
				// not an equality expression
				continue
			}

			if condition.LHS.Ident != nil {
				lhs, err := checkResolution(asts, modelName, condition.LHS)

			}
			if condition.RHS.Ident != nil {
				rhs, err := checkResolution(asts, modelName, condition.RHS)

			}

		}
	}

	return make([]error, 0)
}

func checkResolution(asts []*parser.AST, contextModel string, value *expressions.Value) (*node.Node, error) {
	errors = make([]error, 0)
	fragments := strings.Split(value.ToString(), ".")

	if fragments[0] != strings.ToLower(contextModel) {
		return nil, errors.New("Does not match model context")
	}

	if value.Ident != nil {
		for _, ast := range asts {
			node, err := ast.ResolveAssociation(contextModel, fragments)

			if err != nil {
				return nil, err
			}

			return node, nil
		}
	}

	return node, nil
}
