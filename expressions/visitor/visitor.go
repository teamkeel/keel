package visitor

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var (
	ErrExpressionNotParseable = errors.New("expression is invalid and cannot be parsed")
)

type Visitor[T any] interface {
	// StartCondition is called when a new condition is visited
	StartCondition(nested bool) error
	// EndCondition is called when a condition is finished
	EndCondition(nested bool) error
	// VisitAnd is called when an 'and' operator is visited between conditions
	VisitAnd() error
	// VisitAnd is called when an 'or' operator is visited between conditions
	VisitOr() error
	// VisitLiteral is called when a literal operand is visited (e.g. "Keel")
	VisitLiteral(value any) error
	// VisitVariable is called when a variable operand is visited (e.g. name)
	VisitVariable(name string) error
	// VisitField is called when a field operand is visited (e.g. post.name)
	VisitField(fragments []string) error
	// VisitIdentArray is called when an ident array is visited (e.g. [Category.Sport, Category.Edu])
	VisitIdentArray(idents [][]string) error
	// VisitOperator is called when a condition's operator visited (e.g. ==)
	VisitOperator(operator string) error
	// Returns a value after the visitor has completed executing
	Result() (T, error)
	ModelName() string
}

func RunCelVisitor[T any](expression string, visitor Visitor[T]) (T, error) {
	expression = strings.ReplaceAll(expression, " and ", " && ")
	expression = strings.ReplaceAll(expression, " or ", " || ")

	resolver := &CelVisitor[T]{
		visitor: visitor,
	}

	return resolver.run(expression)
}

// CelVisitor steps through the CEL AST and calls out to the visitor
type CelVisitor[T any] struct {
	visitor Visitor[T]
}

func (w *CelVisitor[T]) run(expression string) (T, error) {
	var zero T
	env, err := cel.NewEnv()
	if err != nil {
		return zero, fmt.Errorf("cel program setup err: %s", err)
	}

	ast, issues := env.Parse(expression)
	if issues != nil && len(issues.Errors()) > 0 {
		return zero, ErrExpressionNotParseable
	}

	checkedExpr, err := cel.AstToParsedExpr(ast)
	if err != nil {
		return zero, err
	}

	nested := strings.Contains(expression, " && ") || strings.Contains(expression, " || ")

	if err := w.eval(checkedExpr.Expr, nested, false); err != nil {
		return zero, err
	}

	return w.visitor.Result()
}

func (w *CelVisitor[T]) eval(expr *exprpb.Expr, nested bool, inBinaryCondition bool) error {
	var err error

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_ConstExpr, *exprpb.Expr_ListExpr, *exprpb.Expr_SelectExpr, *exprpb.Expr_IdentExpr:
		if !inBinaryCondition {
			err := w.visitor.StartCondition(false)
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
		err := w.SelectExpr(expr)
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
			err := w.visitor.EndCondition(false)
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
	op := c.GetFunction()
	args := c.GetArgs()
	lhs := args[0]

	w.visitor.StartCondition(nested)

	inBinary := !(op == operators.LogicalAnd || op == operators.LogicalOr)

	// add parens if the current operator is lower precedence than the lhs expr operator.
	lhsParen := isComplexOperatorWithRespectTo(op, lhs)
	rhs := args[1]
	// add parens if the current operator is lower precedence than the rhs expr operator,
	// or the same precedence and the operator is left recursive.
	rhsParen := isComplexOperatorWithRespectTo(op, rhs)

	if !rhsParen && isLeftRecursive(op) {
		rhsParen = isSamePrecedence(op, rhs)
	}
	if err := w.eval(lhs, lhsParen, inBinary); err != nil {
		return err
	}

	var err error
	switch op {
	case operators.LogicalOr:
		err = w.visitor.VisitOr()
	case operators.LogicalAnd:
		err = w.visitor.VisitAnd()
	default:
		err = w.visitor.VisitOperator(op)
	}
	if err != nil {
		return err
	}

	if err := w.eval(rhs, rhsParen, inBinary); err != nil {
		return err
	}

	return w.visitor.EndCondition(nested)
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

	return w.visitor.VisitLiteral(v)
}

func (w *CelVisitor[T]) listExpr(expr *exprpb.Expr) error {
	elems := expr.GetListExpr().GetElements()

	if len(elems) == 0 {
		return nil
	}

	switch elems[0].ExprKind.(type) {
	case *exprpb.Expr_IdentExpr:
		return w.identArray(expr)
	case *exprpb.Expr_SelectExpr:
		return w.identArray(expr)
	case *exprpb.Expr_ConstExpr:
		return w.constArray(expr)
	}

	return fmt.Errorf("unexpected expr type: %s", expr.ExprKind)
}

func (w *CelVisitor[T]) constArray(expr *exprpb.Expr) error {
	elems := expr.GetListExpr().GetElements()

	arr := make([]any, len(elems))
	for i, elem := range elems {
		c := elem.GetConstExpr()
		v, err := toNative(c)
		if err != nil {
			return err
		}
		arr[i] = v
	}

	return w.visitor.VisitLiteral(arr)
}

func (w *CelVisitor[T]) identArray(expr *exprpb.Expr) error {
	elems := expr.GetListExpr().GetElements()

	arr := make([][]string, len(elems))
	for i, elem := range elems {
		switch elem.ExprKind.(type) {
		case *exprpb.Expr_IdentExpr:
			s := elem.GetIdentExpr()
			arr[i] = []string{s.GetName()}
		case *exprpb.Expr_SelectExpr:
			var err error
			arr[i], err = SelectToFragments(elem)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("not an ident or select: %v", expr.ExprKind)
		}
	}

	return w.visitor.VisitIdentArray(arr)
}

func (w *CelVisitor[T]) identExpr(expr *exprpb.Expr) error {
	ident := expr.GetIdentExpr()

	if ident.Name == strcase.ToLowerCamel(w.visitor.ModelName()) {
		return w.visitor.VisitField([]string{ident.Name, "id"})
	} else {
		return w.visitor.VisitVariable(ident.Name)
	}
}

func (w *CelVisitor[T]) SelectExpr(expr *exprpb.Expr) error {
	sel := expr.GetSelectExpr()

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		nested := isBinaryOrTernaryOperator(sel.GetOperand())
		err := w.eval(sel.GetOperand(), nested, true)
		if err != nil {
			return err
		}
	}

	fragments, err := SelectToFragments(expr)
	if err != nil {
		return err
	}

	return w.visitor.VisitField(fragments)
}

func SelectToFragments(expr *exprpb.Expr) ([]string, error) {
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
			return nil, fmt.Errorf("no support for expr kind in select: %v", expr.ExprKind)
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
		return nil, fmt.Errorf("const kind not implemented: %v", c)
	}
}
