package actions

import (
	"fmt"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
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

// Include a filter (where condition) on the query based on an implicit input filter.
func (query *QueryBuilder) whereByImplicitFilter(scope *Scope, targetField []string, operator ActionOperator, value any) error {
	// Implicit inputs don't include the base model as the first fragment (unlike expressions), so we include it
	fragments := append([]string{casing.ToLowerCamel(scope.Action.ModelName)}, targetField...)

	// The lhs QueryOperand is determined from the fragments in the implicit input field
	left, err := operandFromFragments(scope.Schema, fragments)
	if err != nil {
		return err
	}

	// The rhs QueryOperand is always a value in an implicit input
	right := Value(value)

	// Add join for the implicit input
	err = query.AddJoinFromFragments(scope.Schema, fragments)
	if err != nil {
		return err
	}

	// Add where condition to the query for the implicit input
	err = query.Where(left, operator, right)
	if err != nil {
		return err
	}

	return nil
}

// Applies all exlicit where attribute filters to the query.
func (query *QueryBuilder) applyExpressionFilters(scope *Scope, args map[string]any) error {
	for _, where := range scope.Action.WhereExpressions {
		expression, err := parser.ParseExpression(where.Source)
		if err != nil {
			return err
		}

		_, err = resolve.RunCelVisitor(expression, GenerateFilterQuery(scope.Context, query, scope.Schema, scope.Model, scope.Action, args))
		if err != nil {
			return err
		}

		query.And()
	}

	return nil
}
