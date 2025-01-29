package resolve

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var (
	ErrExpressionNotParseable = errors.New("expression is invalid and cannot be parsed")
)

type Visitor[T any] interface {
	// StartTerm is called when a new term is visited
	StartTerm(nested bool) error
	// EndTerm is called when a term is finished
	EndTerm(nested bool) error
	// StartFunction is called when a function is started
	StartFunction(name string) error
	// EndFunction is called when a function is finished
	EndFunction() error
	// VisitAnd is called when an 'and' operator is visited between conditions
	VisitAnd() error
	// VisitAnd is called when an 'or' operator is visited between conditions
	VisitOr() error
	// VisitNot is called when a logical not '!' is visited before a condition
	VisitNot() error
	// VisitLiteral is called when a literal operand is visited (e.g. "Keel")
	VisitLiteral(value any) error
	// VisitIdent is called when a field operand, variable or enum value is visited (e.g. post.name)
	VisitIdent(ident *parser.ExpressionIdent) error
	// VisitIdentArray is called when an ident array is visited (e.g. [Category.Sport, Category.Edu])
	VisitIdentArray(idents []*parser.ExpressionIdent) error
	// VisitOperator is called when a condition's operator visited (e.g. ==)
	VisitOperator(operator string) error
	// Returns a value after the visitor has completed executing
	Result() (T, error)
}

func RunCelVisitor[T any](expression *parser.Expression, visitor Visitor[T]) (T, error) {
	resolver := &CelVisitor[T]{
		visitor:    visitor,
		expression: expression,
	}

	return resolver.run(expression)
}

// CelVisitor steps through the CEL AST and calls out to the visitor
type CelVisitor[T any] struct {
	visitor    Visitor[T]
	expression *parser.Expression
	ast        *cel.Ast
}

func (w *CelVisitor[T]) run(expression *parser.Expression) (T, error) {
	var zero T
	env, err := cel.NewEnv()
	if err != nil {
		return zero, fmt.Errorf("cel program setup err: %s", err)
	}

	ast, issues := env.Parse(expression.String())
	if issues != nil && len(issues.Errors()) > 0 {
		return zero, ErrExpressionNotParseable
	}

	checkedExpr, err := cel.AstToParsedExpr(ast)
	if err != nil {
		return zero, err
	}

	w.ast = ast

	if err := w.eval(checkedExpr.Expr, isComplexOperatorWithRespectTo(operators.LogicalAnd, checkedExpr.Expr), false); err != nil {
		return zero, err
	}

	return w.visitor.Result()
}

func (w *CelVisitor[T]) eval(expr *exprpb.Expr, nested bool, inBinary bool) error {
	var err error

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_ConstExpr, *exprpb.Expr_ListExpr, *exprpb.Expr_SelectExpr, *exprpb.Expr_IdentExpr:
		if !inBinary {
			err := w.visitor.StartTerm(false)
			if err != nil {
				return err
			}
		}
	}

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		err = w.visitor.StartTerm(nested)
		if err != nil {
			return err
		}

		err := w.callExpr(expr)
		if err != nil {
			return err
		}

		err = w.visitor.EndTerm(nested)
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
		if !inBinary {
			err := w.visitor.EndTerm(false)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (w *CelVisitor[T]) callExpr(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()

	var err error
	switch fun {
	case operators.LogicalNot:
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

		err = w.binaryCall(expr)
	default:
		err = w.functionCall(expr)
	}

	return err
}

func (w *CelVisitor[T]) functionCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	target := c.GetTarget()
	args := c.GetArgs()

	err := w.visitor.StartFunction(fun)
	if err != nil {
		return err
	}

	if target != nil {
		//nested := isBinaryOrTernaryOperator(target)
		err := w.eval(target, false, false)
		if err != nil {
			return err
		}
		//con.str.WriteString(", ")
	}
	for _, arg := range args {
		err := w.eval(arg, false, false)
		if err != nil {
			return err
		}
		// if i < len(args)-1 {
		// 	con.str.WriteString(", ")
		// }
	}

	err = w.visitor.EndFunction()
	if err != nil {
		return err
	}

	return nil
}

func (w *CelVisitor[T]) binaryCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	op := c.GetFunction()
	args := c.GetArgs()
	lhs := args[0]
	lhsParen := isComplexOperatorWithRespectTo(op, lhs)
	var err error

	inBinary := !(op == operators.LogicalAnd || op == operators.LogicalOr)

	if err := w.eval(lhs, lhsParen, inBinary); err != nil {
		return err
	}

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

	rhs := args[1]
	rhsParen := isComplexOperatorWithRespectTo(op, rhs)
	if !rhsParen && isLeftRecursive(op) {
		rhsParen = isSamePrecedence(op, rhs)
	}

	if err := w.eval(rhs, rhsParen, inBinary); err != nil {
		return err
	}

	return nil
}

