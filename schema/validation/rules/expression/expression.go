package expression

import (
	"fmt"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ValidateExpressionRule(asts []*parser.AST) []error {
	attributes := query.Attributes(asts)

	for _, attr := range attributes {
		for _, arg := range attr.Arguments {
			if condition, err := expressions.ToAssignmentCondition(arg.Expression); err == nil {
				lhs := condition.LHS.ToString()
				rhs := condition.RHS.ToString()
				fmt.Print(lhs, rhs)

			} else if condition, err := expressions.ToEqualityCondition(arg.Expression); err == nil {
				lhs := condition.LHS.ToString()
				rhs := condition.RHS.ToString()
				fmt.Print(lhs, rhs)

			} else {
				// value only
			}

		}
	}

	return make([]error, 0)
}
