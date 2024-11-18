package parser

import (
	"errors"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/node"
)

var (
	AssignmentCondition = "assignment"
	LogicalCondition    = "logical"
	ValueCondition      = "value"
	UnknownCondition    = "unknown"
)

type Operator struct {
	node.Node

	// Todo need to figure out how we can share with the consts below
	Symbol string `@( "=" "=" | "!" "=" | ">" "=" | "<" "=" | ">" | "<" | "not" "in" | "in" | "+" "=" | "-" "=" | "=" | "or" | "and" | "+" | "-" | "*" | "/" )`
}

func (o *Operator) ToString() string {
	if o == nil {
		return ""
	}

	if o.Symbol == "and" {
		return "&&"
	}

	if o.Symbol == "or" {
		return "||"
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
