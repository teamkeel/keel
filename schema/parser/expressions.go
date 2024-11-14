package parser

import (
	"errors"

	"github.com/alecthomas/participle/v2"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/node"
)

// type Expression struct {
// 	node.Node

// 	Or []*OrExpression `@@ ("or" @@)*`
// }

// func (e *Expression) Conditions() []*Condition {
// 	conds := []*Condition{}

// 	for _, or := range e.Or {
// 		for _, and := range or.And {
// 			conds = append(conds, and.Condition)

// 			if and.Expression != nil {
// 				conds = append(conds, and.Expression.Conditions()...)
// 			}
// 		}
// 	}

// 	ret := []*Condition{}
// 	for _, cond := range conds {
// 		if cond != nil {
// 			ret = append(ret, cond)
// 		}
// 	}
// 	return ret
// }

// type OrExpression struct {
// 	node.Node

// 	And []*ConditionWrap `@@ ("and" @@)*`
// }

// type ConditionWrap struct {
// 	node.Node

// 	Expression *Expression `( "(" @@ ")"`
// 	Condition  *Condition  `| @@ )`
// }

// type Condition struct {
// 	node.Node

// 	LHS      *Operand  `@@`
// 	Operator *Operator `(@@`
// 	RHS      *Operand  `@@ )?`
// }

var (
	AssignmentCondition = "assignment"
	LogicalCondition    = "logical"
	ValueCondition      = "value"
	UnknownCondition    = "unknown"
)

// func (c *Condition) Type() string {
// 	if c.Operator == nil && c.RHS == nil && c.LHS != nil {
// 		return ValueCondition
// 	}

// 	if lo.Contains(AssignmentOperators, c.Operator.Symbol) {
// 		return AssignmentCondition
// 	}

// 	if lo.Contains(LogicalOperators, c.Operator.Symbol) {
// 		return LogicalCondition
// 	}

// 	return UnknownCondition
// }

type Operator struct {
	node.Node

	// Todo need to figure out how we can share with the consts below
	Symbol string `@( "=" "=" | "!" "=" | ">" "=" | "<" "=" | ">" | "<" | "not" "in" | "in" | "+" "=" | "-" "=" | "=" | "or" | "and" | "+" | "-" | "*" | "/" )`
}

func (o *Operator) ToString() string {
	if o == nil {
		return ""
	}

	return o.Symbol
}

var (
	OperatorEquals               = "=="
	OperatorAssignment           = "="
	OperatorNotEquals            = "!="
	OperatorGreaterThanOrEqualTo = ">="
	OperatorLessThanOrEqualTo    = "<="
	OperatorLessThan             = "<"
	OperatorGreaterThan          = ">"
	OperatorIn                   = "in"
	OperatorNotIn                = "notin"
	OperatorIncrement            = "+="
	OperatorDecrement            = "-="
)

var AssignmentOperators = []string{
	OperatorAssignment,
}

var LogicalOperators = []string{
	OperatorEquals,
	OperatorNotEquals,
	OperatorGreaterThan,
	OperatorGreaterThanOrEqualTo,
	OperatorLessThan,
	OperatorLessThanOrEqualTo,
	OperatorIn,
	OperatorNotIn,
}

// func (condition *Condition) ToString() string {
// 	result := ""

// 	if condition == nil {
// 		panic("condition is nil")
// 	}
// 	if condition.LHS != nil {
// 		result += condition.LHS.ToString()
// 	}

// 	if condition.Operator != nil && condition.RHS != nil {
// 		result += fmt.Sprintf(" %s ", condition.Operator.Symbol)
// 		result += condition.RHS.ToString()
// 	}

// 	return result
// }

func ParseExpression(source string) (*Expression, error) {
	parser, err := participle.Build[Expression]()
	if err != nil {
		return nil, err
	}

	expr, err := parser.ParseString("", source)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

// func (expr *Expression) ToString() (string, error) {
// 	result := ""

// 	for i, orExpr := range expr.Or {
// 		if i > 0 {
// 			result += " or "
// 		}

// 		for j, andExpr := range orExpr.And {
// 			if j > 0 {
// 				result += " and "
// 			}

// 			if andExpr.Expression != nil {
// 				r, err := andExpr.Expression.ToString()
// 				if err != nil {
// 					return result, err
// 				}
// 				result += "(" + r + ")"
// 				continue
// 			}

// 			result += andExpr.Condition.LHS.ToString()

// 			op := andExpr.Condition.Operator

// 			if op != nil && op.Symbol == "" {
// 				continue
// 			}

// 			// special case for "not in"
// 			if op != nil && op.Symbol == "notin" {
// 				result += " not in "
// 			} else if op != nil {
// 				result += fmt.Sprintf(" %s ", op.Symbol)
// 			}

// 			if andExpr.Condition.RHS != nil {
// 				result += andExpr.Condition.RHS.ToString()
// 			}
// 		}
// 	}

// 	return result, nil
// }

// func (expr *Expression) IsValue() bool {
// 	v, _ := expr.ToValue()
// 	return v != nil
// }

var ErrNotValue = errors.New("expression is not a single value")

func (expr *Expression) ToValue() (*Operand, error) {
	if expr.Operator != nil || expr.RHS != nil {
		return nil, ErrNotValue
	}

	if expr.LHS.Operator != nil || expr.LHS.Term != nil {
		return nil, ErrNotValue
	}

	if expr.LHS.Factor.Operand == nil {
		return nil, ErrNotValue
	}

	return expr.LHS.Factor.Operand, nil
}

var ErrNotAssignmentOperand = errors.New("assignment expression requires an ident LHS operand")
var ErrNotAssignmentOperator = errors.New("assignment expression is not using the correct assignment operator")
var ErrNotAssignmentExpression = errors.New("assignment expression does not have a RHS expression")

func (expr *Expression) IsAssignment() bool {
	_, _, err := expr.ToAssignmentExpression()
	return err == ErrNotAssignmentOperand || err == ErrNotAssignmentOperator || err == ErrNotAssignmentExpression
}

func (expr *Expression) ToAssignmentExpression() (*Operand, ExpressionPart, error) {
	if expr.LHS == nil || expr.LHS.Factor == nil || expr.LHS.Factor.Operand == nil {
		return nil, nil, ErrNotAssignmentOperand
	}

	if expr.LHS.Operator == nil || !lo.Contains(AssignmentOperators, expr.LHS.Operator.Symbol) {
		return nil, nil, ErrNotAssignmentOperator
	}

	if expr.LHS.Term == nil {
		return nil, nil, ErrNotAssignmentExpression
	}

	return expr.LHS.Factor.Operand, expr.LHS.Term, nil
}
