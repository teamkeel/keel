package expressions

import (
	"context"
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// TryResolveExpressionEarly attempts to evaluate the expression in the runtime process without generating a row-based query against the database.
func TryResolveExpressionEarly(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, expression *parser.Expression, args map[string]any) (canResolveInMemory bool, resolvedValue bool) {
	can := false
	value := false

	for _, or := range expression.Or {
		currCanResolve := true
		currExpressionValue := true

		for _, and := range or.And {
			if and.Expression != nil {
				currCan, currValue := TryResolveExpressionEarly(ctx, schema, model, action, and.Expression, args)
				currCanResolve = currCan && currCanResolve
				currExpressionValue = currExpressionValue && currValue
			}

			if and.Condition != nil {
				currCan, currValue := resolveConditionEarly(ctx, schema, model, action, and.Condition, args)
				currCanResolve = currCan && currCanResolve
				currExpressionValue = currExpressionValue && currValue
			}
		}

		can = can || currCanResolve
		value = value || currExpressionValue
	}

	return can, value
}

// canResolveConditionEarly determines if a single condition can be resolved in the process without generating a row-based query against the database.
func canResolveConditionEarly(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, condition *parser.Condition) bool {

	// XXXX take this out
	if action.Name == "getFilm" {
		a := 1
		_ = a
	}

	lhsResolver := NewOperandResolver(ctx, schema, model, action, condition.LHS)

	if condition.Type() == parser.ValueCondition {
		// XXXXX remove this intermediate variable.
		isDbColumn := lhsResolver.IsDatabaseColumn()
		_ = isDbColumn

		return !lhsResolver.IsDatabaseColumn()
	}

	rhsResolver := NewOperandResolver(ctx, schema, model, action, condition.RHS)
	referencesDatabaseColumns := lhsResolver.IsDatabaseColumn() || rhsResolver.IsDatabaseColumn()

	return !(referencesDatabaseColumns)
}

// resolveConditionEarly resolves a single condition in the process without generating a row-based query against the database.
func resolveConditionEarly(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, condition *parser.Condition, args map[string]any) (canResolveEarly bool, resolvedValue bool) {
	if !canResolveConditionEarly(ctx, schema, model, action, condition) {
		return false, false
	}

	lhsResolver := NewOperandResolver(ctx, schema, model, action, condition.LHS)
	operandType, _ := lhsResolver.GetOperandType()
	lhsValue, _ := lhsResolver.ResolveValue(args)

	if condition.Type() == parser.ValueCondition {
		result, _ := evaluate(lhsValue, true, operandType, &parser.Operator{Symbol: parser.OperatorEquals})
		return true, result
	}

	rhsResolver := NewOperandResolver(ctx, schema, model, action, condition.RHS)
	rhsValue, _ := rhsResolver.ResolveValue(args)
	result, _ := evaluate(lhsValue, rhsValue, operandType, condition.Operator)

	return true, result
}

// Evaluate lhs and rhs with an operator in this process.
func evaluate(
	lhs any,
	rhs any,
	operandType proto.Type,
	operator *parser.Operator,
) (bool, error) {
	// Evaluate when either operand or both are nil
	if lhs == nil && rhs == nil {
		return true && (operator.Symbol != parser.OperatorNotEquals), nil
	} else if lhs == nil || rhs == nil {
		return false || (operator.Symbol == parser.OperatorNotEquals), nil
	}

	// Evaluate with non-nil operands
	switch operandType {
	case proto.Type_TYPE_STRING:
		return compareString(lhs.(string), rhs.(string), operator)
	case proto.Type_TYPE_INT:
		// todo: unify these to a single type at the source?
		switch v := lhs.(type) {
		case int:
			// Sourced from GraphQL input parameters.
			lhs = int64(v)
		case float64:
			// Sourced from integration test framework.
			lhs = int64(v)
		case int32:
			// Sourced from database.
			lhs = int64(v) // todo: https://linear.app/keel/issue/RUN-98/number-type-as-int32-or-int64
		}
		switch v := rhs.(type) {
		case int:
			// Sourced from GraphQL input parameters.
			rhs = int64(v)
		case float64:
			// Sourced from integration test framework.
			rhs = int64(v)
		case int32:
			// Sourced from database.
			rhs = int64(v) // todo: https://linear.app/keel/issue/RUN-98/number-type-as-int32-or-int64
		}
		return compareInt(lhs.(int64), rhs.(int64), operator)
	case proto.Type_TYPE_BOOL:
		return compareBool(lhs.(bool), rhs.(bool), operator)
	case proto.Type_TYPE_ENUM:
		return compareEnum(lhs.(string), rhs.(string), operator)
	case proto.Type_TYPE_ID, proto.Type_TYPE_MODEL:
		return compareIdentity(lhs.(ksuid.KSUID), rhs.(ksuid.KSUID), operator)
	default:
		return false, fmt.Errorf("cannot yet handle comparision of type: %s", operandType)
	}
}

func compareString(
	lhs string,
	rhs string,
	operator *parser.Operator,
) (bool, error) {
	switch operator.Symbol {
	case parser.OperatorEquals:
		return lhs == rhs, nil
	case parser.OperatorNotEquals:
		return lhs != rhs, nil
	default:
		return false, fmt.Errorf("operator: %s, not supported for type: %s", operator.Symbol, proto.Type_TYPE_STRING)
	}
}

func compareInt(
	lhs int64,
	rhs int64,
	operator *parser.Operator,
) (bool, error) {
	switch operator.Symbol {
	case parser.OperatorEquals:
		return lhs == rhs, nil
	case parser.OperatorNotEquals:
		return lhs != rhs, nil
	case parser.OperatorGreaterThan:
		return lhs > rhs, nil
	case parser.OperatorGreaterThanOrEqualTo:
		return lhs >= rhs, nil
	case parser.OperatorLessThan:
		return lhs < rhs, nil
	case parser.OperatorLessThanOrEqualTo:
		return lhs <= rhs, nil
	default:
		return false, fmt.Errorf("operator: %s, not supported for type: %s", operator.Symbol, proto.Type_TYPE_INT)
	}
}

func compareBool(
	lhs bool,
	rhs bool,
	operator *parser.Operator,
) (bool, error) {
	switch operator.Symbol {
	case parser.OperatorEquals:
		return lhs == rhs, nil
	case parser.OperatorNotEquals:
		return lhs != rhs, nil
	default:
		return false, fmt.Errorf("operator: %s, not supported for type: %s", operator.Symbol, proto.Type_TYPE_BOOL)
	}
}

func compareEnum(
	lhs string,
	rhs string,
	operator *parser.Operator,
) (bool, error) {
	switch operator.Symbol {
	case parser.OperatorEquals:
		return lhs == rhs, nil
	case parser.OperatorNotEquals:
		return lhs != rhs, nil
	default:
		return false, fmt.Errorf("operator: %s, not supported for type: %s", operator.Symbol, proto.Type_TYPE_STRING)
	}
}

func compareIdentity(
	lhs ksuid.KSUID,
	rhs ksuid.KSUID,
	operator *parser.Operator,
) (bool, error) {
	switch operator.Symbol {
	case parser.OperatorEquals:
		return lhs == rhs, nil
	case parser.OperatorNotEquals:
		return lhs != rhs, nil
	default:
		return false, fmt.Errorf("operator: %s, not supported for type: %s", operator.Symbol, proto.Type_TYPE_ID)
	}
}
