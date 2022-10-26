package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"gorm.io/gorm"
)

// The FilterResolver generates database queries for filters which are specified implicitly.
type FilterResolver struct {
	scope *Scope
}

// The ExpressionResolver generates database queries for filters which are specified as expressions.
// Also provides a means to resolve and evaluate conditions in-proc (i.e. outside of the database).
type ExpressionResolver struct {
	scope *Scope
}

func NewFilterResolver(scope *Scope) *FilterResolver {
	return &FilterResolver{
		scope: scope,
	}
}

// Generates a database query statement for a filter.
func (resolver *FilterResolver) ResolveQueryStatement(fieldName string, value any, operator ActionOperator) (*gorm.DB, error) {
	fieldName = strcase.ToSnake(fieldName)

	queryTemplate, err := generateFilterTemplate(fieldName, "?", operator)
	if err != nil {
		return nil, err
	}

	queryArgument := value
	switch operator {
	case StartsWith:
		queryArgument = queryArgument.(string) + "%%"
	case EndsWith:
		queryArgument = "%%" + queryArgument.(string)
	case Contains, NotContains:
		queryArgument = "%%" + queryArgument.(string) + "%%"
	}

	query := resolver.scope.query.
		Session(&gorm.Session{NewDB: true}).
		Where(queryTemplate, queryArgument)

	return query, nil
}

func NewExpressionResolver(scope *Scope) *ExpressionResolver {
	return &ExpressionResolver{
		scope: scope,
	}
}

// Determines if the expression can be evaluated on the runtime process
// as opposed to producing a SQL statement and querying against the database.
func (resolver *ExpressionResolver) CanResolveInMemory(expression *parser.Expression) bool {
	condition := expression.Conditions()[0]

	lhsResolver := NewOperandResolver(resolver.scope.context, resolver.scope.schema, resolver.scope.operation, condition.LHS)

	if condition.Type() == parser.ValueCondition {
		return !lhsResolver.IsDatabaseColumn()
	}

	rhsResolver := NewOperandResolver(resolver.scope.context, resolver.scope.schema, resolver.scope.operation, condition.RHS)
	referencesDatabaseColumns := lhsResolver.IsDatabaseColumn() || rhsResolver.IsDatabaseColumn()

	return !(referencesDatabaseColumns)
}

// Evaluated the expression in the runtime process without generated and query against the database.
func (resolver *ExpressionResolver) ResolveInMemory(expression *parser.Expression, args RequestArguments, writeValues map[string]any) bool {
	condition := expression.Conditions()[0]

	lhsResolver := NewOperandResolver(resolver.scope.context, resolver.scope.schema, resolver.scope.operation, condition.LHS)
	operandType, _ := lhsResolver.GetOperandType()
	lhsValue, _ := lhsResolver.ResolveValue(args, writeValues, operandType)

	if condition.Type() == parser.ValueCondition {
		result, _ := evaluateInProcess(lhsValue, true, operandType, &parser.Operator{Symbol: parser.OperatorEquals})
		return result
	}

	rhsResolver := NewOperandResolver(resolver.scope.context, resolver.scope.schema, resolver.scope.operation, condition.RHS)

	rhsValue, _ := rhsResolver.ResolveValue(args, writeValues, operandType)

	result, _ := evaluateInProcess(lhsValue, rhsValue, operandType, condition.Operator)

	return result
}

// Generates a database query statement for an expression.
func (resolver *ExpressionResolver) ResolveQueryStatement(expression *parser.Expression, args RequestArguments, writeValues map[string]any) (*gorm.DB, error) {
	if len(expression.Conditions()) != 1 {
		return nil, fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(expression.Conditions()))
	}

	condition := expression.Conditions()[0]

	if condition.Type() != parser.ValueCondition && condition.Type() != parser.LogicalCondition {
		return nil, fmt.Errorf("can only handle condition type of LogicalCondition or ValueCondition, have: %s", condition.Type())
	}

	operatorStr := condition.Operator.ToString()
	operator, err := expressionOperatorToActionOperator(operatorStr)
	if err != nil {
		return nil, err
	}

	queryTemplate, queryArguments, err := resolver.generateQuery(condition, operator, args, writeValues)
	if err != nil {
		return nil, err
	}

	query := resolver.scope.query.
		Session(&gorm.Session{NewDB: true}).
		Where(queryTemplate, queryArguments...)

	return query, nil
}

type OperandResolver struct {
	context   context.Context
	schema    *proto.Schema
	operation *proto.Operation
	operand   *parser.Operand
}

func NewOperandResolver(ctx context.Context, schema *proto.Schema, operation *proto.Operation, operand *parser.Operand) *OperandResolver {
	return &OperandResolver{
		context:   ctx,
		schema:    schema,
		operation: operation,
		operand:   operand,
	}
}

func (resolver *OperandResolver) IsLiteral() bool {
	isLiteral, _ := resolver.operand.IsLiteralType()
	isEnumLiteral := resolver.operand.Ident != nil && proto.EnumExists(resolver.schema.Enums, resolver.operand.Ident.Fragments[0].Fragment)
	return isLiteral || isEnumLiteral
}

