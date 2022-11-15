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
)

// Include a filter on the query based on the implicit input provided.
func (query *QueryBuilder) whereByImplicitFilter(scope *Scope, input *proto.OperationInput, fieldName string, operator ActionOperator, value any) error {
	model := input.ModelName

	databaseTable := strcase.ToSnake(model)
	databaseColumn := strcase.ToSnake(fieldName)

	fragments := input.Target
	fragmentCount := len(fragments)

	for i := 0; i < fragmentCount; i++ {
		currentFragment := fragments[i]

		if !proto.ModelHasField(scope.schema, model, currentFragment) {
			return fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
		}

		if i < fragmentCount-1 {
			// We know that the current fragment is a related table because it's not the last fragment
			relatedModelField := proto.FindField(scope.schema.Models, model, currentFragment)
			relatedTable := strcase.ToSnake(relatedModelField.Type.ModelName.Value)

			primaryKeyDbColumn := "id"

			var join string
			if relatedModelField.ForeignKeyFieldName != nil {
				// M:1
				foreignKeyDbColumn := strcase.ToSnake(relatedModelField.ForeignKeyFieldName.Value)

				join = fmt.Sprintf("INNER JOIN %s ON %s = %s",
					relatedTable,
					fmt.Sprintf("%s.%s", databaseTable, foreignKeyDbColumn),
					fmt.Sprintf("%s.%s", relatedTable, primaryKeyDbColumn))
			} else {
				// 1:M
				fkModel := proto.FindModel(scope.schema.Models, relatedModelField.Type.ModelName.Value)
				fkField, found := lo.Find(fkModel.Fields, func(field *proto.Field) bool {
					return field.Type.Type == proto.Type_TYPE_MODEL && field.Type.ModelName.Value == model
				})
				if !found {
					return fmt.Errorf("no foreign key field found on related model %s", model)
				}

				foreignKeyDbColumn := strcase.ToSnake(fkField.ForeignKeyFieldName.Value)

				join = fmt.Sprintf("INNER JOIN %s ON %s = %s",
					relatedTable,
					fmt.Sprintf("%s.%s", databaseTable, primaryKeyDbColumn),
					fmt.Sprintf("%s.%s", relatedTable, foreignKeyDbColumn))
			}

			query.Join(join)

			databaseTable = relatedTable
			model = relatedModelField.Type.ModelName.Value
		} else {
			databaseColumn = strcase.ToSnake(currentFragment)
		}
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

	lhsOperand := Column(databaseTable, databaseColumn)
	rhsOperand := Value(queryArgument)

	query.Where(lhsOperand, operator, rhsOperand)

	return nil
}

// Determines if the expression can be evaluated on the runtime process
// as opposed to producing a SQL statement and querying against the database.
func canResolveInMemory(scope *Scope, expression *parser.Expression) bool {
	condition := expression.Conditions()[0]

	lhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, condition.LHS)

	if condition.Type() == parser.ValueCondition {
		return !lhsResolver.IsDatabaseColumn()
	}

	rhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, condition.RHS)
	referencesDatabaseColumns := lhsResolver.IsDatabaseColumn() || rhsResolver.IsDatabaseColumn()

	return !(referencesDatabaseColumns)
}

// Evaluated the expression in the runtime process without generated and query against the database.
func resolveInMemory(scope *Scope, expression *parser.Expression, args WhereArgs, writeValues map[string]any) bool {
	condition := expression.Conditions()[0]

	lhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, condition.LHS)
	operandType, _ := lhsResolver.GetOperandType()
	lhsValue, _ := lhsResolver.ResolveValue(args, writeValues)

	if condition.Type() == parser.ValueCondition {
		result, _ := evaluateInProcess(lhsValue, true, operandType, &parser.Operator{Symbol: parser.OperatorEquals})
		return result
	}

	rhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, condition.RHS)

	rhsValue, _ := rhsResolver.ResolveValue(args, writeValues)

	result, _ := evaluateInProcess(lhsValue, rhsValue, operandType, condition.Operator)

	return result
}

