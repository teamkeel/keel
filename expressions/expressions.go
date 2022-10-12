package expressions

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ResolveCondition(asts []*parser.AST, c *parser.Condition, context ExpressionContext) (resolvedLhs *ExpressionScopeEntity, resolvedRhs *ExpressionScopeEntity, errors []error) {
	lhs := c.LHS
	rhs := c.RHS

	scope := &ExpressionScope{
		Entities: []*ExpressionScopeEntity{
			{
				Name:  strcase.ToLowerCamel(context.Model.Name.Value),
				Model: context.Model,
			},
		},
	}

	if context.Action != nil {
		// todo: this isnt right
		// the scope logic for inputs should be:
		// if lhs, suggest read and write ONLY for @permission expression, otherwise, dont suggest anything
		// if rhs, suggest write inputs only

		for _, input := range context.Action.AllInputs() {
			// inputs using short-hand syntax that refer to relationships
			// don't get added to the scope
			if input.Label == nil && len(input.Type.Fragments) > 1 {
				continue
			}

			resolvedType := query.ResolveInputType(asts, input, context.Model)
			if resolvedType == "" {
				continue
			}
			scope.Entities = append(scope.Entities, &ExpressionScopeEntity{
				Name: input.Name(),
				Type: resolvedType,
			})
		}
	}

	scope = DefaultExpressionScope(asts).Merge(scope)

	resolvedLhs, lhsErr := ResolveOperand(asts, lhs, scope, OperandPositionLhs)

	if lhsErr != nil {
		errors = append(errors, lhsErr)
	}

	if rhs != nil {
		resolvedRhs, rhsErr := ResolveOperand(asts, rhs, scope, OperandPositionRhs)

		if rhsErr != nil {
			errors = append(errors, rhsErr)
		}

		return resolvedLhs, resolvedRhs, errors
	}

	return resolvedLhs, nil, errors
}
