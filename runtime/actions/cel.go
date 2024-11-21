package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// GenSql will construct a SQL statement for the expression
func (query *QueryBuilder) whereByExpression(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, expression string, inputs map[string]any) error {
	env, err := cel.NewEnv()
	if err != nil {
		return fmt.Errorf("program setup err: %s", err)
	}

	ast, _ := env.Parse(expression)

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
		operators: arraystack.New(),
		operands:  arraystack.New(),
	}

	query.OpenParenthesis()

	if err := un.visit(checkedExpr.Expr); err != nil {
		return err
	}

	query.CloseParenthesis()

	return nil
}

// celSqlGenerator walks through the CEL AST and calls out to our query celSqlGenerator to construct the SQL statement
type celSqlGenerator struct {
	ctx    context.Context
	query  *QueryBuilder
	schema *proto.Schema
	model  *proto.Model
	action *proto.Action

	operators *arraystack.Stack
	operands  *arraystack.Stack
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

		if o, ok := con.operators.Peek(); ok && o != And && o != Or {

			operator, _ := con.operators.Pop()

			r, _ := con.operands.Pop()
			if !ok {
				panic("sd")
			}
			l, ok := con.operands.Pop()
			if !ok {
				panic("sd")
			}

			rhs := r.(*QueryOperand)
			lhs := l.(*QueryOperand)

			fmt.Printf("%swhere(%s %v %s)", indent, lhs.String(), operator.(ActionOperator), rhs.String())
			fmt.Println()

			err = con.query.Where(lhs, operator.(ActionOperator), rhs)
		}

	case *exprpb.Expr_ConstExpr:

		o, err := con.visitConst(expr)
		if err != nil {
			return err
		}
		con.operands.Push(o)

		fmt.Println(fmt.Sprintf(indent+"Const: %v", o.value))

	case *exprpb.Expr_ListExpr:
		fmt.Println(indent + "Expr_ListExpr")
	case *exprpb.Expr_StructExpr:
		fmt.Println(indent + "Expr_StructExpr")
	case *exprpb.Expr_ComprehensionExpr:
		fmt.Println(indent + "Expr_ComprehensionExpr")
	case *exprpb.Expr_SelectExpr:

		o, err := con.visitSelect(expr)
		if err != nil {
			return err
		}
		con.operands.Push(o)

	case *exprpb.Expr_IdentExpr:
		_, err := con.visitIdent(expr)
		if err != nil {
			return err
		}
	}

	indent = strings.TrimSuffix(indent, "   ")

	return err
}

func (con *celSqlGenerator) visitCall(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()

	fun := c.GetFunction()

	fmt.Println(indent + "Call: " + fun)

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

	fmt.Println(indent + "EndCall: " + fun)

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
			fmt.Println(indent + "AND")
			con.query.And()
		case Or:
			fmt.Println(indent + "OR")
			con.query.Or()
		}
	}

	if err := con.visitNested(rhs, rhsParen); err != nil {
		return err
	}

	// if expr.GetCallExpr().Function == "_&&_" {
	// 	fmt.Println(indent + "AND")
	// 	con.query.And()
	// }

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

func (con *celSqlGenerator) visitCallUnary(expr *exprpb.Expr) error {
	c := expr.GetCallExpr()
	fun := c.GetFunction()
	args := c.GetArgs()

	switch fun {
	case "NOT":
		con.operators.Push(Not)
	default:
		return fmt.Errorf("not implemented : %s", fun)
	}

	nested := isComplexOperator(args[0])
	return con.visitNested(args[0], nested)
}

func (con *celSqlGenerator) visitConst(expr *exprpb.Expr) (*QueryOperand, error) {
	c := expr.GetConstExpr()

	switch c.ConstantKind.(type) {
	case *exprpb.Constant_BoolValue:
		return Value(c.GetBoolValue()), nil
	case *exprpb.Constant_DoubleValue:
		return Value(c.GetDoubleValue()), nil
	case *exprpb.Constant_Int64Value:
		return Value(c.GetInt64Value()), nil
	case *exprpb.Constant_NullValue:
		return Null(), nil
	case *exprpb.Constant_StringValue:
		return Value(c.GetStringValue()), nil
	case *exprpb.Constant_Uint64Value:
		return Value(c.GetUint64Value()), nil
	default:
		return nil, fmt.Errorf("not implemented : %v", expr)
	}
}

func (con *celSqlGenerator) visitIdent(expr *exprpb.Expr) (*QueryOperand, error) {
	//TODO: if not a variable (i.e. input), ignore
	//fmt.Println("visitIdent: " + expr.String())
	return nil, nil //Field(expr.GetIdentExpr().GetName()), nil
}

