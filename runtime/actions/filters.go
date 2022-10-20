package actions

import (
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
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

		if err := addWhereClauseForConditional(scope, fieldName, input, Equals, value); err != nil {
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

		if err := addWhereClauseForConditional(scope, field.Name, protoInput, operator, operandValue); err != nil {
			return err
		}
	}

	return nil
}

// addWhereClauseForConditional adds Where clauses to the query field of the given
// scope, corresponding to the given input, the given operator, and using the given value as
// the operand.
func addWhereClauseForConditional(scope *Scope, columnName string, input *proto.OperationInput, operator ActionOperator, value any) error {
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
		} else {
			scope.query = scope.query.Where(w, value)
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
