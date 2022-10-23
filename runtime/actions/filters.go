package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// DefaultApplyImplicitFilters considers all the implicit inputs expected for
// the given operation, and captures the targeted field. It then captures the corresponding value
// operand value provided by the given request arguments, and adds a Where clause to the
// query field in the given scope, using a hard-coded equality operator.
func DefaultApplyImplicitFilters(scope *Scope, args RequestArguments) error {
	for _, input := range scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT || input.Mode == proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			return fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", fieldName, args)
		}

		if err := addFilter(scope, fieldName, input, Equals, value); err != nil {
			return err
		}
	}

	return nil
}

func DefaultApplyExplicitFilters(scope *Scope, args RequestArguments) error {
	operation := scope.operation

	for _, where := range operation.WhereExpressions {
		expr, err := parser.ParseExpression(where.Source) // E.g. post.title == requiredTitle

		if err != nil {
			return err
		}

		// Map the "requiredTitle" part to the correct model field - e.g. "the title" field, and
		// capture the "==" part as a machine-readable ActionOperator type.
		field, operator, err := interpretExpressionField(expr, operation, scope.schema)
		if err != nil {
			return err
		}

		conditions := expr.Conditions()

		// todo: look into refactoring interpretExpressionField to support handling
		// of multiple conditions in an expression and also literal values
		condition := conditions[0]

		argName := condition.RHS.Ident.ToString() // E.g. "requiredTitle"

		operandValue, ok := args[argName]
		if !ok {
			return fmt.Errorf("argument not provided for %s", field.Name)
		}

		// The function we are going to call, requires access to the corresponding Input object.
		protoInput, ok := lo.Find(scope.operation.Inputs, func(input *proto.OperationInput) bool {
			return input.Name == argName
		})
		if !ok {
			return fmt.Errorf("cannot find input of name: %s", argName)
		}

		if err := addFilter(scope, field.Name, protoInput, operator, operandValue); err != nil {
			return err
		}
	}

	return nil
}

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

// TODO: refactor and decouple this function from columnName and Input?
// Any operand can be literal, input, context or database field.
// @where doesn't only operate on database columns
func toSql(scope *Scope, condition *parser.Condition, operator ActionOperator, data map[string]any) (result string, queryArguments []any) {

	if condition.Type() != parser.ValueCondition && condition.Type() != parser.LogicalCondition {
		return "", nil // fmt.Errorf("can only handle condition type of LogicalCondition, have: %s", condition.Type())
	}

	var lhsOperandType, rhsOperandType proto.Type

	lhsOperand := condition.LHS
	rhsOperand := condition.RHS

	lhsResolver := NewOperandResolver(scope.context, lhsOperand, scope.operation, scope.schema)
	rhsResolver := NewOperandResolver(scope.context, rhsOperand, scope.operation, scope.schema)
	lhsOperandType, _ = lhsResolver.GetOperandType()

	if condition.Type() == parser.ValueCondition {
		if lhsOperandType != proto.Type_TYPE_BOOL {
			//todo: err - must be a bool
		}

		// A value condition only has one operand in the expression,
		// so we must set the operator and RHS value (= true) ourselves.
		rhsOperandType = lhsOperandType
		operator = Equals
	} else {
		rhsOperandType, _ = rhsResolver.GetOperandType()
	}

	var query string
	var queryArgs []any

	lhsSqlOperand := "?"
	rhsSqlOperand := "?"

	if !lhsResolver.IsModelField() {
		lhsValue, _ := lhsResolver.ResolveValue(data, lhsOperandType)
		// todo: date and time parsing
		queryArgs = append(queryArgs, lhsValue)
	} else {
		modelTarget := strcase.ToSnake(lhsResolver.operand.Ident.Fragments[0].Fragment)
		fieldName := strcase.ToSnake(lhsResolver.operand.Ident.Fragments[1].Fragment)
		lhsSqlOperand = fmt.Sprintf("%s.%s", modelTarget, fieldName)
	}

	if condition.Type() == parser.ValueCondition {
		queryArgs = append(queryArgs, true)
	} else if !rhsResolver.IsModelField() {
		rhsValue, _ := rhsResolver.ResolveValue(data, rhsOperandType)
		// todo: date and time parsing
		queryArgs = append(queryArgs, rhsValue)
	} else {
		modelTarget := strcase.ToSnake(rhsResolver.operand.Ident.Fragments[0].Fragment)
		fieldName := strcase.ToSnake(rhsResolver.operand.Ident.Fragments[1].Fragment)
		rhsSqlOperand = fmt.Sprintf("%s.%s", modelTarget, fieldName)
	}

	switch operator {
	case Equals:
		if rhsOperandType == proto.Type_TYPE_UNKNOWN {
			query = fmt.Sprintf("%s IS %s", lhsSqlOperand, rhsSqlOperand)
		} else {
			query = fmt.Sprintf("%s = %s", lhsSqlOperand, rhsSqlOperand)
		}
	case NotEquals:
		if rhsOperandType == proto.Type_TYPE_UNKNOWN {
			query = fmt.Sprintf("%s IS NOT %s", lhsSqlOperand, rhsSqlOperand)
		} else {
			query = fmt.Sprintf("%s != %s", lhsSqlOperand, rhsSqlOperand)
		}
	case StartsWith:
		query = fmt.Sprintf("%s LIKE %s%s", lhsSqlOperand, "%%", rhsSqlOperand)
	case EndsWith:
		query = fmt.Sprintf("%s LIKE %s%s", lhsSqlOperand, rhsSqlOperand, "%%")
	case Contains:
		query = fmt.Sprintf("%s LIKE %s%s%s", lhsSqlOperand, "%%", rhsSqlOperand, "%%")
	case OneOf:
		query = fmt.Sprintf("%s in %s", lhsSqlOperand, rhsSqlOperand)
	case LessThan:
		query = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case LessThanEquals:
		query = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThan:
		query = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThanEquals:
		query = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
	case Before:
		query = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case After:
		query = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrBefore:
		query = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrAfter:
		query = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
		//default:
		//	return fmt.Errorf("operator: %v is not yet supported", operator)
	}

	return query, queryArgs
}

