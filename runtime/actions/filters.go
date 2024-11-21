package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
)

// Applies all implicit input filters to the query.
func (query *QueryBuilder) ApplyImplicitFilters(scope *Scope, args map[string]any) error {
	message := proto.FindWhereInputMessage(scope.Schema, scope.Action.Name)
	if message == nil {
		return nil
	}

	for _, input := range message.Fields {
		if !input.IsModelField() {
			// Skip if this is an explicit input (probably used in a @where)
			continue
		}

		value, ok := args[input.Name]

		if !ok {
			return fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", input.Name, args)
		}

		err := query.whereByImplicitFilter(scope, input.Target, Equals, value)
		if err != nil {
			return err
		}

		// Implicit input filters are ANDed together
		query.And()
	}

	return nil
}

// // Applies all exlicit where attribute filters to the query.
// func (query *QueryBuilder) applyExpressionFilters(scope *Scope, args map[string]any) error {
// 	for _, where := range scope.Action.WhereExpressions {
// 		expression, err := parser.ParseExpression(where.Source)
// 		if err != nil {
// 			return err
// 		}

// 		// Resolve the database statement for this expression
// 		err = query.whereByExpression(scope, expression, args)
// 		if err != nil {
// 			return err
// 		}

// 		// Where attributes are ANDed together
// 		query.And()
// 	}

// 	return nil
// }

// Applies all exlicit where attribute filters to the query.
func (query *QueryBuilder) applyExpressionFiltersWithCel(scope *Scope, args map[string]any) error {
	for _, where := range scope.Action.WhereExpressions {

		err := query.whereByExpression(scope.Context, scope.Schema, scope.Model, scope.Action, where.Source, args)
		if err != nil {
			return err
		}
	}

	return nil
}
