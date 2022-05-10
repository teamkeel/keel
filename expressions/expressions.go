package expressions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Expression struct {
	Pos lexer.Position

	Or []*OrExpression `@@ ("or" @@)*`
}

type OrExpression struct {
	Pos lexer.Position

	And []*ConditionWrap `@@ ("and" @@)*`
}

type ConditionWrap struct {
	Pos lexer.Position

	Expression *Expression `( "(" @@ ")"`
	Condition  *Condition  `| @@ )`
}

type Condition struct {
	Pos lexer.Position

	LHS      *Value `@@`
	Operator string `( @( "=" "=" | "!" "=" | ">" "=" | "<" "=" | ">" | "<" | "not" "in" | "in" | "+" "=" | "-" "=" | "=" )`
	RHS      *Value `@@ )?`
}

type Value struct {
	Pos lexer.Position

	Number *int64   `  @Int`
	String *string  `| @String`
	Null   bool     `| @"null"`
	True   bool     `| @"true"`
	False  bool     `| @"false"`
	Array  *Array   `| @@`
	Ident  []string `| ( @Ident ( "." @Ident )* )`
}

type Array struct {
	Pos lexer.Position

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

			result += valueToString(andExpr.Condition.LHS)

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

			result += " " + valueToString(andExpr.Condition.RHS)
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

func valueToString(v *Value) string {
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
			r += valueToString(el)
		}
		return r + "]"
	case len(v.Ident) > 0:
		return strings.Join(v.Ident, ".")
	default:
		return ""
	}
}
