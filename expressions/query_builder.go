package expressions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/teamkeel/keel/runtime/actions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func Build(ast *cel.Ast, query *actions.QueryBuilder, input map[string]any) (string, error) {
	checkedExpr, err := cel.AstToCheckedExpr(ast)
	if err != nil {
		return "", err
	}
	un := &builder{
		typeMap: checkedExpr.TypeMap,
		query:   query,
	}
	if err := un.visit(checkedExpr.Expr); err != nil {
		return "", err
	}
	return un.str.String(), nil
}

type builder struct {
	str     strings.Builder
	typeMap map[int64]*exprpb.Type
	query   *actions.QueryBuilder

	operator actions.ActionOperator
	operands []*actions.QueryOperand
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
		//con.operands = append(con.operands, o)

	case *exprpb.Expr_SelectExpr:
		o, err := con.visitSelect(expr)
		if err != nil {
			return err
		}

		con.operands = append(con.operands, o)
	}

	if len(con.operands) == 2 {
		fmt.Printf("WHERE %s %s %s", con.operands[0].String(), "==", con.operands[1].String())
		err = con.query.Where(con.operands[0], con.operator, con.operands[1])

	}

	return err
}

func (con *builder) visitCall(expr *exprpb.Expr) error {

	fmt.Println("visitCall")

	c := expr.GetCallExpr()

	fmt.Println(c.String())

	fun := c.GetFunction()
	switch fun {
	// unary operators
	case operators.LogicalNot, operators.Negate:
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
		return con.visitCallBinary(expr)
	// standard function calls.
	default:
		return con.visitCallFunc(expr)
	}
}

func (con *builder) visitCallBinary(expr *exprpb.Expr) error {
	fmt.Println("visitCallBinary")

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
	lhsType := con.getType(lhs)
	rhsType := con.getType(rhs)

	if !rhsParen && isLeftRecursive(fun) {
		rhsParen = isSamePrecedence(fun, rhs)
	}
	if err := con.visitMaybeNested(lhs, lhsParen); err != nil {
		return err
	}
	var operator string
	if fun == operators.Add && (lhsType.GetPrimitive() == exprpb.Type_STRING && rhsType.GetPrimitive() == exprpb.Type_STRING) {
		operator = "||"
	} else if fun == operators.Add && (rhsType.GetPrimitive() == exprpb.Type_BYTES && lhsType.GetPrimitive() == exprpb.Type_BYTES) {
		operator = "||"
	} else if fun == operators.Equals && (isNullLiteral(rhs) || isBoolLiteral(rhs)) {
		operator = "IS"
	} else if fun == operators.NotEquals && (isNullLiteral(rhs) || isBoolLiteral(rhs)) {
		operator = "IS NOT"
	} else if op, found := standardSQLBinaryOperators[fun]; found {
		operator = op
	} else if op, found := operators.FindReverseBinaryOperator(fun); found {
		operator = op
	} else {
		return fmt.Errorf("cannot unmangle operator: %s", fun)
	}
	con.str.WriteString(" ")
	con.str.WriteString(operator)
	con.str.WriteString(" ")

	//con.query.Where(con.operands[0], actions.Equals, con.operands[1])

	con.operator = actions.Equals

	if err := con.visitMaybeNested(rhs, rhsParen); err != nil {
		return err
	}

	return nil
}

func (con *builder) visitCallFunc(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	target := c.GetTarget()
	args := c.GetArgs()

	sqlFun, ok := standardSQLFunctions[fun]
	if !ok {
		if fun == overloads.Size {
			argType := con.getType(args[0])
			switch {
			case argType.GetPrimitive() == exprpb.Type_STRING:
				sqlFun = "LENGTH"
			case argType.GetPrimitive() == exprpb.Type_BYTES:
				sqlFun = "LENGTH"
			default:
				return fmt.Errorf("unsupported type: %v", argType)
			}
		} else {
			sqlFun = strings.ToUpper(fun)
		}
	}
	con.str.WriteString(sqlFun)
	con.str.WriteString("(")
	if target != nil {
		nested := isBinaryOrTernaryOperator(target)
		err := con.visitMaybeNested(target, nested)
		if err != nil {
			return err
		}
		con.str.WriteString(", ")
	}
	for i, arg := range args {
		err := con.visit(arg)
		if err != nil {
			return err
		}
		if i < len(args)-1 {
			con.str.WriteString(", ")
		}
	}
	con.str.WriteString(")")
	return nil
}

