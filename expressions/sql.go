package expressions

// import (
// 	"fmt"
// 	"strconv"
// 	"strings"

// 	"github.com/google/cel-go/cel"
// 	"github.com/google/cel-go/common/operators"
// 	"github.com/google/cel-go/common/overloads"
// 	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
// )

// func Convert(ast *cel.Ast) (string, error) {
// 	checkedExpr, err := cel.AstToCheckedExpr(ast)
// 	if err != nil {
// 		return "", err
// 	}
// 	un := &converter{
// 		typeMap: checkedExpr.TypeMap,
// 	}
// 	if err := un.visit(checkedExpr.Expr); err != nil {
// 		return "", err
// 	}
// 	return un.str.String(), nil
// }

// type converter struct {
// 	str     strings.Builder
// 	typeMap map[int64]*exprpb.Type
// }

// func (con *converter) visit(expr *exprpb.Expr) error {
// 	switch expr.ExprKind.(type) {
// 	case *exprpb.Expr_CallExpr:
// 		return con.visitCall(expr)
// 	case *exprpb.Expr_ConstExpr:
// 		return con.visitConst(expr)
// 	case *exprpb.Expr_IdentExpr:
// 		return con.visitIdent(expr)
// 	case *exprpb.Expr_SelectExpr:
// 		return con.visitSelect(expr)
// 	}
// 	return fmt.Errorf("unsupported expr: %v", expr)
// }

// func (con *converter) visitCall(expr *exprpb.Expr) error {

// 	c := expr.GetCallExpr()

// 	fmt.Println(c.String())

// 	fun := c.GetFunction()
// 	switch fun {
// 	// unary operators
// 	case operators.LogicalNot, operators.Negate:
// 		return con.visitCallUnary(expr)
// 	// binary operators
// 	case operators.Add,
// 		operators.Divide,
// 		operators.Equals,
// 		operators.Greater,
// 		operators.GreaterEquals,
// 		operators.In,
// 		operators.Less,
// 		operators.LessEquals,
// 		operators.LogicalAnd,
// 		operators.LogicalOr,
// 		operators.Multiply,
// 		operators.NotEquals,
// 		operators.OldIn,
// 		operators.Subtract:
// 		return con.visitCallBinary(expr)
// 	// standard function calls.
// 	default:
// 		return con.visitCallFunc(expr)
// 	}
// }

// var standardSQLBinaryOperators = map[string]string{
// 	operators.LogicalAnd: "AND",
// 	operators.LogicalOr:  "OR",
// 	operators.Equals:     "=",
// 	operators.In:         "IN",
// }

// func (con *converter) visitCallBinary(expr *exprpb.Expr) error {
// 	c := expr.GetCallExpr()
// 	fun := c.GetFunction()
// 	args := c.GetArgs()
// 	lhs := args[0]
// 	// add parens if the current operator is lower precedence than the lhs expr operator.
// 	lhsParen := isComplexOperatorWithRespectTo(fun, lhs)
// 	rhs := args[1]
// 	// add parens if the current operator is lower precedence than the rhs expr operator,
// 	// or the same precedence and the operator is left recursive.
// 	rhsParen := isComplexOperatorWithRespectTo(fun, rhs)
// 	lhsType := con.getType(lhs)
// 	rhsType := con.getType(rhs)

// 	if !rhsParen && isLeftRecursive(fun) {
// 		rhsParen = isSamePrecedence(fun, rhs)
// 	}
// 	if err := con.visitMaybeNested(lhs, lhsParen); err != nil {
// 		return err
// 	}
// 	var operator string
// 	if fun == operators.Add && (lhsType.GetPrimitive() == exprpb.Type_STRING && rhsType.GetPrimitive() == exprpb.Type_STRING) {
// 		operator = "||"
// 	} else if fun == operators.Add && (rhsType.GetPrimitive() == exprpb.Type_BYTES && lhsType.GetPrimitive() == exprpb.Type_BYTES) {
// 		operator = "||"
// 	} else if fun == operators.Equals && (isNullLiteral(rhs) || isBoolLiteral(rhs)) {
// 		operator = "IS"
// 	} else if fun == operators.NotEquals && (isNullLiteral(rhs) || isBoolLiteral(rhs)) {
// 		operator = "IS NOT"
// 	} else if op, found := standardSQLBinaryOperators[fun]; found {
// 		operator = op
// 	} else if op, found := operators.FindReverseBinaryOperator(fun); found {
// 		operator = op
// 	} else {
// 		return fmt.Errorf("cannot unmangle operator: %s", fun)
// 	}
// 	con.str.WriteString(" ")
// 	con.str.WriteString(operator)
// 	con.str.WriteString(" ")

