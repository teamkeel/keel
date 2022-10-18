package actions

import (
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// DefaultApplyImplicitFilters considers all the implicit inputs expected for
// the given operation, and captures the targeted field. It then captures the corresponding value
// operand value provided by the given request arguments, and adds a Where clause to the
// query field in the given scope, using a hard-coded equality operator.
func DefaultApplyImplicitFilters(scope *Scope, args RequestArguments) error {
	for _, input := range scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			return fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", fieldName, args)
		}

		if err := addImplicitFilter(scope, input, OperatorEquals, value); err != nil {
			return err
		}
	}

	return nil
}

func DefaultApplyExplicitFilters(scope *Scope, args RequestArguments) error {
	operation := scope.operation

	for _, where := range operation.WhereExpressions {
		expr, err := parser.ParseExpression(where.Source)

		if err != nil {
			return err
		}

		// todo: look into refactoring interpretExpressionField to support handling
		// of multiple conditions in an expression and also literal values
		field, err := interpretExpressionField(expr, operation, scope.schema)
		if err != nil {
			return err
		}

		conditions := expr.Conditions()

		condition := conditions[0]

		match, ok := args[condition.RHS.Ident.ToString()]

		if !ok {
			return fmt.Errorf("argument not provided for %s", field.Name)
		}

		addExplicitFilter(scope, field, OperatorEquals, match)
	}

	return nil
}

// todo:
// addExplicitFilter and  addImplicitFilter should be the same method
// we just need to find a common syntax for expressing operators from grapqhql implicit operators or expression operators

// addImplicitFilter adds Where clauses to the query field of the given scope, corresponding to
// the given input, the given operator, and using the given value as the operand.
func addImplicitFilter(scope *Scope, input *proto.OperationInput, operator Operator, value any) error {

	inputType := input.Type.Type
	columnName := input.Target[0]

	switch operator {
	case OperatorEquals:
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))

		if inputType == proto.Type_TYPE_DATE || inputType == proto.Type_TYPE_DATETIME || inputType == proto.Type_TYPE_TIMESTAMP {
			time, err := parseTimeOperand(value, inputType)

			if err != nil {
				return err
			}

			scope.query = scope.query.Where(w, time)
		} else {
			scope.query = scope.query.Where(w, value)
		}
	case OperatorStartsWith:
		operandStr, ok := value.(string)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandStr+"%%")
	case OperatorEndsWith:
		operandStr, ok := value.(string)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, "%%"+operandStr)
	case OperatorContains:
		operandStr, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, "%%"+operandStr+"%%")
	case OperatorOneOf:
		operandStrings, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("cannot cast this: %v to a []interface{}", value)
		}

		w := fmt.Sprintf("%s in ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandStrings)
	case OperatorLessThan:
		operandInt, ok := value.(int)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}

		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)
	case OperatorLessThanEquals:
		operandInt, ok := value.(int)

		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}

		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)
	case OperatorGreaterThan:
		operandInt, ok := value.(int)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}
		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)

	case OperatorGreaterThanEquals:
		operandInt, ok := value.(int)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to an int", value)
		}
		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandInt)

	case OperatorBefore:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))

		scope.query = scope.query.Where(w, operandTime)
	case OperatorAfter:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))

		scope.query = scope.query.Where(w, operandTime)
	case OperatorOnOrBefore:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandTime)
	case OperatorOnOrAfter:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return err
		}

		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandTime)
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

func addExplicitFilter(scope *Scope, field *proto.Field, operator Operator, value any) error {
	if operator != OperatorEquals {
		return fmt.Errorf("operator %s not yet supported", operator)
	}

	w := fmt.Sprintf("%s = ?", strcase.ToSnake(field.Name))
	scope.query = scope.query.Where(w, value)

	return nil
}