func (w *CelVisitor[T]) unaryCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()

	isComplex := isComplexOperator(args[0])

	switch fun {
	case operators.LogicalNot:
		err := w.visitor.VisitNot()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("not implemented: %s", fun)
	}

	if err := w.eval(args[0], isComplex, false); err != nil {
		return err
	}

	return nil
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

	// If it is empty, assume it's a literal array
	if len(elems) == 0 {
		return w.constArray(expr)
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

	arr := make([]*parser.ExpressionIdent, len(elems))
	for i, elem := range elems {
		switch elem.ExprKind.(type) {
		case *exprpb.Expr_IdentExpr:
			s := elem.GetIdentExpr()

			offsets := w.ast.NativeRep().SourceInfo().OffsetRanges()[elem.GetId()]
			start := w.ast.NativeRep().SourceInfo().GetStartLocation(elem.GetId())
			end := w.ast.NativeRep().SourceInfo().GetStopLocation(elem.GetId())

			exprIdent := parser.ExpressionIdent{
				Fragments: []string{s.GetName()},
				Node: node.Node{
					Pos: lexer.Position{
						Filename: w.expression.Pos.Filename,
						Line:     w.expression.Pos.Line + start.Line() - 1,
						Column:   w.expression.Pos.Column + start.Column(),
						Offset:   w.expression.Pos.Offset + int(offsets.Start),
					},
					EndPos: lexer.Position{
						Filename: w.expression.Pos.Filename,
						Line:     w.expression.Pos.Line + end.Line() - 1,
						Column:   w.expression.Pos.Column + end.Column(),
						Offset:   w.expression.Pos.Offset + int(offsets.Stop),
					},
				},
			}
			arr[i] = &exprIdent

		case *exprpb.Expr_SelectExpr:
			var err error
			arr[i], err = w.selectToIdent(elem)
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

	offsets := w.ast.NativeRep().SourceInfo().OffsetRanges()[expr.GetId()]
	start := w.ast.NativeRep().SourceInfo().GetStartLocation(expr.GetId())
	end := w.ast.NativeRep().SourceInfo().GetStopLocation(expr.GetId())

	exprIdent := &parser.ExpressionIdent{
		Fragments: []string{ident.GetName()},
		Node: node.Node{
			Pos: lexer.Position{
				Filename: w.expression.Pos.Filename,
				Line:     w.expression.Pos.Line + start.Line() - 1,
				Column:   w.expression.Pos.Column + start.Column(),
				Offset:   w.expression.Pos.Offset + int(offsets.Start),
			},
			EndPos: lexer.Position{
				Filename: w.expression.Pos.Filename,
				Line:     w.expression.Pos.Line + end.Line() - 1,
				Column:   w.expression.Pos.Column + end.Column(),
				Offset:   w.expression.Pos.Offset + int(offsets.Stop),
			},
		},
	}

	return w.visitor.VisitIdent(exprIdent)
}

func (w *CelVisitor[T]) SelectExpr(expr *exprpb.Expr) error {
	sel := expr.GetSelectExpr()

	switch expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		err := w.eval(sel.GetOperand(), true, true)
		if err != nil {
			return err
		}
	}

	ident, err := w.selectToIdent(expr)
	if err != nil {
		return err
	}

	return w.visitor.VisitIdent(ident)
}

func (w *CelVisitor[T]) selectToIdent(expr *exprpb.Expr) (*parser.ExpressionIdent, error) {
	ident := parser.ExpressionIdent{}
	e := expr

	offset := 0
	for {
		if s, ok := e.ExprKind.(*exprpb.Expr_SelectExpr); ok {
			offsets := w.ast.NativeRep().SourceInfo().OffsetRanges()[s.SelectExpr.Operand.Id]
			start := w.ast.NativeRep().SourceInfo().GetStartLocation(s.SelectExpr.Operand.Id)

			ident.Pos = lexer.Position{
				Filename: w.expression.Pos.Filename,
				Line:     w.expression.Pos.Line,
				Column:   w.expression.Pos.Column + start.Column(),
				Offset:   w.expression.Pos.Offset + int(offsets.Start),
			}

			offset += len(s.SelectExpr.GetField()) + 1

			ident.Fragments = append([]string{s.SelectExpr.GetField()}, ident.Fragments...)
			e = s.SelectExpr.Operand
		} else if _, ok := e.ExprKind.(*exprpb.Expr_IdentExpr); ok {
			offset += len(e.GetIdentExpr().Name)

			ident.Fragments = append([]string{e.GetIdentExpr().Name}, ident.Fragments...)
			break
		} else {
			return nil, fmt.Errorf("no support for expr kind in select: %v", expr.ExprKind)
		}
	}

	ident.EndPos = lexer.Position{
		Filename: w.expression.Pos.Filename,
		Line:     w.expression.Pos.Line,
		Column:   ident.Pos.Column + offset,
		Offset:   ident.Pos.Offset + offset,
	}

	return &ident, nil
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
