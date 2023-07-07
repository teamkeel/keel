package expressions

import (
	"context"
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// // Determines if the expression can be evaluated on the runtime process
// // as opposed to producing a SQL statement and querying against the database.
// func CanResolveInMemory(ctx context.Context, schema *proto.Schema, model *proto.Model, operation *proto.Operation, expression *parser.Expression) bool {
// 	can := false

// 	for _, or := range expression.Or {
// 		currExpressionCan := true

// 		for _, and := range or.And {
// 			if and.Expression != nil {
// 				currExpressionCan = currExpressionCan && CanResolveInMemory(ctx, schema, model, operation, and.Expression)
// 			}

// 			if and.Condition != nil {
// 				currExpressionCan = currExpressionCan && canResolveConditionInMemory(ctx, schema, model, operation, and.Condition)
// 			}
// 		}

// 		can = can || currExpressionCan
// 	}

// 	return can
// }

func canResolveConditionInMemory(ctx context.Context, schema *proto.Schema, model *proto.Model, operation *proto.Operation, condition *parser.Condition) bool {
	lhsResolver := NewOperandResolver(ctx, schema, model, operation, condition.LHS)

	if condition.Type() == parser.ValueCondition {
		return !lhsResolver.IsDatabaseColumn()
	}

	rhsResolver := NewOperandResolver(ctx, schema, model, operation, condition.RHS)
	referencesDatabaseColumns := lhsResolver.IsDatabaseColumn() || rhsResolver.IsDatabaseColumn()

	return !(referencesDatabaseColumns)
}

// Evaluated the expression in the runtime process without generated a query against the database.
func ResolveInMemory(ctx context.Context, schema *proto.Schema, model *proto.Model, operation *proto.Operation, expression *parser.Expression, args map[string]any) (canResolveInMemory bool, resolvedValue bool) {
	can := false
	value := false

	for _, or := range expression.Or {
		currCanResolve := true
		currExpressionValue := true

		for _, and := range or.And {
			if and.Expression != nil {
				currCan, currValue := ResolveInMemory(ctx, schema, model, operation, and.Expression, args)
				currCanResolve = currCan && currCanResolve
				currExpressionValue = currExpressionValue && currValue
			}

			if and.Condition != nil {
				currCan, currValue := resolveConditionInMemory(ctx, schema, model, operation, and.Condition, args)
				currCanResolve = currCan && currCanResolve
				currExpressionValue = currExpressionValue && currValue
			}
		}

		can = can || currCanResolve
		value = value || currExpressionValue
	}

	return can, value
}

func resolveConditionInMemory(ctx context.Context, schema *proto.Schema, model *proto.Model, operation *proto.Operation, condition *parser.Condition, args map[string]any) (canResolveInMemory bool, resolvedValue bool) {
	if !canResolveConditionInMemory(ctx, schema, model, operation, condition) {
		return false, false
	}

	lhsResolver := NewOperandResolver(ctx, schema, model, operation, condition.LHS)
	operandType, _ := lhsResolver.GetOperandType()
	lhsValue, _ := lhsResolver.ResolveValue(args)

	if condition.Type() == parser.ValueCondition {
		result, _ := evaluate(lhsValue, true, operandType, &parser.Operator{Symbol: parser.OperatorEquals})
		return true, result
	}

	rhsResolver := NewOperandResolver(ctx, schema, model, operation, condition.RHS)
	rhsValue, _ := rhsResolver.ResolveValue(args)
	result, _ := evaluate(lhsValue, rhsValue, operandType, condition.Operator)

	return true, result
}

// // Evaluated the expression in the runtime process without generated a query against the database.
// func ResolveInMemory(ctx context.Context, schema *proto.Schema, model *proto.Model, operation *proto.Operation, expression *parser.Expression, args map[string]any) bool {
// 	// We don't yet support running multiple conditions in memory
// 	if len(expression.Conditions()) > 1 {
// 		return false
// 	}

// 	condition := expression.Conditions()[0]

// 	lhsResolver := NewOperandResolver(ctx, schema, model, operation, condition.LHS)
// 	operandType, _ := lhsResolver.GetOperandType()
// 	lhsValue, _ := lhsResolver.ResolveValue(args)

// 	if condition.Type() == parser.ValueCondition {
// 		result, _ := evaluate(lhsValue, true, operandType, &parser.Operator{Symbol: parser.OperatorEquals})
// 		return result
// 	}

// 	rhsResolver := NewOperandResolver(ctx, schema, model, operation, condition.RHS)

// 	rhsValue, _ := rhsResolver.ResolveValue(args)

// 	result, _ := evaluate(lhsValue, rhsValue, operandType, condition.Operator)

// 	return result
// }

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
