package expressions_old

import (
	"errors"
	"fmt"

	"github.com/google/cel-go/common/operators"
	"github.com/teamkeel/keel/runtime/actions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// builder walks through the CEL AST and calls out to our query builder to construct the SQL statement
type builder struct {
	query *actions.QueryBuilder

	operator actions.ActionOperator
	operands []*actions.QueryOperand
	callType string // either "unary", "binary" or "func"
}

func (con *builder) visit(expr *exprpb.Expr) error {
	var err error

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		return con.visitCall(expr)
	case *exprpb.Expr_ConstExpr:
		o, err := con.visitConst(expr)
		if err != nil {
			return err
		}
		con.operands = append(con.operands, o)
	case *exprpb.Expr_IdentExpr:
		_, err := con.visitIdent(expr)
		if err != nil {
			return err
		}
	case *exprpb.Expr_SelectExpr:
		o, err := con.visitSelect(expr)
		if err != nil {
			return err
		}
		con.operands = append(con.operands, o)
	}

	if con.callType == "unary" && len(con.operands) == 1 {
		err = con.query.Where(con.operands[0], actions.NotEquals, actions.Null())
		con.operands = []*actions.QueryOperand{}
	}

	if con.callType == "binary" && len(con.operands) == 2 {
		err = con.query.Where(con.operands[0], con.operator, con.operands[1])
		con.operands = []*actions.QueryOperand{}
	}

	if con.callType == "function" {
		return errors.New("functions not supported yet")
	}

	return err
}

func (con *builder) visitCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()

	fun := c.GetFunction()
	switch fun {
	// unary operators
	case operators.LogicalNot, operators.Negate:
		con.callType = "unary"
		return con.visitCallUnary(expr)
	// binary operators
	case operators.Add,
		operators.Divide,
		operators.Equals,
		operators.Greater,
		operators.GreaterEquals,
		operators.In,
		operators.Less,
		operators.LessEquals,
		operators.LogicalAnd,
		operators.LogicalOr,
		operators.Multiply,
		operators.NotEquals,
		operators.OldIn,
		operators.Subtract:
		con.callType = "binary"
		return con.visitCallBinary(expr)
	// standard function calls.
	default:
		con.callType = "func"
		return con.visitCallFunc(expr)
	}
}

func (con *builder) visitCallBinary(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()
	lhs := args[0]

	// add parens if the current operator is lower precedence than the lhs expr operator.
	lhsParen := isComplexOperatorWithRespectTo(fun, lhs)
	rhs := args[1]
	// add parens if the current operator is lower precedence than the rhs expr operator,
	// or the same precedence and the operator is left recursive.
	rhsParen := isComplexOperatorWithRespectTo(fun, rhs)

	if !rhsParen && isLeftRecursive(fun) {
		rhsParen = isSamePrecedence(fun, rhs)
	}
	if err := con.visitNested(lhs, lhsParen); err != nil {
		return err
	}

	switch fun {
	case operators.Add:
		con.operator = actions.Addition
	case operators.Equals:
		con.operator = actions.Equals
	case operators.NotEquals:
		con.operator = actions.NotEquals
	case operators.Greater:
		con.operator = actions.GreaterThan
	case operators.GreaterEquals:
		con.operator = actions.GreaterThanEquals
	case operators.Less:
		con.operator = actions.LessThan
	case operators.LessEquals:
		con.operator = actions.LessThanEquals
	case operators.LogicalOr:
		con.query.Or()
	case operators.LogicalAnd:
		con.query.And()
	default:
		return fmt.Errorf("not implemeneted yet: %s", fun)
	}

	if err := con.visitNested(rhs, rhsParen); err != nil {
		return err
	}

	return nil
}

func (con *builder) visitCallFunc(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	target := c.GetTarget()
	args := c.GetArgs()

	var sqlFun string
	switch fun {
	case "UPPER":
		sqlFun = "UPPER"
	default:
		return fmt.Errorf("not implemeneted yet: %s", fun)
	}

	con.query.OpenFunction(sqlFun)

	if target != nil {
		nested := isBinaryOrTernaryOperator(target)
		err := con.visitNested(target, nested)
		if err != nil {
			return err
		}
	}
	for _, arg := range args {
		err := con.visit(arg)
		if err != nil {
			return err
		}
		// TODO: add next arg
	}

	con.query.CloseFunction(sqlFun)

	return nil
}

