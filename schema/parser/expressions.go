package parser

import (
	"errors"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/google/cel-go/cel"
)

var (
	AssignmentCondition = "assignment"
	LogicalCondition    = "logical"
	ValueCondition      = "value"
	UnknownCondition    = "unknown"
)

// type Operator struct {
// 	node.Node

// 	// Todo need to figure out how we can share with the consts below
// 	Symbol string `@( "=" "=" | "!" "=" | ">" "=" | "<" "=" | ">" | "<" | "not" "in" | "in" | "+" "=" | "-" "=" | "=" | "or" | "and" )`
// }

// func (o *Operator) ToString() string {
// 	if o == nil {
// 		return ""
// 	}

// 	if o.Symbol == "and" {
// 		return "&&"
// 	}

// 	if o.Symbol == "or" {
// 		return "||"
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
// 	OperatorLessThan,
// 	OperatorLessThanOrEqualTo,
// 	OperatorIn,
// 	OperatorNotIn,
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

var ErrNotValue = errors.New("expression is not a single value")

//func (expr *Expression) ToValue() (any, error) {

func (expr *Expression) ToValue() (any, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return nil, errors.New("failed to parse expression")
	}

	ast, issues := env.Parse(expr.String())
	if issues != nil && len(issues.Errors()) > 0 {
		return nil, errors.New("failed to parse expression")
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, errors.New("failed to parse expression")
	}

	out, _, err := prg.Eval(nil)

	if err != nil {
		return nil, errors.New("failed to evaluate expression")
	}

	return out.Value(), nil
}

var ErrInvalidAssignmentExpression = errors.New("assignment expression is not valid")

func (expr *Expression) ToAssignmentExpression() ([]string, string, error) {

	parts := strings.Split(expr.String(), "=")

	if len(parts) != 2 {
		return nil, "", ErrInvalidAssignmentExpression
	}

	return strings.Split(parts[0], "."), parts[1], nil
}
