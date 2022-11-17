package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// Updates the query with all set attributes defined on the operation.
func (query *QueryBuilder) captureSetValues(scope *Scope, args ValueArgs) error {
	for _, setExpression := range scope.operation.SetExpressions {
		expression, err := parser.ParseExpression(setExpression.Source)
		if err != nil {
			return err
		}

		assignment, err := expression.ToAssignmentCondition()
		if err != nil {
			return err
		}

		lhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.LHS)
		rhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.RHS)

		if !(lhsResolver.IsDatabaseColumn() || lhsResolver.IsWriteValue()) {
			return fmt.Errorf("lhs operand of assignment expression must be a model field")
		}

		value, err := rhsResolver.ResolveValue(args, query.writeValues)
		if err != nil {
			return err
		}

		fieldName := assignment.LHS.Ident.Fragments[1].Fragment

		// Add a value to be written during an INSERT or UPDATE
		query.AddWriteValue(fieldName, value)
	}
	return nil
}

// Updates the query with all inputs defined on the operation.
func (query *QueryBuilder) captureWriteValues(scope *Scope, args ValueArgs) error {
	for _, input := range scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		if input.Mode != proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			continue
		}

		// Add a value to be written during an INSERT or UPDATE
		query.AddWriteValue(fieldName, value)
	}
	return nil
}
