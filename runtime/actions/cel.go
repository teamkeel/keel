package actions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type expressionVisitor[T any] interface {
	startCondition(nested bool) error
	endCondition(nested bool) error
	visitAnd() error
	visitOr() error
	visitLiteral(value any) error
	visitInput(name string) error
	visitField(fragments []string) error
	visitOperator(operator ActionOperator) error
	result() T
	modelName() string
}

func RunCelVisitor[T any](expression string, visitor expressionVisitor[T]) (T, error) {
	resolver := &CelVisitor[T]{
		visitor: visitor,
	}

	return resolver.run(expression)
}

// CelVisitor steps through the CEL AST and calls out to the visitor
type CelVisitor[T any] struct {
	visitor expressionVisitor[T]
}

func (w *CelVisitor[T]) run(expression string) (T, error) {
	var zero T
	env, err := cel.NewEnv()
	if err != nil {
		return zero, fmt.Errorf("program setup err: %s", err)
	}

	ast, issues := env.Parse(expression)
	if issues != nil && len(issues.Errors()) > 0 {
		return zero, errors.New("unexpected ast parsing issues")
	}

	checkedExpr, err := cel.AstToParsedExpr(ast)
	if err != nil {
		return zero, err
	}

	nested := strings.Contains(expression, " && ") || strings.Contains(expression, " || ")

	if err := w.eval(checkedExpr.Expr, nested, false); err != nil {
		return zero, err
	}

	return w.visitor.result(), nil
}

func (w *CelVisitor[T]) eval(expr *exprpb.Expr, nested bool, inBinaryCondition bool) error {
	var err error

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_ConstExpr, *exprpb.Expr_ListExpr, *exprpb.Expr_SelectExpr, *exprpb.Expr_IdentExpr:
		if !inBinaryCondition {
			err := w.visitor.startCondition(false)
			if err != nil {
				return err
			}
		}
	}

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		err := w.callExpr(expr, nested)
		if err != nil {
			return err
		}
	case *exprpb.Expr_ConstExpr:
		err := w.constExpr(expr)
		if err != nil {
			return err
		}
	case *exprpb.Expr_ListExpr:
		err := w.listExpr(expr)
		if err != nil {
			return err
		}
	case *exprpb.Expr_SelectExpr:
		err := w.selectExpr(expr)
		if err != nil {
			return err
		}
	case *exprpb.Expr_IdentExpr:
		err := w.identExpr(expr)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("no support for expr type: %v", expr.ExprKind)
	}

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_ConstExpr, *exprpb.Expr_ListExpr, *exprpb.Expr_SelectExpr, *exprpb.Expr_IdentExpr:
		if !inBinaryCondition {
			err := w.visitor.endCondition(false)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (w *CelVisitor[T]) callExpr(expr *exprpb.Expr, nested bool) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()

	var err error
	switch fun {
	case operators.LogicalNot, operators.Negate:
		err = w.unaryCall(expr)
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

		err = w.binaryCall(expr, nested)
	default:
		return errors.New("function calls not supported yet")
	}

	return err
}

func (w *CelVisitor[T]) binaryCall(expr *exprpb.Expr, nested bool) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()
	lhs := args[0]

	w.visitor.startCondition(nested)

	inBinary := !(fun == operators.LogicalAnd || fun == operators.LogicalOr)

	// add parens if the current operator is lower precedence than the lhs expr operator.
	lhsParen := isComplexOperatorWithRespectTo(fun, lhs)
	rhs := args[1]
	// add parens if the current operator is lower precedence than the rhs expr operator,
	// or the same precedence and the operator is left recursive.
	rhsParen := isComplexOperatorWithRespectTo(fun, rhs)

	if !rhsParen && isLeftRecursive(fun) {
		rhsParen = isSamePrecedence(fun, rhs)
	}
	if err := w.eval(lhs, lhsParen, inBinary); err != nil {
		return err
	}

	var err error
	switch fun {
	case operators.Equals:
		err = w.visitor.visitOperator(Equals)
	case operators.NotEquals:
		err = w.visitor.visitOperator(NotEquals)
	case operators.Greater:
		err = w.visitor.visitOperator(GreaterThan)
	case operators.GreaterEquals:
		err = w.visitor.visitOperator(GreaterThanEquals)
	case operators.Less:
		err = w.visitor.visitOperator(LessThan)
	case operators.LessEquals:
		err = w.visitor.visitOperator(LessThanEquals)
	case operators.In:
		err = w.visitor.visitOperator(OneOf)
	case operators.LogicalOr:
		err = w.visitor.visitOr()
	case operators.LogicalAnd:
		err = w.visitor.visitAnd()
	default:
		return fmt.Errorf("not implemeneted yet: %s", fun)
	}
	if err != nil {
		return err
	}

	if err := w.eval(rhs, rhsParen, inBinary); err != nil {
		return err
	}

	return w.visitor.endCondition(nested)
}