// 	if err := con.visitMaybeNested(rhs, rhsParen); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (con *converter) visitCallFunc(expr *exprpb.Expr) error {
// 	c := expr.GetCallExpr()
// 	fun := c.GetFunction()
// 	target := c.GetTarget()
// 	args := c.GetArgs()

// 	sqlFun, ok := standardSQLFunctions[fun]
// 	if !ok {
// 		if fun == overloads.Size {
// 			argType := con.getType(args[0])
// 			switch {
// 			case argType.GetPrimitive() == exprpb.Type_STRING:
// 				sqlFun = "LENGTH"
// 			case argType.GetPrimitive() == exprpb.Type_BYTES:
// 				sqlFun = "LENGTH"
// 			default:
// 				return fmt.Errorf("unsupported type: %v", argType)
// 			}
// 		} else {
// 			sqlFun = strings.ToUpper(fun)
// 		}
// 	}
// 	con.str.WriteString(sqlFun)
// 	con.str.WriteString("(")
// 	if target != nil {
// 		nested := isBinaryOrTernaryOperator(target)
// 		err := con.visitMaybeNested(target, nested)
// 		if err != nil {
// 			return err
// 		}
// 		con.str.WriteString(", ")
// 	}
// 	for i, arg := range args {
// 		err := con.visit(arg)
// 		if err != nil {
// 			return err
// 		}
// 		if i < len(args)-1 {
// 			con.str.WriteString(", ")
// 		}
// 	}
// 	con.str.WriteString(")")
// 	return nil
// }

// var standardSQLUnaryOperators = map[string]string{
// 	operators.LogicalNot: "NOT ",
// }

// func (con *converter) visitCallUnary(expr *exprpb.Expr) error {
// 	c := expr.GetCallExpr()
// 	fun := c.GetFunction()
// 	args := c.GetArgs()
// 	var operator string
// 	if op, found := standardSQLUnaryOperators[fun]; found {
// 		operator = op
// 	} else if op, found := operators.FindReverse(fun); found {
// 		operator = op
// 	} else {
// 		return fmt.Errorf("cannot unmangle operator: %s", fun)
// 	}
// 	con.str.WriteString(operator)
// 	nested := isComplexOperator(args[0])
// 	return con.visitMaybeNested(args[0], nested)
// }

// func (con *converter) visitConst(expr *exprpb.Expr) error {
// 	c := expr.GetConstExpr()
// 	switch c.ConstantKind.(type) {
// 	case *exprpb.Constant_BoolValue:
// 		if c.GetBoolValue() {
// 			con.str.WriteString("TRUE")
// 		} else {
// 			con.str.WriteString("FALSE")
// 		}
// 	case *exprpb.Constant_DoubleValue:
// 		d := strconv.FormatFloat(c.GetDoubleValue(), 'g', -1, 64)
// 		con.str.WriteString(d)
// 	case *exprpb.Constant_Int64Value:
// 		i := strconv.FormatInt(c.GetInt64Value(), 10)
// 		con.str.WriteString(i)
// 	case *exprpb.Constant_NullValue:
// 		con.str.WriteString("NULL")
// 	case *exprpb.Constant_StringValue:
// 		con.str.WriteString(strconv.Quote(c.GetStringValue()))
// 	case *exprpb.Constant_Uint64Value:
// 		ui := strconv.FormatUint(c.GetUint64Value(), 10)
// 		con.str.WriteString(ui)
// 	default:
// 		return fmt.Errorf("unimplemented : %v", expr)
// 	}
// 	return nil
// }

// func (con *converter) visitIdent(expr *exprpb.Expr) error {
// 	con.str.WriteString("'")
// 	con.str.WriteString(expr.GetIdentExpr().GetName())
// 	con.str.WriteString("'")
// 	return nil
// }

// func (con *converter) visitSelect(expr *exprpb.Expr) error {
// 	sel := expr.GetSelectExpr()
// 	// handle the case when the select expression was generated by the has() macro.
// 	if sel.GetTestOnly() {
// 		con.str.WriteString("has(")
// 	}
// 	nested := !sel.GetTestOnly() && isBinaryOrTernaryOperator(sel.GetOperand())
// 	err := con.visitMaybeNested(sel.GetOperand(), nested)
// 	if err != nil {
// 		return err
// 	}
// 	con.str.WriteString(".'")
// 	con.str.WriteString(sel.GetField())
// 	con.str.WriteString("'")
// 	if sel.GetTestOnly() {
// 		con.str.WriteString(")")
// 	}
// 	return nil
// }

