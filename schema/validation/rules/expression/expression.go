package expression

import (
	"fmt"

	"github.com/teamkeel/keel/schema/associations"
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
				// get all of the nested conditions in the expression
				conditions := arg.Expression.Conditions()

				for _, condition := range conditions {

					// conditionType := condition.Type()
					lhs, _, _ := condition.ToFragments()

					if lhs.Ident != nil {
						tree, err := associations.TryResolveIdent(asts, lhs.Ident)

						if err != nil {
							fmt.Printf("%s (%s)\n", err, tree.PrettyPrint())
						}
					}
				}
			}
		}
	}

	return errs
}
