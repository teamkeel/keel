package expressions

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/util/collection"
)

type Expression struct {
	node.Node

	Or []*OrExpression `@@ ("or" @@)*`
}

func (e *Expression) Conditions() []*Condition {
	conds := []*Condition{}

	for _, or := range e.Or {
		for _, and := range or.And {
			conds = append(conds, and.Condition)

			if and.Expression != nil {
				conds = append(conds, and.Expression.Conditions()...)
			}
		}
	}

	ret := []*Condition{}
	for _, cond := range conds {
		if cond != nil {
			ret = append(ret, cond)
		}
	}
	return ret
}

type OrExpression struct {
	node.Node

	And []*ConditionWrap `@@ ("and" @@)*`
}

type ConditionWrap struct {
	node.Node

	Expression *Expression `( "(" @@ ")"`
	Condition  *Condition  `| @@ )`
}

type Condition struct {
	node.Node

	LHS      *Operand `@@`
	Operator Operator `( @@`
	RHS      *Operand `@@ )?`
}

var equalityOperators = []string{"==", "<", ">", ">=", "<=", "!=", "in", "notin", "contains"}

var (
	AssignmentCondition = "assignment"
	LogicalCondition    = "logical"
	ValueCondition      = "value"
)

func (c *Condition) Type() string {

	if collection.Contains(equalityOperators, c.Operator.Symbol) {
		return LogicalCondition

	} else if (c.LHS.False || c.LHS.True) && c.RHS == nil {
		return LogicalCondition
	} else if c.Operator.Symbol == "=" {
		return AssignmentCondition
	} else if c.Operator.Symbol == "" && c.RHS == nil && c.LHS != nil {
		return ValueCondition
	}

	panic("not a known condition type")
}

type Operator struct {
	Symbol string `@( "=" "=" | "!" "=" | ">" "=" | "<" "=" | ">" | "<" | "not" "in" | "in" | "+" "=" | "-" "=" | "=")`
}

func (o *Operator) ToString() string {
	if o == nil {
		return ""
	}

	return o.Symbol
}

var operators = map[string]string{
	"==":    "Equals",
	"=":     "Assignment",
	"!=":    "NotEquals",
	">=":    "GreaterThanOrEqual",
	"<=":    "LessThanOrEqual",
	"<":     "LessThan",
	">":     "GreaterThan",
	"in":    "In",
	"notin": "NotIn",
	"+=":    "Increment",
	"--":    "Decrement",
}

func (o *Operator) Name() string {
	if o.Symbol == "" {
		return ""
	}

	if operator, ok := operators[o.Symbol]; ok {
		return operator
	}

	return ""
}

// Returns the respective fragments (lhs, operator, rhs) of an expression
// For value expressions, operator and rhs will be nil
// For equality & assignment expressions, lhs, operator and rhs will be populated
func (condition *Condition) ToFragments() (*Operand, *Operator, *Operand) {
	if condition == nil {
		return nil, nil, nil
	}
	return condition.LHS, &condition.Operator, condition.RHS
}

func (condition *Condition) ToString() string {
	result := ""

	if condition == nil {
		panic("condition is nil")
	}
	if condition.LHS != nil {
		result += condition.LHS.ToString()
	}

	if condition.Operator.Symbol != "" && condition.RHS != nil {
		result += fmt.Sprintf(" %s ", condition.Operator.Symbol)
		result += condition.RHS.ToString()
	}

	return result
}

func Parse(source string) (*Expression, error) {
	parser, err := participle.Build(&Expression{})
	if err != nil {
		return nil, err
	}

	expr := &Expression{}
	err = parser.ParseString("", source, expr)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func ToString(expr *Expression) (string, error) {
	result := ""

	for i, orExpr := range expr.Or {

		if i > 0 {
			result += " or "
		}

		for j, andExpr := range orExpr.And {

			if j > 0 {
				result += " and "
			}

			if andExpr.Expression != nil {
				r, err := ToString(andExpr.Expression)
				if err != nil {
					return result, err
				}
				result += "(" + r + ")"
				continue
			}

			result += andExpr.Condition.LHS.ToString()

			op := andExpr.Condition.Operator

			if op.Symbol == "" {
				continue
			}

			result += " "

			// special case for "not in"
			if op.Symbol == "notin" {
				result += "not in"
			} else {
				result += op.Symbol
			}

			result += " " + andExpr.Condition.RHS.ToString()
		}
	}

	return result, nil
}

func IsValue(expr *Expression) bool {
	v, _ := ToValue(expr)
	return v != nil
}

var ErrNotValue = errors.New("expression is not a single value")

func ToValue(expr *Expression) (*Operand, error) {
	if len(expr.Or) > 1 {
		return nil, ErrNotValue
	}

	or := expr.Or[0]
	if len(or.And) > 1 {
		return nil, ErrNotValue
	}
	and := or.And[0]

	if and.Expression != nil {
		return nil, ErrNotValue
	}

	cond := and.Condition

	if cond.Operator.Symbol != "" {
		return nil, ErrNotValue
	}

	return cond.LHS, nil
}

func IsEquality(expr *Expression) bool {
	v, _ := ToEqualityCondition(expr)
	return v != nil
}

var ErrNotEquality = errors.New("expression does not check for equality")

func ToEqualityCondition(expr *Expression) (*Condition, error) {
	or := expr.Or[0]

	and := or.And[0]

	if and.Expression != nil {
		return nil, ErrNotEquality
	}

	cond := and.Condition

	if cond == nil {
		return nil, ErrNotEquality
	}

	if !collection.Contains(equalityOperators, cond.Operator.Symbol) {
		return nil, ErrNotEquality
	}

	if cond.LHS == nil || cond.RHS == nil {
		return nil, ErrNotEquality
	}
	return cond, nil
}

var ErrNotAssignment = errors.New("expression is not using an assignment, e.g. a = b")

func IsAssignment(expr *Expression) bool {
	v, _ := ToAssignmentCondition(expr)
	return v != nil
}

func ToAssignmentCondition(expr *Expression) (*Condition, error) {
	if len(expr.Or) > 1 {
		return nil, ErrNotAssignment
	}

	or := expr.Or[0]
	if len(or.And) > 1 {
		return nil, ErrNotAssignment
	}

	and := or.And[0]

	if and.Expression != nil {
		return nil, ErrNotAssignment
	}
	cond := and.Condition
	if cond.Operator.Symbol != "=" {
		return nil, ErrNotAssignment
	}

	return cond, nil
}
