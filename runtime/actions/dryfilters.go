package actions

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// DRYApplyExplicitFilters marries up the given operation's Where expressions, with operands
// provided in the given request arguments. It then adds corresponding Where clauses to the
// query field in the given scope object.
func DRYApplyExplicitFilters(scope *Scope, args RequestArguments) error {

	for _, where := range scope.operation.WhereExpressions {
		expr, err := parser.ParseExpression(where.Source)

		if err != nil {
			return err
		}

		// todo: look into refactoring interpretExpressionField to support handling
		// of multiple conditions in an expression and also literal values
		field, err := interpretExpressionField(expr, scope.operation, scope.schema)
		if err != nil {
			return err
		}

		// @where(expression: post.title == coolTitle and post.title == somethingElse)

		conditions := expr.Conditions()

		condition := conditions[0]

		match, ok := args[condition.RHS.Ident.ToString()]

		if !ok {
			return fmt.Errorf("argument not provided for %s", field.Name)
		}

		DRYaddExplicitFilter(scope, field.Name, condition.Operator.Symbol, match)
	}

	return nil
}

// DRYaddExplicitFilter updates the query inside the given scope with
// Where clauses that represent filters specified by an operation's Where expression,
// using the given value as an operand.
func DRYaddExplicitFilter(scope *Scope, fieldName string, operator string, value any) error {
	// todo: support all operator types
	if operator != parser.OperatorEquals {
		panic("this operator is not supported yet...")
	}

	w := fmt.Sprintf("%s = ?", strcase.ToSnake(fieldName))
	scope.query = scope.query.Where(w, value)

	return nil
}

// todo:
// addExplicitFilter and  addImplicitFilter should be the same method
// we just need to find a common syntax for expressing operators from grapqhql implicit operators or expression operators

// DRYaddImplicitFilter adds Where clauses to the query field of the given scope, corresponding to
// the given input, the given operator, and using the given value as the operand.
func DRYaddImplicitFilter(scope *Scope, input *proto.OperationInput, operator Operator, value any) error {

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
