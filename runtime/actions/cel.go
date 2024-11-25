package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func (query *QueryBuilder) whereByExpression(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, expression string, inputs map[string]any) error {
	env, err := cel.NewEnv()
	if err != nil {
		return fmt.Errorf("program setup err: %s", err)
	}

	ast, issues := env.Parse(expression)
	if issues != nil && len(issues.Errors()) > 0 {
		return errors.New("unexpected ast parsing issues")
	}

	checkedExpr, err := cel.AstToParsedExpr(ast)
	if err != nil {
		return err
	}

	un := &celSqlGenerator{
		ctx:       ctx,
		query:     query,
		schema:    schema,
		model:     model,
		action:    action,
		inputs:    inputs,
		operators: arraystack.New(),
		operands:  arraystack.New(),
	}

	if strings.Contains(expression, " && ") || strings.Contains(expression, " || ") {
		query.OpenParenthesis()
	}

	if err := un.visit(checkedExpr.Expr); err != nil {
		return err
	}

	if strings.Contains(expression, " && ") || strings.Contains(expression, " || ") {
		query.CloseParenthesis()
	}

	return nil
}

// celSqlGenerator walks through the CEL AST and calls out to our query celSqlGenerator to construct the SQL statement
type celSqlGenerator struct {
	ctx       context.Context
	query     *QueryBuilder
	schema    *proto.Schema
	model     *proto.Model
	action    *proto.Action
	inputs    map[string]any
	operators *arraystack.Stack
	operands  *arraystack.Stack
}

func (con *celSqlGenerator) visit(expr *exprpb.Expr) error {
	var err error

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		err := con.visitCall(expr)
		if err != nil {
			return err
		}

	case *exprpb.Expr_ConstExpr:
		err := con.visitConst(expr)
		if err != nil {
			return err
		}
	case *exprpb.Expr_ListExpr:
		err := con.visitList(expr)
		if err != nil {
			return err
		}

	case *exprpb.Expr_StructExpr:
		panic("no Expr_StructExpr support")
	case *exprpb.Expr_ComprehensionExpr:
		panic("no Expr_ComprehensionExpr support")
	case *exprpb.Expr_SelectExpr:
		err := con.visitSelect(expr)
		if err != nil {
			return err
		}

	case *exprpb.Expr_IdentExpr:
		err := con.visitIdent(expr)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("no support for expr type: %v", expr.ExprKind)

	}

	// This handles single operand conditions, such is post.IsActive
	if operator, ok := con.operators.Peek(); !ok || operator == And || operator == Or {
		l, hasOperand := con.operands.Pop()
		if hasOperand {
			lhs := l.(*QueryOperand)
			err = con.query.Where(lhs, Equals, Value(true))
		}
	}

	return err
}

func (con *celSqlGenerator) visitCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()

	var err error
	switch fun {
	// unary operators
	case operators.LogicalNot, operators.Negate:
		err = con.visitCallUnary(expr)
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

		err = con.visitCallBinary(expr)
	// standard function calls.
	default:
		err = con.visitCallFunc(expr)
	}

	// This handles double operand conditions, such is post.IsActive == false
	if o, ok := con.operators.Peek(); ok && o != And && o != Or {
		operator, _ := con.operators.Pop()

		r, ok := con.operands.Pop()
		if !ok {
			panic("no rhs operand")
		}
		l, ok := con.operands.Pop()
		if !ok {
			panic("no lhs operand")
		}

		lhs := l.(*QueryOperand)
		rhs := r.(*QueryOperand)

		err = con.query.Where(lhs, operator.(ActionOperator), rhs)
		if err != nil {
			return err
		}
	}

	return err
}

func (con *celSqlGenerator) visitCallBinary(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()
	lhs := args[0]

	switch fun {
	case operators.Add:
		con.operators.Push(Addition)
	case operators.Equals:
		con.operators.Push(Equals)
	case operators.NotEquals:
		con.operators.Push(NotEquals)
	case operators.Greater:
		con.operators.Push(GreaterThan)
	case operators.GreaterEquals:
		con.operators.Push(GreaterThanEquals)
	case operators.Less:
		con.operators.Push(LessThan)
	case operators.LessEquals:
		con.operators.Push(LessThanEquals)
	case operators.LogicalOr:
		con.operators.Push(Or)
	case operators.LogicalAnd:
		con.operators.Push(And)
	case operators.In:
		con.operators.Push(OneOf)
	default:
		return fmt.Errorf("not implemeneted yet: %s", fun)
	}

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

	if o, ok := con.operators.Peek(); ok && (o == And || o == Or) {
		operator, _ := con.operators.Pop()
		switch operator {
		case And:
			con.query.And()
		case Or:
			con.query.Or()
		}
	}

	if err := con.visitNested(rhs, rhsParen); err != nil {
		return err
	}

	return nil
}

func (con *celSqlGenerator) visitCallFunc(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	target := c.GetTarget()
	args := c.GetArgs()

	var sqlFun string
	switch fun {
	case "UPPER":
		sqlFun = "UPPER"
	default:
		return fmt.Errorf("not implemented: %s", fun)
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
	}

	con.query.CloseFunction(sqlFun)

	return nil
}

func (con *celSqlGenerator) visitCallUnary(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()

	switch fun {
	case "NOT":
		con.operators.Push(Not)
	default:
		return fmt.Errorf("not implemented: %s", fun)
	}

	nested := isComplexOperator(args[0])
	return con.visitNested(args[0], nested)
}

func (con *celSqlGenerator) visitConst(expr *exprpb.Expr) error {
	resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)
	operand, err := resolver.QueryOperand()
	if err != nil {
		return err
	}
	con.operands.Push(operand)

	return nil
}

func (con *celSqlGenerator) visitList(expr *exprpb.Expr) error {
	resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)

	operand, err := resolver.QueryOperand()
	if err != nil {
		return err
	}

	con.operands.Push(operand)

	return nil
}

func (con *celSqlGenerator) visitIdent(expr *exprpb.Expr) error {
	ident := expr.GetIdentExpr()

	if ident.Name == strcase.ToLowerCamel(con.model.Name) {
		con.operands.Push(IdField())
	} else {
		resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)

		operand, err := resolver.QueryOperand()
		if err != nil {
			return err
		}

		con.operands.Push(operand)
	}
	return nil
}

func (con *celSqlGenerator) visitSelect(expr *exprpb.Expr) error {
	sel := expr.GetSelectExpr()

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		nested := isBinaryOrTernaryOperator(sel.GetOperand())
		err := con.visitNested(sel.GetOperand(), nested)
		if err != nil {
			return err
		}
	}

	resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)

	if resolver.IsModelDbColumn() {
		fragments, _ := resolver.NormalisedFragments()

		err := con.query.AddJoinFromFragments(con.schema, fragments)
		if err != nil {
			return err
		}
	}

	operand, err := resolver.QueryOperand()
	if err != nil {
		return err
	}

	con.operands.Push(operand)

	return nil
}

func (con *celSqlGenerator) visitNested(expr *exprpb.Expr, nested bool) error {
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
