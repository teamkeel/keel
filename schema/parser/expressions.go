package parser

// import (
// 	"errors"
// 	"fmt"

// 	"github.com/alecthomas/participle/v2"
// 	"github.com/samber/lo"
// 	"github.com/teamkeel/keel/schema/node"
// )

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

// var (
// 	AssignmentCondition = "assignment"
// 	LogicalCondition    = "logical"
// 	ValueCondition      = "value"
// 	UnknownCondition    = "unknown"
// )

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

// type Operator struct {
// 	node.Node

// 	// Todo need to figure out how we can share with the consts below
// 	Symbol string `@( "=" "=" | "!" "=" | ">" "=" | "<" "=" | ">" | "<" | "not" "in" | "in" | "+" "=" | "-" "=" | "=")`
// }

// func (o *Operator) ToString() string {
// 	if o == nil {
// 		return ""
// 	}

// 	return o.Symbol
// }

// var (
// 	OperatorEquals               = "=="
// 	OperatorAssignment           = "="
// 	OperatorNotEquals            = "!="
// 	OperatorGreaterThanOrEqualTo = ">="
// 	OperatorLessThanOrEqualTo    = "<="
// 	OperatorLessThan             = "<"
// 	OperatorGreaterThan          = ">"
// 	OperatorIn                   = "in"
// 	OperatorNotIn                = "notin"
// 	OperatorIncrement            = "+="
// 	OperatorDecrement            = "-="
// )

// var AssignmentOperators = []string{
// 	OperatorAssignment,
// }

// var LogicalOperators = []string{
// 	OperatorEquals,
// 	OperatorNotEquals,
// 	OperatorGreaterThan,
// 	OperatorGreaterThanOrEqualTo,
// 	OperatorLessThan,ƒ
// 	OperatorLessThanOrEqualTo,
// 	OperatorIn,
// 	OperatorNotIn,
// }

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

// func ParseExpression(source string) (*Expression, error) {
// 	parser, err := participle.Build[Expression]()
// 	if err != nil {
// 		return nil, err
// 	}

// 	expr, err := parser.ParseString("", source)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return expr, nil
// }

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

// var ErrNotValue = errors.New("expression is not a single value")

// func (expr *Expression) ToValue() (*Operand, error) {
// 	if len(expr.Or) > 1 {
// 		return nil, ErrNotValue
// 	}

// 	or := expr.Or[0]
// 	if len(or.And) > 1 {
// 		return nil, ErrNotValue
// 	}
// 	and := or.And[0]

// 	if and.Expression != nil {
// 		return nil, ErrNotValue
// 	}

// 	cond := and.Condition

// 	if cond.Operator != nil && cond.Operator.Symbol != "" {
// 		return nil, ErrNotValue
// 	}

// 	return cond.LHS, nil
// }

// var ErrNotAssignment = errors.New("expression is not using an assignment, e.g. a = b")

// func (expr *Expression) IsAssignment() bool {
// 	v, _ := expr.ToAssignmentCondition()
// 	return v != nil
// }

// func (expr *Expression) ToAssignmentCondition() (*Condition, error) {
// 	if len(expr.Or) > 1 {
// 		return nil, ErrNotAssignment
// 	}

// 	or := expr.Or[0]
// 	if len(or.And) > 1 {
// 		return nil, ErrNotAssignment
// 	}

// 	and := or.And[0]

// 	if and.Expression != nil {
// 		return nil, ErrNotAssignment
// 	}
// 	cond := and.Condition

// 	if cond.Operator == nil || !lo.Contains(AssignmentOperators, cond.Operator.Symbol) {
// 		return nil, ErrNotAssignment
// 	}

// 	return cond, nil
// }
