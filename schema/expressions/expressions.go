package expressions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/util/collection"
)

type Expression struct {
	node.Node

	Or []*OrExpression `@@ ("or" @@)*`
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

	LHS      *Value `@@`
	Operator string `( @( "=" "=" | "!" "=" | ">" "=" | "<" "=" | ">" | "<" | "not" "in" | "in" | "+" "=" | "-" "=" | "=")`
	RHS      *Value `@@ )?`
}

func (condition *Condition) ToString() string {
	result := ""

	if condition == nil {
		return "oh no"
	}
	if condition.LHS != nil {
		result += condition.LHS.ToString()
	}
	result += condition.Operator
	if condition.RHS != nil {
		result += condition.RHS.ToString()
	}

	return result
}

type Value struct {
	node.Node

	Number *int64   `  @Int`
	String *string  `| @String`
	Null   bool     `| @"null"`
	True   bool     `| @"true"`
	False  bool     `| @"false"`
	Array  *Array   `| @@`
	Ident  []string `| ( @Ident ( "." @Ident )* )`
}

type Array struct {
	node.Node

	Values []*Value `"[" @@ ( "," @@ )* "]"`
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

			if op == "" {
				continue
			}

			result += " "

			// special case for "not in"
			if op == "notin" {
				result += "not in"
			} else {
				result += op
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

func ToValue(expr *Expression) (*Value, error) {
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
	if cond.Operator != "" {
		return nil, ErrNotValue
	}

	return cond.LHS, nil
}

func IsEquality(expr *Expression) bool {
	v, _ := ToEqualityCondition(expr)
	return v != nil
}

var ErrNotEquality = errors.New("expression does not check for equality")

var equalityOperators = []string{"==", "<", ">", ">=", "<=", "!=", "in", "notin", "contains"}

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

	if !collection.Contains(equalityOperators, cond.Operator) {
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
	if cond.Operator != "=" {
		return nil, ErrNotAssignment
	}

	return cond, nil
}

func (v *Value) ToString() string {
	switch {
	case v.Number != nil:
		return fmt.Sprintf("%d", *v.Number)
	case v.String != nil:
		return *v.String
	case v.Null:
		return "null"
	case v.False:
		return "false"
	case v.True:
		return "true"
	case v.Array != nil:
		r := "["
		for i, el := range v.Array.Values {
			if i > 0 {
				r += ", "
			}
			r += el.ToString()
		}
		return r + "]"
	case len(v.Ident) > 0:
		return strings.Join(v.Ident, ".")
	default:
		return ""
	}
}
