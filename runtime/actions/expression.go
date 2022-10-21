package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// interpretExpressionField examines the given expression, in order to work out how to construct a gorm WHERE clause.
func interpretExpressionField(
	expr *parser.Expression,
	operation *proto.Operation,
	schema *proto.Schema,
) (*proto.Field, ActionOperator, error) {

	// Make sure the expression is in the form we can handle.

	conditions := expr.Conditions()
	if len(conditions) != 1 {
		return nil, Unknown, fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(conditions))
	}
	condition := conditions[0]
	cType := condition.Type()
	if cType != parser.LogicalCondition {
		return nil, Unknown, fmt.Errorf("cannot yet handle condition types other than LogicalCondition, have: %s", cType)
	}

	operatorStr := condition.Operator.ToString()
	operator, err := expressionOperatorToActionOperator(operatorStr)
	if err != nil {
		return nil, Unknown, err
	}

	if condition.LHS.Type() != parser.TypeIdent {
		return nil, operator, fmt.Errorf("cannot handle LHS of type other than TypeIdent, have: %s", condition.LHS.Type())
	}
	if condition.RHS.Type() != parser.TypeIdent {
		return nil, operator, fmt.Errorf("cannot handle RHS of type other than TypeIdent, have: %s", condition.LHS.Type())
	}

	lhs := condition.LHS
	if len(lhs.Ident.Fragments) != 2 {
		return nil, operator, fmt.Errorf("cannot handle LHS identifier unless it has 2 fragments, have: %d", len(lhs.Ident.Fragments))
	}

	rhs := condition.RHS
	if len(rhs.Ident.Fragments) != 1 {
		return nil, operator, fmt.Errorf("cannot handle RHS identifier unless it has 1 fragment, have: %d", len(rhs.Ident.Fragments))
	}

	// Make sure the first fragment in the LHS is the name of the model of which this operation is part.
	// e.g. "person" in the example above.
	modelTarget := strcase.ToCamel(lhs.Ident.Fragments[0].Fragment)

	if modelTarget != operation.ModelName {
		return nil, operator, fmt.Errorf("can only handle the first LHS fragment referencing the Operation's model, have: %s", modelTarget)
	}

	// Make sure the second fragment in the LHS is the name of a field of the model of which this operation is part.
	// e.g. "name" in the example above.
	fieldName := lhs.Ident.Fragments[1].Fragment

	field := proto.FindField(schema.Models, modelTarget, fieldName)
	if !proto.ModelHasField(schema, modelTarget, fieldName) {
		return nil, operator, fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
	}

	// Now we have all the data we need to return
	return field, operator, nil
}

// getOperandType determines the underlying type to compare with for an operand
func getOperandType(
	operand *parser.Operand,
	operation *proto.Operation,
	schema *proto.Schema,
) (proto.Type, error) {

	if operand.Ident == nil {
		switch {
		case operand.String != nil:
			return proto.Type_TYPE_STRING, nil
		case operand.Number != nil:
			return proto.Type_TYPE_INT, nil
		case operand.True || operand.False:
			return proto.Type_TYPE_BOOL, nil
		case operand.Null:
			return proto.Type_TYPE_UNKNOWN, nil
		default:
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("cannot handle operand type")
		}
	}

	target := operand.Ident.Fragments[0].Fragment
	switch {
	case strcase.ToCamel(target) == operation.ModelName:
		// implicit input as database field (with the model name)
		modelTarget := strcase.ToCamel(target)
		fieldName := operand.Ident.Fragments[1].Fragment

		if !proto.ModelHasField(schema, strcase.ToCamel(modelTarget), fieldName) {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
		}

		operandType := proto.FindField(schema.Models, strcase.ToCamel(modelTarget), fieldName).Type.Type
		return operandType, nil
	case operand.Ident != nil && len(operand.Ident.Fragments) == 1 && proto.ModelHasField(schema, operation.ModelName, operand.Ident.Fragments[0].Fragment):
		// implicit input (without the model name)
		modelTarget := strcase.ToCamel(operation.ModelName)
		fieldName := operand.Ident.Fragments[0].Fragment
		operandType := proto.FindField(schema.Models, modelTarget, fieldName).Type.Type
		return operandType, nil

	case operand.Ident.IsContext():
		fieldName := operand.Ident.Fragments[1].Fragment
		return runtimectx.ContextFieldTypes[fieldName], nil // todo: if not found
	default:
		return proto.Type_TYPE_UNKNOWN, fmt.Errorf("cannot handle operand target %s", target)
	}
}