func (con *celSqlGenerator) visitSelect(expr *exprpb.Expr) (*QueryOperand, error) {
	sel := expr.GetSelectExpr()

	resolver := expressions.NewOperandResolverCel(con.ctx, con.schema, con.model, con.action, expr)
	f, err := resolver.Fragments()
	if err != nil {
		return nil, err
	}
	fmt.Println(indent + "Select: " + strings.Join(f, "."))

	nested := isBinaryOrTernaryOperator(sel.GetOperand())
	err = con.visitNested(sel.GetOperand(), nested)
	if err != nil {
		return nil, err
	}

	if resolver.IsModelDbColumn() {
		fragments, _ := resolver.NormalisedFragments()

		err := con.query.addJoinFromFragments(NewModelScope(con.ctx, con.model, con.schema), fragments)
		if err != nil {
			return nil, err
		}
	}

	operand, err := generateQueryOperandFromExpr(resolver)
	if err != nil {
		return nil, err
	}

	return operand, nil
}

// Generates a database QueryOperand, either representing a field, inline query, a value or null.
func generateQueryOperandFromExpr(resolver *expressions.OperandResolverCel) (*QueryOperand, error) { //, args map[string]any
	var queryOperand *QueryOperand

	switch {
	case resolver.IsContextDbColumn():
		// If this is a value from ctx that requires a database read (such as with identity backlinks),
		// then construct an inline query for this operand.  This is necessary because we can't retrieve this value
		// from the current query builder.

		fragments, err := resolver.NormalisedFragments()
		if err != nil {
			return nil, err
		}

		// Remove the ctx fragment
		fragments = fragments[1:]

		identityModel := resolver.Schema.FindModel(strcase.ToCamel(fragments[0]))
		ctxScope := NewModelScope(resolver.Context, identityModel, resolver.Schema)
		query := NewQuery(identityModel)

		identityId := ""
		if auth.IsAuthenticated(resolver.Context) {
			identity, err := auth.GetIdentity(resolver.Context)
			if err != nil {
				return nil, err
			}
			identityId = identity[parser.FieldNameId].(string)
		}

		err = query.addJoinFromFragments(ctxScope, fragments)
		if err != nil {
			return nil, err
		}

		err = query.Where(IdField(), Equals, Value(identityId))
		if err != nil {
			return nil, err
		}

		selectField := ExpressionField(fragments[:len(fragments)-1], fragments[len(fragments)-1])

		// If there are no matches in the subquery then null will be returned, but null
		// will cause IN and NOT IN filtering of this subquery result to always evaluate as false.
		// Therefore we need to filter out null.
		query.And()
		err = query.Where(selectField, NotEquals, Null())
		if err != nil {
			return nil, err
		}

		currModel := identityModel
		for i := 1; i < len(fragments)-1; i++ {
			name := proto.FindField(resolver.Schema.Models, currModel.Name, fragments[i]).Type.ModelName.Value
			currModel = resolver.Schema.FindModel(name)
		}
		currField := proto.FindField(resolver.Schema.Models, currModel.Name, fragments[len(fragments)-1])

		if currField.Type.Repeated {
			query.SelectUnnested(selectField)
		} else {
			query.Select(selectField)
		}

		queryOperand = InlineQuery(query, selectField)

	case resolver.IsModelDbColumn():
		// If this is a model field then generate the appropriate column operand for the database query.

		fragments, err := resolver.NormalisedFragments()
		if err != nil {
			return nil, err
		}

		// Generate QueryOperand from the fragments that make up the expression operand
		queryOperand, err = operandFromFragments(resolver.Schema, fragments)
		if err != nil {
			return nil, err
		}
	default:
		c := resolver.Operand.GetConstExpr()
		switch c.ConstantKind.(type) {
		case *exprpb.Constant_BoolValue:
			return Value(c.GetBoolValue()), nil
		case *exprpb.Constant_DoubleValue:
			return Value(c.GetDoubleValue()), nil
		case *exprpb.Constant_Int64Value:
			return Value(c.GetInt64Value()), nil
		case *exprpb.Constant_NullValue:
			return Null(), nil
		case *exprpb.Constant_StringValue:
			return Value(c.GetStringValue()), nil
		case *exprpb.Constant_Uint64Value:
			return Value(c.GetUint64Value()), nil
		default:
			return nil, fmt.Errorf("not implemented : %v", c.ConstantKind)
		}
	}

	return queryOperand, nil
}

func (con *celSqlGenerator) visitNested(expr *exprpb.Expr, nested bool) error {

	if nested {
		fmt.Println(indent + "(")
		con.query.OpenParenthesis()
	}

	err := con.visit(expr)
	if err != nil {
		return err
	}

	if nested {
		fmt.Println(indent + ")")

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
