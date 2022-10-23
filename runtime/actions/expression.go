package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// // interpretExpressionField examines the given expression, in order to work out how to construct a gorm WHERE clause.
// func interpretExpressionField(
// 	expr *parser.Expression,
// 	operation *proto.Operation,
// 	schema *proto.Schema,
// ) (*proto.Field, ActionOperator, error) {

// 	// Make sure the expression is in the form we can handle.

// 	conditions := expr.Conditions()
// 	if len(conditions) != 1 {
// 		return nil, Unknown, fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(conditions))
// 	}
// 	condition := conditions[0]
// 	cType := condition.Type()
// 	if cType != parser.LogicalCondition {
// 		return nil, Unknown, fmt.Errorf("cannot yet handle condition types other than LogicalCondition, have: %s", cType)
// 	}

// 	operatorStr := condition.Operator.ToString()
// 	operator, err := expressionOperatorToActionOperator(operatorStr)
// 	if err != nil {
// 		return nil, Unknown, err
// 	}

// 	if condition.LHS.Type() != parser.TypeIdent {
// 		return nil, operator, fmt.Errorf("cannot handle LHS of type other than TypeIdent, have: %s", condition.LHS.Type())
// 	}
// 	if condition.RHS.Type() != parser.TypeIdent {
// 		return nil, operator, fmt.Errorf("cannot handle RHS of type other than TypeIdent, have: %s", condition.LHS.Type())
// 	}

// 	lhs := condition.LHS
// 	if len(lhs.Ident.Fragments) != 2 {
// 		return nil, operator, fmt.Errorf("cannot handle LHS identifier unless it has 2 fragments, have: %d", len(lhs.Ident.Fragments))
// 	}

// 	rhs := condition.RHS
// 	if len(rhs.Ident.Fragments) != 1 {
// 		return nil, operator, fmt.Errorf("cannot handle RHS identifier unless it has 1 fragment, have: %d", len(rhs.Ident.Fragments))
// 	}

// 	// Make sure the first fragment in the LHS is the name of the model of which this operation is part.
// 	// e.g. "person" in the example above.
// 	modelTarget := strcase.ToCamel(lhs.Ident.Fragments[0].Fragment)

// 	if modelTarget != operation.ModelName {
// 		return nil, operator, fmt.Errorf("can only handle the first LHS fragment referencing the Operation's model, have: %s", modelTarget)
// 	}

// 	// Make sure the second fragment in the LHS is the name of a field of the model of which this operation is part.
// 	// e.g. "name" in the example above.
// 	fieldName := lhs.Ident.Fragments[1].Fragment

// 	field := proto.FindField(schema.Models, modelTarget, fieldName)
// 	if !proto.ModelHasField(schema, modelTarget, fieldName) {
// 		return nil, operator, fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
// 	}

// 	// Now we have all the data we need to return
// 	return field, operator, nil
// }

type OperandResolver struct {
	context   context.Context
	operand   *parser.Operand
	operation *proto.Operation
	schema    *proto.Schema
}

func NewOperandResolver(ctx context.Context, operand *parser.Operand, operation *proto.Operation, schema *proto.Schema) *OperandResolver {
	return &OperandResolver{
		context:   ctx,
		operand:   operand,
		operation: operation,
		schema:    schema,
	}
}

func (resolver *OperandResolver) IsLiteral() bool {
	isLiteral, _ := resolver.operand.IsLiteralType()
	isEnumLiteral := resolver.operand.Ident != nil && proto.EnumExists(resolver.schema.Enums, resolver.operand.Ident.Fragments[0].Fragment)
	return isLiteral || isEnumLiteral
}

func (resolver *OperandResolver) IsImplicitInput() bool {
	// implicit and explicit inputs are strictly one fragment
	// todo: with exception of create()?
	isSingleFragment := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) == 1

	if !isSingleFragment {
		return false
	}

	input, found := lo.Find(resolver.operation.Inputs, func(in *proto.OperationInput) bool {
		return in.Name == resolver.operand.Ident.Fragments[0].Fragment
	})

	return found && input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT
}

func (resolver *OperandResolver) IsExplicitInput() bool {
	// implicit and explicit inputs are strictly one fragment
	// todo: with exception of create()?
	isSingleFragmentIdent := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) == 1

	if !isSingleFragmentIdent {
		return false
	}

	input, found := lo.Find(resolver.operation.Inputs, func(in *proto.OperationInput) bool {
		return in.Name == resolver.operand.Ident.Fragments[0].Fragment
	})

	return found && input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_EXPLICIT
}

func (resolver *OperandResolver) IsModelField() bool {
	isMultiFragmentIdent := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) > 1

	if !isMultiFragmentIdent {
		return false
	}

	modelTarget := resolver.operand.Ident.Fragments[0].Fragment

	return isMultiFragmentIdent && modelTarget == strcase.ToLowerCamel(resolver.operation.ModelName)
}

func (resolver *OperandResolver) IsContextField() bool {
	return resolver.operand.Ident.IsContext()
}

func (resolver *OperandResolver) GetOperandType() (proto.Type, error) {
	operand := resolver.operand
	operation := resolver.operation
	schema := resolver.schema

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

func (resolver *OperandResolver) ResolveValue(
	data map[string]any,
	operandType proto.Type,
) (any, error) {

	switch {
	case resolver.IsLiteral():
		isLiteral, _ := resolver.operand.IsLiteralType()
		if isLiteral {
			return toNative(resolver.operand, operandType)
		} else if resolver.operand.Ident != nil && proto.EnumExists(resolver.schema.Enums, resolver.operand.Ident.Fragments[0].Fragment) {
			return resolver.operand.Ident.Fragments[1].Fragment, nil
		} else {
			panic("unknown literal type")
		}

	case resolver.IsImplicitInput(), resolver.IsExplicitInput():
		inputValue := data[resolver.operand.Ident.Fragments[0].Fragment] // todo: convert to type?
		return inputValue, nil
	case resolver.IsModelField():
		panic("cannot resolve operand value when IsModelField() is true")
	case resolver.IsContextField() && resolver.operand.Ident.IsContextIdentityField():
		isAuthenticated := runtimectx.IsAuthenticated(resolver.context)

		if !isAuthenticated {
			return nil, nil // todo: err?
		}

		ksuid, err := runtimectx.GetIdentity(resolver.context)
		if err != nil {
			return nil, err
		}
		return *ksuid, nil
	case resolver.IsContextField() && resolver.operand.Ident.IsContextIsAuthenticatedField():
		isAuthenticated := runtimectx.IsAuthenticated(resolver.context)
		return isAuthenticated, nil
	case resolver.IsContextField() && resolver.operand.Ident.IsContextNowField():
		return nil, fmt.Errorf("cannot yet handle ctx field now")
	case resolver.operand.Type() == parser.TypeArray:
		return nil, fmt.Errorf("cannot yet handle operand of type non-literal array")
	default:
		return nil, fmt.Errorf("cannot handle operand of unknown type")

	}
}