func (con *builder) visitCallUnary(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()

	switch fun {
	case "NOT":
		con.operator = actions.Not
	default:
		return fmt.Errorf("not implemented : %s", fun)
	}

	nested := isComplexOperator(args[0])
	return con.visitNested(args[0], nested)
}

func (con *builder) visitConst(expr *exprpb.Expr) (*actions.QueryOperand, error) {
	c := expr.GetConstExpr()
	switch c.ConstantKind.(type) {
	case *exprpb.Constant_BoolValue:
		return actions.Value(c.GetBoolValue()), nil
	case *exprpb.Constant_DoubleValue:
		return actions.Value(c.GetDoubleValue()), nil
	case *exprpb.Constant_Int64Value:
		return actions.Value(c.GetInt64Value()), nil
	case *exprpb.Constant_NullValue:
		return actions.Null(), nil
	case *exprpb.Constant_StringValue:
		return actions.Value(c.GetStringValue()), nil
	case *exprpb.Constant_Uint64Value:
		return actions.Value(c.GetUint64Value()), nil
	default:
		return nil, fmt.Errorf("not implemented : %v", expr)
	}
}

func (con *builder) visitIdent(expr *exprpb.Expr) (*actions.QueryOperand, error) {
	return actions.Field(expr.GetIdentExpr().GetName()), nil
}

func (con *builder) visitSelect(expr *exprpb.Expr) (*actions.QueryOperand, error) {
	sel := expr.GetSelectExpr()

	nested := isBinaryOrTernaryOperator(sel.GetOperand())
	err := con.visitNested(sel.GetOperand(), nested)
	if err != nil {
		return nil, err
	}

	expre := []string{sel.GetOperand().GetIdentExpr().GetName()}
	o := actions.ExpressionField(expre, sel.GetField())

	return o, nil
}

func (con *builder) visitNested(expr *exprpb.Expr, nested bool) error {
	if nested {
		con.query.OpenParenthesis()
	}
	err := con.visit(expr)
	if err != nil {
		return err
	}
	if nested {
		con.query.CloseParenthesis()
	}
	return nil
}

// isLeftRecursive indicates whether the parser resolves the call in a left-recursive manner as
// this can have an effect of how parentheses affect the order of operations in the AST.
func isLeftRecursive(op string) bool {
	return op != operators.LogicalAnd && op != operators.LogicalOr
}

// isSamePrecedence indicates whether the precedence of the input operator is the same as the
// precedence of the (possible) operation represented in the input Expr.
//
// If the expr is not a Call, the result is false.
func isSamePrecedence(op string, expr *exprpb.Expr) bool {
	if expr.GetCallExpr() == nil {
		return false
	}
	c := expr.GetCallExpr()
	other := c.GetFunction()
	return operators.Precedence(op) == operators.Precedence(other)
}

// isLowerPrecedence indicates whether the precedence of the input operator is lower precedence
// than the (possible) operation represented in the input Expr.
//
// If the expr is not a Call, the result is false.
func isLowerPrecedence(op string, expr *exprpb.Expr) bool {
	if expr.GetCallExpr() == nil {
		return false
	}
	c := expr.GetCallExpr()
	other := c.GetFunction()
	return operators.Precedence(op) < operators.Precedence(other)
}

// Indicates whether the expr is a complex operator, i.e., a call expression
// with 2 or more arguments.
func isComplexOperator(expr *exprpb.Expr) bool {
	if expr.GetCallExpr() != nil && len(expr.GetCallExpr().GetArgs()) >= 2 {
		return true
	}
	return false
}

// Indicates whether it is a complex operation compared to another.
// expr is *not* considered complex if it is not a call expression or has
// less than two arguments, or if it has a higher precedence than op.
func isComplexOperatorWithRespectTo(op string, expr *exprpb.Expr) bool {
	if expr.GetCallExpr() == nil || len(expr.GetCallExpr().GetArgs()) < 2 {
		return false
	}
	return isLowerPrecedence(op, expr)
}

// Indicate whether this is a binary or ternary operator.
func isBinaryOrTernaryOperator(expr *exprpb.Expr) bool {
	if expr.GetCallExpr() == nil || len(expr.GetCallExpr().GetArgs()) < 2 {
		return false
	}
	_, isBinaryOp := operators.FindReverseBinaryOperator(expr.GetCallExpr().GetFunction())
	return isBinaryOp || isSamePrecedence(operators.Conditional, expr)
}