func (resolver *OperandResolver) IsImplicitInput() bool {
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
	isSingleFragmentIdent := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) == 1

	if !isSingleFragmentIdent {
		return false
	}

	input, found := lo.Find(resolver.operation.Inputs, func(in *proto.OperationInput) bool {
		return in.Name == resolver.operand.Ident.Fragments[0].Fragment
	})

	return found && input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_EXPLICIT
}

func (resolver *OperandResolver) IsDatabaseColumn() bool {
	// It is not possible to reference model fields on create, when no data exists in the database.
	// Therefore a model name used in the expression will actually refer to a write value
	// (i.e. new value to be written to the database)
	if resolver.operation.Type == proto.OperationType_OPERATION_TYPE_CREATE {
		return false
	}

	isMultiFragmentIdent := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) > 1

	if !isMultiFragmentIdent {
		return false
	}

	modelTarget := resolver.operand.Ident.Fragments[0].Fragment

	return modelTarget == strcase.ToLowerCamel(resolver.operation.ModelName)
}

func (resolver *OperandResolver) IsWriteValue() bool {
	if resolver.operation.Type != proto.OperationType_OPERATION_TYPE_CREATE {
		return false
	}

	isMultiFragmentIdent := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) > 1

	if !isMultiFragmentIdent {
		return false
	}

	modelTarget := resolver.operand.Ident.Fragments[0].Fragment

	return modelTarget == strcase.ToLowerCamel(resolver.operation.ModelName)
}

func (resolver *OperandResolver) IsContextField() bool {
	return resolver.operand.Ident.IsContext()
}

func (resolver *OperandResolver) GetOperandType() (proto.Type, error) {
	operand := resolver.operand
	operation := resolver.operation
	schema := resolver.schema

	switch {
	case resolver.IsLiteral():
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
		} else if resolver.operand.Ident != nil && proto.EnumExists(resolver.schema.Enums, resolver.operand.Ident.Fragments[0].Fragment) {
			return proto.Type_TYPE_ENUM, nil
		} else {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("unknown literal type")
		}
	case resolver.IsDatabaseColumn():
		modelTarget := strcase.ToCamel(operand.Ident.Fragments[0].Fragment)
		fieldName := operand.Ident.Fragments[1].Fragment

		if !proto.ModelHasField(schema, strcase.ToCamel(modelTarget), fieldName) {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
		}

		operandType := proto.FindField(schema.Models, strcase.ToCamel(modelTarget), fieldName).Type.Type
		return operandType, nil
	case resolver.IsWriteValue():
		if operation.Type != proto.OperationType_OPERATION_TYPE_CREATE {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("only the create operation can refer to write values in expressions")
		}
		modelTarget := resolver.operand.Ident.Fragments[0].Fragment
		fieldName := operand.Ident.Fragments[1].Fragment
		operandType := proto.FindField(schema.Models, strcase.ToCamel(modelTarget), fieldName).Type.Type
		return operandType, nil
	case resolver.IsImplicitInput():
		modelTarget := strcase.ToCamel(operation.ModelName)
		inputName := operand.Ident.Fragments[0].Fragment
		operandType := proto.FindField(schema.Models, modelTarget, inputName).Type.Type
		return operandType, nil
	case resolver.IsExplicitInput():
		inputName := operand.Ident.Fragments[0].Fragment
		input := proto.FindInput(operation, inputName)
		return input.Type.Type, nil
	case operand.Ident.IsContext():
		fieldName := operand.Ident.Fragments[1].Fragment
		return runtimectx.ContextFieldTypes[fieldName], nil
	default:
		return proto.Type_TYPE_UNKNOWN, fmt.Errorf("cannot handle operand target %s", operand.Ident.Fragments[0].Fragment)
	}
}