// Include a filter on the query based on the expression provided.
// Include a filter on the query based on the expression provided.
func (query *QueryBuilder) whereByExpression(scope *Scope, expression *parser.Expression, args WhereArgs) error {
	if len(expression.Conditions()) != 1 {
		return fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(expression.Conditions()))
	}

	condition := expression.Conditions()[0]

	if condition.Type() != parser.ValueCondition && condition.Type() != parser.LogicalCondition {
		return fmt.Errorf("can only handle condition type of LogicalCondition or ValueCondition, have: %s", condition.Type())
	}

	lhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, condition.LHS)
	rhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, condition.RHS)
	var operator ActionOperator

	lhsOperandType, err := lhsResolver.GetOperandType()
	if err != nil {
		return fmt.Errorf("cannot resolve operand type of LHS operand")
	}

	if condition.Type() == parser.ValueCondition {
		if lhsOperandType != proto.Type_TYPE_BOOL {
			return fmt.Errorf("single operands in a value condition must be of type boolean")
		}

		// A value condition only has one operand in the expression,
		// for example, permission(expression: ctx.isAuthenticated),
		// so we must set the operator and RHS value (== true) ourselves.
		operator = Equals
	} else {
		operator, err = expressionOperatorToActionOperator(condition.Operator.ToString())
		if err != nil {
			return err
		}
	}

	lhsOperand, err := lhsResolver.generateQueryOperand(operator, args, query.writeValues)
	if err != nil {
		return err
	}

	lhsJoins, err := lhsResolver.generateJoins()
	if err != nil {
		return err
	}

	for _, j := range lhsJoins {
		query.Join(j)
	}

	var rhsOperand *QueryOperand
	if condition.Type() == parser.ValueCondition {
		// A value condition only has one operand in the expression.
		// If the expression is a value condition, then we bake in the RHS argument (== true) ourselves.
		rhsOperand = Value(true)
	} else {

		rhsOperand, err = rhsResolver.generateQueryOperand(operator, args, query.writeValues)
		if err != nil {
			return err
		}

		rhsJoins, err := lhsResolver.generateJoins()
		if err != nil {
			return err
		}

		for _, j := range rhsJoins {
			query.Join(j)
		}
	}

	query.Where(lhsOperand, operator, rhsOperand)

	return nil
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
		fragmentCount := len(operand.Ident.Fragments)
		modelTarget := strcase.ToCamel(operand.Ident.Fragments[0].Fragment)

		if fragmentCount > 2 {
			for i := 1; i < fragmentCount-1; i++ {
				field := proto.FindField(schema.Models, strcase.ToCamel(modelTarget), operand.Ident.Fragments[i].Fragment)
				modelTarget = field.Type.ModelName.Value
			}
		}

		fieldName := operand.Ident.Fragments[fragmentCount-1].Fragment
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
) (any, error) {

	operandType, err := resolver.GetOperandType()
	if err != nil {
		return nil, err
	}

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
		return runtimectx.GetNow(), nil
	case resolver.operand.Type() == parser.TypeArray:
		return nil, fmt.Errorf("cannot yet handle operand of type non-literal array")
	default:
		return nil, fmt.Errorf("cannot handle operand of unknown type")

	}
}

