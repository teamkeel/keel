package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// DefaultApplyImplicitFilters considers all the implicit inputs expected for
// the given operation, and captures the targeted field. It then captures the corresponding value
// operand value provided by the given request arguments, and adds a Where clause to the
// query field in the given scope, using a hard-coded equality operator.
func DefaultApplyImplicitFilters(scope *Scope, args WhereArgs) error {
	for _, input := range scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT || input.Mode == proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			return fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", fieldName, args)
		}

		// New filter resolver to generate a database query statement
		resolver := NewFilterResolver(scope)

		// Resolve the database statement for this expression
		statement, err := resolver.ResolveQueryStatement(fieldName, value, Equals, input.Type.Type)
		if err != nil {
			return err
		}

		// Logical AND between all the expressions
		scope.query = scope.query.Where(statement)
	}

	return nil
}

func DefaultApplyExplicitFilters(scope *Scope, args WhereArgs) error {
	operation := scope.operation

	for _, where := range operation.WhereExpressions {
		expression, err := parser.ParseExpression(where.Source) // E.g. post.title == requiredTitle
		if err != nil {
			return err
		}

		// New expression resolver to generate a database query statement
		resolver := NewExpressionResolver(scope)

		// Resolve the database statement for this expression
		statement, err := resolver.ResolveQueryStatement(expression, args, scope.writeValues)
		if err != nil {
			return err
		}

		// Logical AND between all the expressions
		scope.query = scope.query.Where(statement)
	}

	return nil
}