func (w *CelVisitor[T]) unaryCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()

	switch fun {
	case "NOT":
		//con.operators.Push(Not)
	default:
		return fmt.Errorf("not implemented: %s", fun)
	}

	nested := isComplexOperator(args[0])
	return w.eval(args[0], nested, false)
}

func (w *CelVisitor[T]) constExpr(expr *exprpb.Expr) error {
	c := expr.GetConstExpr()

	v, err := toNative(c)
	if err != nil {
		return err
	}

	return w.visitor.visitLiteral(v)
}

func (w *CelVisitor[T]) listExpr(expr *exprpb.Expr) error {
	l := expr.GetListExpr()
	elems := l.GetElements()
	arr := make([]any, len(elems))

	for i, elem := range elems {
		switch elem.ExprKind.(type) {
		case *exprpb.Expr_SelectExpr:
			// Enum values
			s := elem.GetSelectExpr()
			op := s.GetField()
			arr[i] = op
		case *exprpb.Expr_ConstExpr:
			// Literal values
			c := elem.GetConstExpr()
			v, err := toNative(c)
			if err != nil {
				return err
			}
			arr[i] = v
		}
	}

	return w.visitor.visitLiteral(arr)
}

func (w *CelVisitor[T]) identExpr(expr *exprpb.Expr) error {
	ident := expr.GetIdentExpr()

	var err error
	if ident.Name == strcase.ToLowerCamel(w.visitor.modelName()) {
		err = w.visitor.visitField([]string{ident.Name, "id"})
	} else {
		err = w.visitor.visitInput(ident.Name)
	}

	return err
}

func (w *CelVisitor[T]) selectExpr(expr *exprpb.Expr) error {
	sel := expr.GetSelectExpr()

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		nested := isBinaryOrTernaryOperator(sel.GetOperand())
		err := w.eval(sel.GetOperand(), nested, true)
		if err != nil {
			return err
		}
	}

	fragments, err := selectToFragments(expr)
	if err != nil {
		return err
	}

	return w.visitor.visitField(fragments)
}

func selectToFragments(expr *exprpb.Expr) ([]string, error) {
	fragments := []string{}
	e := expr
	for {
		if s, ok := e.ExprKind.(*exprpb.Expr_SelectExpr); ok {
			fragments = append([]string{e.GetSelectExpr().GetField()}, fragments...)
			e = s.SelectExpr.Operand
		} else if _, ok := e.ExprKind.(*exprpb.Expr_IdentExpr); ok {
			fragments = append([]string{e.GetIdentExpr().Name}, fragments...)
			break
		} else {
			return nil, errors.New("unhandled expression type")
		}
	}

	return fragments, nil
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

func toNative(c *exprpb.Constant) (any, error) {
	switch c.ConstantKind.(type) {
	case *exprpb.Constant_BoolValue:
		return c.GetBoolValue(), nil
	case *exprpb.Constant_DoubleValue:
		return c.GetDoubleValue(), nil
	case *exprpb.Constant_Int64Value:
		return c.GetInt64Value(), nil
	case *exprpb.Constant_StringValue:
		return c.GetStringValue(), nil
	case *exprpb.Constant_Uint64Value:
		return c.GetUint64Value(), nil
	case *exprpb.Constant_NullValue:
		return nil, nil
	default:
		return nil, fmt.Errorf("not implemented : %v", c)
	}
}