// Generates a single gorm-compatible operand and argument list:
//   - The gorm operand is used to build a template for use in gorm.DB.Where(template, args).
//   - The query arguments are used to populate the gorm operand in gorm.DB.Where(template, args).
func (resolver *OperandResolver) generateQueryOperand(
	operator ActionOperator,
	args map[string]any,
	writeValues map[string]any) (*QueryOperand, error) {

	// ? is the gorm template token which is replaced with values when populating the operand arguments when calling gorm.DB.Where(template, args).
	var queryOperand *QueryOperand

	if !resolver.IsDatabaseColumn() {
		value, err := resolver.ResolveValue(args, writeValues)
		if err != nil {
			return nil, err
		}

		if value == nil {
			queryOperand = Null()
		} else {
			switch operator {
			case StartsWith:
				value = value.(string) + "%%"
			case EndsWith:
				value = "%%" + value.(string)
			case Contains, NotContains:
				value = "%%" + value.(string) + "%%"
			}

			queryOperand = Value(value)
		}
	} else {
		// Generate the table's column name from the fragments (e.g. post.sub_title)
		// And replace the ? gorm template token with the column name
		var databaseField string
		model := strcase.ToCamel(resolver.operand.Ident.Fragments[0].Fragment)
		fragmentCount := len(resolver.operand.Ident.Fragments)
		databaseTable := strcase.ToSnake(resolver.operand.Ident.Fragments[0].Fragment)

		for i := 1; i < fragmentCount; i++ {
			currentFragment := resolver.operand.Ident.Fragments[i].Fragment

			if !proto.ModelHasField(resolver.schema, model, currentFragment) {
				return nil, fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
			}

			if i < fragmentCount-1 {
				// We know that the current fragment is a related table because it's not the last fragment
				relatedModelField := proto.FindField(resolver.schema.Models, model, currentFragment)
				relatedTable := strcase.ToSnake(relatedModelField.Type.ModelName.Value)

				model = relatedModelField.Type.ModelName.Value
				databaseTable = relatedTable
			} else {
				databaseField = strcase.ToSnake(currentFragment)
			}
		}

		queryOperand = Column(databaseTable, databaseField)
	}

	return queryOperand, nil
}

// Generates JOIN statements based on the operand of an expression.
func (resolver *OperandResolver) generateJoins() (joins []string, err error) {
	if resolver.IsDatabaseColumn() {
		databaseTable := strcase.ToSnake(resolver.operand.Ident.Fragments[0].Fragment)
		model := strcase.ToCamel(resolver.operand.Ident.Fragments[0].Fragment)
		fragmentCount := len(resolver.operand.Ident.Fragments)

		for i := 1; i < fragmentCount; i++ {
			currentFragment := resolver.operand.Ident.Fragments[i].Fragment

			if !proto.ModelHasField(resolver.schema, model, currentFragment) {
				return nil, fmt.Errorf("this model: %s, does not have a field of name: %s", model, currentFragment)
			}

			if i < fragmentCount-1 {
				// We know that the current fragment is a related table because it's not the last fragment
				relatedModelField := proto.FindField(resolver.schema.Models, model, currentFragment)
				relatedTable := strcase.ToSnake(relatedModelField.Type.ModelName.Value)

				primaryKeyDbColumn := "id"

				var join string
				if relatedModelField.ForeignKeyFieldName != nil {
					// M:1
					foreignKeyDbColumn := strcase.ToSnake(relatedModelField.ForeignKeyFieldName.Value)

					join = fmt.Sprintf("INNER JOIN %s ON %s = %s",
						relatedTable,
						fmt.Sprintf("%s.%s", databaseTable, foreignKeyDbColumn),
						fmt.Sprintf("%s.%s", relatedTable, primaryKeyDbColumn))
				} else {
					// 1:M
					fkModel := proto.FindModel(resolver.schema.Models, relatedModelField.Type.ModelName.Value)
					fkField, found := lo.Find(fkModel.Fields, func(field *proto.Field) bool {
						return field.Type.Type == proto.Type_TYPE_MODEL && field.Type.ModelName.Value == model
					})
					if !found {
						return nil, fmt.Errorf("no foreign key field found on related model %s", model)
					}

					foreignKeyDbColumn := strcase.ToSnake(fkField.ForeignKeyFieldName.Value)

					join = fmt.Sprintf("INNER JOIN %s ON %s = %s",
						relatedTable,
						fmt.Sprintf("%s.%s", databaseTable, primaryKeyDbColumn),
						fmt.Sprintf("%s.%s", relatedTable, foreignKeyDbColumn))
				}

				//if !lo.Contains(joins, join) {
				joins = append(joins, join)
				//}

				databaseTable = relatedTable
				model = relatedModelField.Type.ModelName.Value
			}
		}
	}

	return joins, nil
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