// addFilter adds Where clauses to the query field of the given
// scope, corresponding to the given input, the given operator, and using the given value as
// the operand.
func addFilter(scope *Scope, columnName string, input *proto.OperationInput, operator ActionOperator, value any) error {
	inputType := input.Type.Type

	// todo: the use of parseTimeOperand is conflicting with our current integration test framework, as this
	// generates typescript that expects the input objects to be native javascript Date/Time types.
	// See for example integration/operation_list_explicit.
	switch operator {
	case Equals:
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))

		if inputType == proto.Type_TYPE_DATE || inputType == proto.Type_TYPE_DATETIME || inputType == proto.Type_TYPE_TIMESTAMP {
			time, err := parseTimeOperand(value, inputType)

			if err != nil {
				return err
			}

			scope.query = scope.query.Where(w, time)
			scope.permissionQuery = scope.permissionQuery.Where(w, time)
		} else {
			scope.query = scope.query.Where(w, value)
			scope.permissionQuery = scope.permissionQuery.Where(w, value)
		}
	case NotEquals:
		w := fmt.Sprintf("%s != ?", strcase.ToSnake(columnName))

		if inputType == proto.Type_TYPE_DATE || inputType == proto.Type_TYPE_DATETIME || inputType == proto.Type_TYPE_TIMESTAMP {
			time, err := parseTimeOperand(value, inputType)

			if err != nil {
				return err
			}

			scope.query = scope.query.Where(w, time)
		} else {
			scope.query = scope.query.Where(w, value)
		}

	case StartsWith:
		operandStr, ok := value.(string)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandStr+"%%")
	case EndsWith:
		operandStr, ok := value.(string)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, "%%"+operandStr)
	case Contains:
		operandStr, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, "%%"+operandStr+"%%")
	case OneOf:
		operandStrings, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("cannot cast this: %v to a []interface{}", value)
		}

		w := fmt.Sprintf("%s in ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandStrings)
	case LessThan:
		operandInt, ok := value.(int)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}

		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)
	case LessThanEquals:
		operandInt, ok := value.(int)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}

		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)
	case GreaterThan:
		operandInt, ok := value.(int)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}
		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)

	case GreaterThanEquals:
		operandInt, ok := value.(int)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}
		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)

	case Before:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))

		scope.query = scope.query.Where(w, operandTime)
	case After:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))

		scope.query = scope.query.Where(w, operandTime)
	case OnOrBefore:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandTime)
	case OnOrAfter:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandTime)
		scope.permissionQuery = scope.permissionQuery.Where(w, operandTime)
	default:
		return fmt.Errorf("operator: %v is not yet supported", operator)
	}

	return nil
}

// parseTimeOperand extract and parses time for date/time based operators
// Supports timestamps passed in map[seconds:int] and dates passesd as map[day:int month:int year:int]
func parseTimeOperand(operand any, inputType proto.Type) (t *time.Time, err error) {
	operandMap, ok := operand.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot cast this: %v to a map[string]interface{}", operand)
	}

	switch inputType {
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		seconds := operandMap["seconds"]
		secondsInt, ok := seconds.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to int", seconds)
		}
		unix := time.Unix(int64(secondsInt), 0).UTC()
		t = &unix

	case proto.Type_TYPE_DATE:
		day := operandMap["day"]
		month := operandMap["month"]
		year := operandMap["year"]

		dayInt, ok := day.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast days: %v to int", day)
		}
		monthInt, ok := month.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast month: %v to int", month)
		}
		yearInt, ok := year.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast year: %v to int", year)
		}

		time, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-%02d", yearInt, monthInt, dayInt))
		if err != nil {
			return nil, fmt.Errorf("cannot parse date %s", err)
		}
		t = &time

	default:
		return nil, fmt.Errorf("unknown time field type")
	}

	return t, nil
}
