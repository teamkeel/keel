package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// Applies all implicit input filters to the query.
func (query *QueryBuilder) applyImplicitFilters(scope *Scope, args WhereArgs) error {
	for _, input := range scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT || input.Mode == proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Name
		value, ok := args[fieldName]

		if !ok {
			return fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", fieldName, args)
		}

		err := query.whereByImplicitFilter(scope, input, fieldName, Equals, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// Applies all exlicit where attribute filters to the query.
func (query *QueryBuilder) applyExplicitFilters(scope *Scope, args WhereArgs) error {
	for _, where := range scope.operation.WhereExpressions {
		expression, err := parser.ParseExpression(where.Source)
		if err != nil {
			return err
		}

		// Resolve the database statement for this expression
		err = query.whereByExpression(scope, expression, args)
		if err != nil {
			return err
		}
	}

	return nil
}