// func (con *converter) visitMaybeNested(expr *exprpb.Expr, nested bool) error {
// 	if nested {
// 		con.str.WriteString("(")
// 	}
// 	err := con.visit(expr)
// 	if err != nil {
// 		return err
// 	}
// 	if nested {
// 		con.str.WriteString(")")
// 	}
// 	return nil
// }

// func (con *converter) getType(node *exprpb.Expr) *exprpb.Type {
// 	return con.typeMap[node.GetId()]
// }

// // isLeftRecursive indicates whether the parser resolves the call in a left-recursive manner as
// // this can have an effect of how parentheses affect the order of operations in the AST.
// func isLeftRecursive(op string) bool {
// 	return op != operators.LogicalAnd && op != operators.LogicalOr
// }

// // isSamePrecedence indicates whether the precedence of the input operator is the same as the
// // precedence of the (possible) operation represented in the input Expr.
// //
// // If the expr is not a Call, the result is false.
// func isSamePrecedence(op string, expr *exprpb.Expr) bool {
// 	if expr.GetCallExpr() == nil {
// 		return false
// 	}
// 	c := expr.GetCallExpr()
// 	other := c.GetFunction()
// 	return operators.Precedence(op) == operators.Precedence(other)
// }

// // isLowerPrecedence indicates whether the precedence of the input operator is lower precedence
// // than the (possible) operation represented in the input Expr.
// //
// // If the expr is not a Call, the result is false.
// func isLowerPrecedence(op string, expr *exprpb.Expr) bool {
// 	if expr.GetCallExpr() == nil {
// 		return false
// 	}
// 	c := expr.GetCallExpr()
// 	other := c.GetFunction()
// 	return operators.Precedence(op) < operators.Precedence(other)
// }

// // Indicates whether the expr is a complex operator, i.e., a call expression
// // with 2 or more arguments.
// func isComplexOperator(expr *exprpb.Expr) bool {
// 	if expr.GetCallExpr() != nil && len(expr.GetCallExpr().GetArgs()) >= 2 {
// 		return true
// 	}
// 	return false
// }

// // Indicates whether it is a complex operation compared to another.
// // expr is *not* considered complex if it is not a call expression or has
// // less than two arguments, or if it has a higher precedence than op.
// func isComplexOperatorWithRespectTo(op string, expr *exprpb.Expr) bool {
// 	if expr.GetCallExpr() == nil || len(expr.GetCallExpr().GetArgs()) < 2 {
// 		return false
// 	}
// 	return isLowerPrecedence(op, expr)
// }

// // Indicate whether this is a binary or ternary operator.
// func isBinaryOrTernaryOperator(expr *exprpb.Expr) bool {
// 	if expr.GetCallExpr() == nil || len(expr.GetCallExpr().GetArgs()) < 2 {
// 		return false
// 	}
// 	_, isBinaryOp := operators.FindReverseBinaryOperator(expr.GetCallExpr().GetFunction())
// 	return isBinaryOp || isSamePrecedence(operators.Conditional, expr)
// }

// func isNullLiteral(node *exprpb.Expr) bool {
// 	_, isConst := node.ExprKind.(*exprpb.Expr_ConstExpr)
// 	if !isConst {
// 		return false
// 	}
// 	_, isNull := node.GetConstExpr().ConstantKind.(*exprpb.Constant_NullValue)
// 	return isNull
// }

// func isBoolLiteral(node *exprpb.Expr) bool {
// 	_, isConst := node.ExprKind.(*exprpb.Expr_ConstExpr)
// 	if !isConst {
// 		return false
// 	}
// 	_, isBool := node.GetConstExpr().ConstantKind.(*exprpb.Constant_BoolValue)
// 	return isBool
// }

// func isStringLiteral(node *exprpb.Expr) bool {
// 	_, isConst := node.ExprKind.(*exprpb.Expr_ConstExpr)
// 	if !isConst {
// 		return false
// 	}
// 	_, isString := node.GetConstExpr().ConstantKind.(*exprpb.Constant_StringValue)
// 	return isString
// }

// // bytesToOctets converts byte sequences to a string using a three digit octal encoded value
// // per byte.
// func bytesToOctets(byteVal []byte) string {
// 	var b strings.Builder
// 	for _, c := range byteVal {
// 		_, _ = fmt.Fprintf(&b, "\\%03o", c)
// 	}
// 	return b.String()
// }
