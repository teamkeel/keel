package actions

import (
	"fmt"

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
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT || input.Mode == proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			return fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", fieldName, args)
		}

		if err := addFilter(scope, fieldName, Equals, value); err != nil {
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

		if err := addFilter(scope, field.Name, operator, operandValue); err != nil {
			return err
		}
	}

	return nil
}

// addFilter adds Where clauses to the query field of the given
// scope, corresponding to the given input, the given operator, and using the given value as
// the operand.
func addFilter(scope *Scope, columnName string, operator ActionOperator, value any) error {
	switch operator {
	case Equals:
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, value)
	case NotEquals:
		w := fmt.Sprintf("%s != ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, value)
	case StartsWith:
		operandStr, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, operandStr+"%")
	case EndsWith:
		operandStr, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, "%"+operandStr)
	case Contains:
		operandStr, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot cast this: %v to a string", value)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, "%"+operandStr+"%")
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
		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, value)
	case After:
		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, value)
	case OnOrBefore:
		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, value)
	case OnOrAfter:
		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		scope.query = scope.query.Where(w, value)
	default:
		return fmt.Errorf("operator: %v is not yet supported", operator)
	}

	return nil
}

// parseTimeOperand extract and parses time for date/time based operators
// Supports timestamps passed in map[seconds:int] and dates passesd as map[day:int month:int year:int]
// func parseTimeOperand(operand any, inputType proto.Type) (t *time.Time, err error) {
// 	operandMap, ok := operand.(map[string]interface{})
// 	if !ok {
// 		return nil, fmt.Errorf("cannot cast this: %v to a map[string]interface{}", operand)
// 	}

// 	switch inputType {
// 	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
// 		seconds := operandMap["seconds"]
// 		secondsInt, ok := seconds.(int)
// 		if !ok {
// 			return nil, fmt.Errorf("cannot cast this: %v to int", seconds)
// 		}
// 		unix := time.Unix(int64(secondsInt), 0).UTC()
// 		t = &unix

// 	case proto.Type_TYPE_DATE:
// 		day := operandMap["day"]
// 		month := operandMap["month"]
// 		year := operandMap["year"]

// 		dayInt, ok := day.(int)
// 		if !ok {
// 			return nil, fmt.Errorf("cannot cast days: %v to int", day)
// 		}
// 		monthInt, ok := month.(int)
// 		if !ok {
// 			return nil, fmt.Errorf("cannot cast month: %v to int", month)
// 		}
// 		yearInt, ok := year.(int)
// 		if !ok {
// 			return nil, fmt.Errorf("cannot cast year: %v to int", year)
// 		}

// 		time, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-%02d", yearInt, monthInt, dayInt))
// 		if err != nil {
// 			return nil, fmt.Errorf("cannot parse date %s", err)
// 		}
// 		t = &time

// 	default:
// 		return nil, fmt.Errorf("unknown time field type")
// 	}

// 	return t, nil
// }
