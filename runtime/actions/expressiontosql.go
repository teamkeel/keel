package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// Produces a complete SQL condition from an expression. It is intended for use to construct a filtered SQL statement.
func expressionToSqlCondition(
	context context.Context,
	expression *parser.Expression,
	operation *proto.Operation,
	schema *proto.Schema,
	data map[string]any,
) (result string, err error) {

	conditions := expression.Conditions()
	if len(conditions) != 1 {
		return "", fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(conditions))
	}
	condition := conditions[0]

	if condition.Type() == parser.ValueCondition {
		valueType, _ := getOperandType(condition.LHS, operation, schema)
		if valueType != proto.Type_TYPE_BOOL {
			return "", fmt.Errorf("value operand must be of type bool, not %s", condition.Type())
		}

		value, err := generateSqlOperand(context, condition.LHS, operation, schema, data, valueType)
		if err != nil {
			return "", err
		}

		return value, nil
	}

	if condition.Type() != parser.LogicalCondition {
		return "", fmt.Errorf("can only handle condition type of LogicalCondition, have: %s", condition.Type())
	}

	// Determine the native protobuf type underlying the expression comparison
	var operandType proto.Type
	lhsType, _ := getOperandType(condition.LHS, operation, schema)
	rhsType, _ := getOperandType(condition.RHS, operation, schema)
	switch {
	case lhsType != proto.Type_TYPE_UNKNOWN && (lhsType == rhsType || rhsType == proto.Type_TYPE_UNKNOWN):
		operandType = lhsType
	case rhsType != proto.Type_TYPE_UNKNOWN && (lhsType == rhsType || lhsType == proto.Type_TYPE_UNKNOWN):
		operandType = rhsType
	default:
		return "", fmt.Errorf("lhs: %s, and rhs: %s, are not of the same native type", lhsType, rhsType)
	}

	lhsSqlSegment, _ := generateSqlOperand(context, condition.LHS, operation, schema, data, operandType)
	rhsSqlSegment, _ := generateSqlOperand(context, condition.RHS, operation, schema, data, operandType)

	// fmt.Printf("%s = %s", lhsSqlSegment, rhsSqlSegment)
	// fmt.Println()

	// The LHS and RHS types must be equal unless the RHS is a null literal
	// if lhsType != rhsType && rhsValue != nil {
	// 	return "", fmt.Errorf("lhs type: %s, and rhs type: %s, are not the same", lhsType, rhsType)
	// }

	return fmt.Sprintf("%s = %s", lhsSqlSegment, rhsSqlSegment), nil
}

// Produces a SQL operand from an expression operand. It is intended for use to construct a complete SQL condition.
// An expression operand can be of the following types:
//   - a literal in the schema; @where(post.isActive == false)
//   - an implicit input value on the request; @where(post.isActive == isActive)
//   - an explicit input value on the request; @permission(expression: hasElevatedPriveledges == true)
//   - a database field for persisted data; @permission(post.isAdmin == false)
//   - a value in the context; @permission(expression: post.Identity == ctx.Identity)
func generateSqlOperand(
	context context.Context,
	operand *parser.Operand,
	operation *proto.Operation,
	schema *proto.Schema,
	args map[string]any,
	operandType proto.Type,
) (string, error) {

	isLiteral, _ := operand.IsLiteralType()

	switch {
	case isLiteral:
		// if literal, then pull the value from the schema
		value, err := toNative(operand, operandType)

		if err != nil {
			return "", fmt.Errorf("unexpected error parsing literal")
		}

		// todo: operator must change to IS (IS NULL) in sql condition
		if value == nil {
			return "null", nil
		}

		if operandType == proto.Type_TYPE_STRING {
			return fmt.Sprintf("'%v'", value), nil
		} else {
			return fmt.Sprintf("%v", value), nil
		}
	case operand.Ident != nil && proto.EnumExists(schema.Enums, operand.Ident.Fragments[0].Fragment):
		// if enum literal, then pull the value from the schema
		return fmt.Sprintf("'%v'", operand.Ident.Fragments[1].Fragment), nil
	case operand.Ident != nil && len(operand.Ident.Fragments) == 1 && args[operand.Ident.Fragments[0].Fragment] != nil:
		// if implicit or explicit input, then evaluate the value from the request args
		inputValue := args[operand.Ident.Fragments[0].Fragment]

		if operandType == proto.Type_TYPE_STRING {
			return fmt.Sprintf("'%v'", inputValue), nil
		} else {
			return fmt.Sprintf("%v", inputValue), nil
		}
	case operand.Ident != nil && strcase.ToCamel(operand.Ident.Fragments[0].Fragment) == strcase.ToCamel(operation.ModelName):
		// if field name, then build sql segment
		modelTarget := strcase.ToSnake(operand.Ident.Fragments[0].Fragment)
		fieldName := strcase.ToSnake(operand.Ident.Fragments[1].Fragment)

		return fmt.Sprintf("%s.%s", modelTarget, fieldName), nil
	case operand.Ident != nil && operand.Ident.IsContextIdentityField():
		isAuthenticated := runtimectx.IsAuthenticated(context)

		if !isAuthenticated {
			return "", nil // todo: err
		}

		ksuid, err := runtimectx.GetIdentity(context)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("'%v'", ksuid), err
	case operand.Ident != nil && operand.Ident.IsContextIsAuthenticatedField():
		isAuthenticated := runtimectx.IsAuthenticated(context)

		return fmt.Sprintf("%v", isAuthenticated), nil
	case operand.Ident != nil && operand.Ident.IsContextNowField():
		return "", fmt.Errorf("cannot yet handle ctx field now")
	case operand.Type() == parser.TypeArray:
		return "", fmt.Errorf("cannot yet handle operand of type non-literal array")
	default:
		return "", fmt.Errorf("cannot handle operand of unknown type")

	}
}
