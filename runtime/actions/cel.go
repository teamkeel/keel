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

// GenSql will construct a SQL statement for the expression
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
		ctx:        ctx,
		query:      query,
		schema:     schema,
		model:      model,
		action:     action,
		inputs:     inputs,
		operators:  arraystack.New(),
		operands:   arraystack.New(),
		printStack: false,
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
	ctx    context.Context
	query  *QueryBuilder
	schema *proto.Schema
	model  *proto.Model
	action *proto.Action
	inputs map[string]any

	operators  *arraystack.Stack
	operands   *arraystack.Stack
	printStack bool
}

var indent string

func (con *celSqlGenerator) visit(expr *exprpb.Expr) error {
	var err error

	indent = indent + "   "

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		err := con.visitCall(expr)
		if err != nil {
			return err
		}

		// This handles double operand conditions, such is post.IsActive == false
		if o, ok := con.operators.Peek(); ok && o != And && o != Or {
			operator, _ := con.operators.Pop()

			r, _ := con.operands.Pop()
			if !ok {
				panic("no rhs operand")
			}
			l, ok := con.operands.Pop()
			if !ok {
				panic("no lhs operand")
			}

			lhs := l.(*QueryOperand)
			rhs := r.(*QueryOperand)

			if con.printStack {
				fmt.Printf("%swhere(%s %v %s)", indent, lhs.String(), operator.(ActionOperator), rhs.String())
				fmt.Println()
			}

			err = con.query.Where(lhs, operator.(ActionOperator), rhs)
			if err != nil {
				return err
			}
		}
	case *exprpb.Expr_ConstExpr:
		o, err := con.visitConst(expr)
		if err != nil {
			return err
		}
		con.operands.Push(o)

		if con.printStack {
			fmt.Println(fmt.Sprintf(indent+"Const: %v", o.value))
		}
	case *exprpb.Expr_ListExpr:
		o, err := con.visitList(expr)
		if err != nil {
			return err
		}
		con.operands.Push(o)

		if con.printStack {
			fmt.Println(fmt.Sprintf(indent+"List: %v", o.value))
		}
	case *exprpb.Expr_StructExpr:
		panic("no Expr_StructExpr support")
	case *exprpb.Expr_ComprehensionExpr:
		panic("no Expr_ComprehensionExpr support")
	case *exprpb.Expr_SelectExpr:
		o, err := con.visitSelect(expr)
		if err != nil {
			return err
		}
		con.operands.Push(o)
	case *exprpb.Expr_IdentExpr:
		o, err := con.visitIdent(expr)
		if err != nil {
			return err
		}
		con.operands.Push(o)
	default:
		return fmt.Errorf("no support for expr type: %v", expr.ExprKind)

	}

	indent = strings.TrimSuffix(indent, "   ")

	// fmt.Print(indent + " ")
	// o, ok := con.operators.Peek()
	// if ok {
	// 	fmt.Printf(" (next operator: %v)", o.(ActionOperator))
	// } else {
	// 	fmt.Printf(" (no next operator)")
	// }
	// op, ok := con.operands.Peek()
	// if ok {
	// 	fmt.Printf(" (next operand: %s)", op.(*QueryOperand).String())
	// } else {
	// 	fmt.Printf(" (no next operand)")
	// }
	// fmt.Println()

	// This handles single operand conditions, such is post.IsActive
	if operator, ok := con.operators.Peek(); !ok || operator == And || operator == Or {
		l, hasOperand := con.operands.Pop()
		if hasOperand {
			lhs := l.(*QueryOperand)
			err = con.query.Where(lhs, Equals, Value(true))

			if con.printStack {
				fmt.Printf("%swhere(%s %v %s)", indent, lhs.String(), Equals, Value(true))
				fmt.Println()
			}
		}
	}

	return err
}

func (con *celSqlGenerator) visitCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()

	if con.printStack {
		fmt.Println(indent + "Call: " + fun)
	}

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

	if con.printStack {
		fmt.Println(indent + "EndCall: " + fun)
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
			if con.printStack {
				fmt.Println(indent + "AND")
			}
			con.query.And()
		case Or:
			if con.printStack {
				fmt.Println(indent + "OR")
			}
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
		// TODO: add next arg
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

func (con *celSqlGenerator) visitConst(expr *exprpb.Expr) (*QueryOperand, error) {
	resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)
	return resolver.QueryOperand()
}

func (con *celSqlGenerator) visitList(expr *exprpb.Expr) (*QueryOperand, error) {

	resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)

	return resolver.QueryOperand()

	// l := expr.GetListExpr()
	// elems := l.GetElements()

	// arr := make([]any, len(elems))

	// for i, elem := range elems {
	// 	switch elem.ExprKind.(type) {
	// 	case *exprpb.Expr_SelectExpr:
	// 		// Enum values
	// 		s := elem.GetSelectExpr()
	// 		op := s.GetField()
	// 		arr[i] = op
	// 	case *exprpb.Expr_ConstExpr:
	// 		// Literal values
	// 		c := elem.GetConstExpr()
	// 		v, err := toNative(c)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		arr[i] = v
	// 	}

	// }

	// return Value(arr), nil
}

func (con *celSqlGenerator) visitIdent(expr *exprpb.Expr) (*QueryOperand, error) {

	ident := expr.GetIdentExpr()
	if con.printStack {
		fmt.Println(fmt.Sprintf(indent+"Ident: %v", ident.Name))
	}

	if ident.Name == strcase.ToLowerCamel(con.model.Name) {
		return IdField(), nil
	}

	resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)

	operand, err := resolver.QueryOperand()
	if err != nil {
		return nil, err
	}

	// if inputs, ok := con.inputs["where"]; ok {
	// 	for k, value := range inputs.(map[string]any) {
	// 		if k == ident.Name {
	// 			return Value(value), nil
	// 		}
	// 	}
	// }

	return operand, nil
}

func (con *celSqlGenerator) visitSelect(expr *exprpb.Expr) (*QueryOperand, error) {
	sel := expr.GetSelectExpr()

	resolver := NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr, con.inputs)
	// f, err := resolver.Fragments()
	// if err != nil {
	// 	return nil, err
	// }

	// if con.printStack {
	// 	fmt.Println(indent + "Select: " + strings.Join(f, "."))
	// }

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		nested := isBinaryOrTernaryOperator(sel.GetOperand())
		err := con.visitNested(sel.GetOperand(), nested)
		if err != nil {
			return nil, err
		}
	}

	if resolver.IsModelDbColumn() {
		fragments, _ := resolver.NormalisedFragments()

		err := con.query.AddJoinFromFragments(con.schema, fragments)
		if err != nil {
			return nil, err
		}
	}

	operand, err := resolver.QueryOperand()
	if err != nil {
		return nil, err
	}

	return operand, nil
}

func (con *celSqlGenerator) visitNested(expr *exprpb.Expr, nested bool) error {

	if nested {
		if con.printStack {
			fmt.Println(indent + "(")
		}
		con.query.OpenParenthesis()
	}

	err := con.visit(expr)
	if err != nil {
		return err
	}

	if nested {
		if con.printStack {
			fmt.Println(indent + ")")
		}

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