func (con *builder) visitCallUnary(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()
	var operator string
	if op, found := standardSQLUnaryOperators[fun]; found {
		operator = op
	} else if op, found := operators.FindReverse(fun); found {
		operator = op
	} else {
		return fmt.Errorf("cannot unmangle operator: %s", fun)
	}
	con.str.WriteString(operator)
	nested := isComplexOperator(args[0])
	return con.visitMaybeNested(args[0], nested)
}

func (con *builder) visitConst(expr *exprpb.Expr) (*actions.QueryOperand, error) {
	fmt.Println("visitConst")
	c := expr.GetConstExpr()
	switch c.ConstantKind.(type) {
	case *exprpb.Constant_BoolValue:
		if c.GetBoolValue() {
			return actions.Value(true), nil
			con.str.WriteString("TRUE")
		} else {
			return actions.Value(false), nil
			con.str.WriteString("FALSE")
		}
	case *exprpb.Constant_DoubleValue:
		d := strconv.FormatFloat(c.GetDoubleValue(), 'g', -1, 64)
		con.str.WriteString(d)
	case *exprpb.Constant_Int64Value:
		i := strconv.FormatInt(c.GetInt64Value(), 10)
		con.str.WriteString(i)
	case *exprpb.Constant_NullValue:
		con.str.WriteString("NULL")
	case *exprpb.Constant_StringValue:
		con.str.WriteString(strconv.Quote(c.GetStringValue()))
	case *exprpb.Constant_Uint64Value:
		ui := strconv.FormatUint(c.GetUint64Value(), 10)
		con.str.WriteString(ui)
	default:
		return nil, fmt.Errorf("unimplemented : %v", expr)
	}
	return nil, nil
}

func (con *builder) visitIdent(expr *exprpb.Expr) (*actions.QueryOperand, error) {
	fmt.Println("visitIdent " + expr.GetIdentExpr().GetName())

	con.str.WriteString("'")
	con.str.WriteString(expr.GetIdentExpr().GetName())
	con.str.WriteString("'")

	o := actions.Field(expr.GetIdentExpr().GetName())

	return o, nil
}

func (con *builder) visitSelect(expr *exprpb.Expr) (*actions.QueryOperand, error) {
	sel := expr.GetSelectExpr()

	fmt.Println("visitSelect " + sel.GetField())

	// handle the case when the select expression was generated by the has() macro.
	if sel.GetTestOnly() {
		con.str.WriteString("has(")
	}
	nested := !sel.GetTestOnly() && isBinaryOrTernaryOperator(sel.GetOperand())
	err := con.visitMaybeNested(sel.GetOperand(), nested)
	if err != nil {
		return nil, err
	}
	con.str.WriteString(".'")
	con.str.WriteString(sel.GetField())
	con.str.WriteString("'")
	if sel.GetTestOnly() {
		con.str.WriteString(")")
	}

	expre := []string{sel.GetOperand().GetIdentExpr().GetName()}
	o := actions.ExpressionField(expre, sel.GetField())

	return o, nil
}

func (con *builder) visitMaybeNested(expr *exprpb.Expr, nested bool) error {
	if nested {
		con.str.WriteString("(")
	}
	err := con.visit(expr)
	if err != nil {
		return err
	}
	if nested {
		con.str.WriteString(")")
	}
	return nil
}

func (con *builder) getType(node *exprpb.Expr) *exprpb.Type {
	return con.typeMap[node.GetId()]
}