// evaluateOperandValue evaluates the value to compare with for an operand
func evaluateOperandValue(
	context context.Context,
	operand *parser.Operand,
	operation *proto.Operation,
	schema *proto.Schema,
	data map[string]any,
	operandType proto.Type,
) (any, error) {

	isLiteral, _ := operand.IsLiteralType()

	switch {
	case isLiteral:
		return toNative(operand, operandType)
	case operand.Ident != nil && proto.EnumExists(schema.Enums, operand.Ident.Fragments[0].Fragment):
		return operand.Ident.Fragments[1].Fragment, nil
	case operand.Ident != nil && len(operand.Ident.Fragments) == 1 && data[operand.Ident.Fragments[0].Fragment] != nil:
		inputValue := data[operand.Ident.Fragments[0].Fragment]
		return inputValue, nil
	case operand.Ident != nil && strcase.ToCamel(operand.Ident.Fragments[0].Fragment) == strcase.ToCamel(operation.ModelName):
		modelTarget := strcase.ToCamel(operand.Ident.Fragments[0].Fragment)
		fieldName := operand.Ident.Fragments[1].Fragment

		if !proto.ModelHasField(schema, strcase.ToCamel(modelTarget), fieldName) {
			return nil, fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
		}

		field := proto.FindField(schema.Models, strcase.ToCamel(modelTarget), fieldName)
		operandType := field.Type.Type
		isOptional := field.Optional
		fieldValue := data[fieldName]

		// If the value of the optional field then return nil for later comparison
		if fieldValue == nil {
			if isOptional {
				return nil, nil
			} else {
				return nil, fmt.Errorf("required field is nil: %s", fieldName)
			}
		}

		switch operandType {
		case proto.Type_TYPE_STRING, proto.Type_TYPE_BOOL:
			return fieldValue, nil
		case proto.Type_TYPE_INT:
			// todo: unify these to a single type at the source?
			switch v := fieldValue.(type) {
			case int:
				// Sourced from GraphQL input parameters.
				return int64(fieldValue.(int)), nil
			case float64:
				// Sourced from integration test framework.
				return int64(fieldValue.(float64)), nil
			case int32:
				// Sourced from database.
				return int64(fieldValue.(int32)), nil // todo: https://linear.app/keel/issue/RUN-98/number-type-as-int32-or-int64
			case int64:
				// Sourced from a default set value on a field.
				return fieldValue, nil
			default:
				return nil, fmt.Errorf("cannot yet parse %s to int64", v)
			}
		case proto.Type_TYPE_ENUM:
			return fieldValue, nil
		case proto.Type_TYPE_IDENTITY:
			switch v := fieldValue.(type) {
			case *ksuid.KSUID:
				// Sourced from GraphQL input parameters.
				return *fieldValue.(*ksuid.KSUID), nil
			case string:
				// Sourced from database.
				value, err := ksuid.Parse(fieldValue.(string))
				if err != nil {
					return nil, fmt.Errorf("cannot parse %s to ksuid", fieldValue)
				}
				return value, nil
			default:
				return nil, fmt.Errorf("cannot yet parse %s to ksuid.KSUID", v)
			}
		default:
			return nil, fmt.Errorf("cannot yet compare operand of type %s", operandType)
		}
	case operand.Ident != nil && operand.Ident.IsContextIdentityField():
		isAuthenticated := runtimectx.IsAuthenticated(context)

		if !isAuthenticated {
			return nil, nil // todo: err?
		}

		ksuid, err := runtimectx.GetIdentity(context)
		if err != nil {
			return nil, err
		}
		return *ksuid, nil
	case operand.Ident != nil && operand.Ident.IsContextIsAuthenticatedField():
		isAuthenticated := runtimectx.IsAuthenticated(context)

		return isAuthenticated, nil
	case operand.Ident != nil && operand.Ident.IsContextNowField():
		return nil, fmt.Errorf("cannot yet handle ctx field now")
	case operand.Type() == parser.TypeArray:
		return nil, fmt.Errorf("cannot yet handle operand of type non-literal array")
	default:
		return nil, fmt.Errorf("cannot handle operand of unknown type")

	}
}