func (resolver *OperandResolver) ResolveValue(
	args map[string]any,
	writeValues map[string]any,
	operandType proto.Type, //todo: infer from schema
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
		inputName := resolver.operand.Ident.Fragments[0].Fragment
		value, ok := args[inputName]
		if !ok {
			return nil, fmt.Errorf("implicit or explicit input '%s' does not exist in arguments", inputName)
		}
		return value, nil
	case resolver.IsWriteValue():
		inputName := strcase.ToSnake(resolver.operand.Ident.Fragments[1].Fragment)
		value, ok := writeValues[inputName]
		if !ok {
			return nil, fmt.Errorf("value '%s' does not exist in write values", inputName)
		}
		return value, nil
	case resolver.IsDatabaseColumn():
		panic("cannot resolve operand value when IsDatabaseColumn() is true")
	case resolver.IsContextField() && resolver.operand.Ident.IsContextIdentityField():
		isAuthenticated := runtimectx.IsAuthenticated(resolver.context)

		if !isAuthenticated {
			return nil, nil
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

func (resolver *ExpressionResolver) generateQuery(condition *parser.Condition, operator ActionOperator, args map[string]any, writeValues map[string]any) (queryTemplate string, queryArguments []any, err error) {

	var lhsOperandType, rhsOperandType proto.Type

	lhsResolver := NewOperandResolver(resolver.scope.context, resolver.scope.schema, resolver.scope.operation, condition.LHS)
	rhsResolver := NewOperandResolver(resolver.scope.context, resolver.scope.schema, resolver.scope.operation, condition.RHS)
	lhsOperandType, err = lhsResolver.GetOperandType()
	if err != nil {
		return "", nil, fmt.Errorf("cannot resolve operand type of LHS operand")
	}

	if condition.Type() == parser.ValueCondition {
		if lhsOperandType != proto.Type_TYPE_BOOL {
			return "", nil, fmt.Errorf("single operands in a value condition must be of type boolean")
		}

		// A value condition only has one operand in the expression,
		// for example, permission(expression: ctx.isAuthenticated),
		// so we must set the operator and RHS value (== true) ourselves.
		rhsOperandType = lhsOperandType
		operator = Equals
	} else {
		rhsOperandType, err = rhsResolver.GetOperandType()
		if err != nil {
			return "", nil, fmt.Errorf("cannot resolve operand type of RHS operand")
		}
	}

	var template string
	var queryArgs []any

	// ? is the gorm template token which is replaced when populating the operand arguments.
	var lhsSqlOperand, rhsSqlOperand any
	lhsSqlOperand = "?"
	rhsSqlOperand = "?"

	if !lhsResolver.IsDatabaseColumn() {
		lhsValue, err := lhsResolver.ResolveValue(args, writeValues, lhsOperandType)
		if err != nil {
			return "", nil, err
		}

		switch operator {
		case StartsWith:
			lhsValue = lhsValue.(string) + "%%"
		case EndsWith:
			lhsValue = "%%" + lhsValue.(string)
		case Contains, NotContains:
			lhsValue = "%%" + lhsValue.(string) + "%%"
		}

		queryArgs = append(queryArgs, lhsValue)
	} else {
		// Generate the table's column name from the fragments (e.g. post.sub_title)
		// And replace the ? gorm template token with the column name
		modelTarget := strcase.ToSnake(lhsResolver.operand.Ident.Fragments[0].Fragment)
		fieldName := strcase.ToSnake(lhsResolver.operand.Ident.Fragments[1].Fragment)
		lhsSqlOperand = fmt.Sprintf("%s.%s", modelTarget, fieldName)
	}

	if condition.Type() == parser.ValueCondition {
		queryArgs = append(queryArgs, true)
	} else if !rhsResolver.IsDatabaseColumn() {
		rhsValue, err := rhsResolver.ResolveValue(args, writeValues, rhsOperandType)
		if err != nil {
			return "", nil, err
		}

		// If the value is nil, then we can bake this straight into the template
		if rhsValue == nil {
			rhsSqlOperand = nil
		} else {
			switch operator {
			case StartsWith:
				rhsValue = rhsValue.(string) + "%%"
			case EndsWith:
				rhsValue = "%%" + rhsValue.(string)
			case Contains, NotContains:
				rhsValue = "%%" + rhsValue.(string) + "%%"
			}

			queryArgs = append(queryArgs, rhsValue)
		}
	} else {
		// Generate the table's column operand from the fragments (e.g. post.sub_title)
		// And replace the ? gorm template token with the column name
		modelTarget := strcase.ToSnake(rhsResolver.operand.Ident.Fragments[0].Fragment)
		fieldName := strcase.ToSnake(rhsResolver.operand.Ident.Fragments[1].Fragment)
		rhsSqlOperand = fmt.Sprintf("%s.%s", modelTarget, fieldName)
	}

	template, err = generateFilterTemplate(lhsSqlOperand, rhsSqlOperand, operator)
	if err != nil {
		return "", nil, err
	}

	return template, queryArgs, nil
}

func generateFilterTemplate(lhsSqlOperand any, rhsSqlOperand any, operator ActionOperator) (string, error) {
	var template string

	switch operator {
	case Equals:
		if rhsSqlOperand == nil {
			template = fmt.Sprintf("%s IS NULL", lhsSqlOperand)
		} else {
			template = fmt.Sprintf("%s = %s", lhsSqlOperand, rhsSqlOperand)
		}
	case NotEquals:
		if rhsSqlOperand == nil {
			template = fmt.Sprintf("%s IS NOT NULL", lhsSqlOperand)
		} else {
			template = fmt.Sprintf("%s != %s", lhsSqlOperand, rhsSqlOperand)
		}
	case StartsWith, EndsWith, Contains:
		template = fmt.Sprintf("%s LIKE %s", lhsSqlOperand, rhsSqlOperand)
	case NotContains:
		template = fmt.Sprintf("%s NOT LIKE %s", lhsSqlOperand, rhsSqlOperand)
	case OneOf:
		template = fmt.Sprintf("%s in %s", lhsSqlOperand, rhsSqlOperand)
	case LessThan:
		template = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case LessThanEquals:
		template = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThan:
		template = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThanEquals:
		template = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
	case Before:
		template = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case After:
		template = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrBefore:
		template = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrAfter:
		template = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
	default:
		return "", fmt.Errorf("operator: %v is not yet supported", operator)
	}

	return template, nil
}

func evaluateInProcess(
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
	case proto.Type_TYPE_IDENTITY:
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
