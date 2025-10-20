package actions

import (
	"context"
	"errors"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/google/cel-go/common/operators"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// IdentHandler is a function that processes an ident and returns an operand.
// It can also modify the query (e.g., add joins, add WHERE clauses).
type IdentHandler func(ctx context.Context, query *QueryBuilder, schema *proto.Schema, ident *parser.ExpressionIdent, operands *arraystack.Stack) error

// baseQueryGen is a common implementation for query generation visitors.
// It handles the standard visitor methods and delegates ident processing to a handler function.
type baseQueryGen struct {
	ctx          context.Context
	query        *QueryBuilder
	schema       *proto.Schema
	entity       proto.Entity
	action       *proto.Action
	inputs       map[string]any
	operators    *arraystack.Stack
	operands     *arraystack.Stack
	identHandler IdentHandler
}

var _ resolve.Visitor[*QueryBuilder] = new(baseQueryGen)

func (v *baseQueryGen) StartTerm(nested bool) error {
	if op, ok := v.operators.Peek(); ok && op == Not {
		_, _ = v.operators.Pop()
		v.query.Not()
	}

	// Only add parenthesis if we're in a nested condition
	if nested {
		v.query.OpenParenthesis()
	}

	return nil
}

func (v *baseQueryGen) EndTerm(nested bool) error {
	if _, ok := v.operators.Peek(); ok && v.operands.Size() == 2 {
		operator, _ := v.operators.Pop()

		r, ok := v.operands.Pop()
		if !ok {
			return errors.New("expected rhs operand")
		}
		l, ok := v.operands.Pop()
		if !ok {
			return errors.New("expected lhs operand")
		}

		lhs := l.(*QueryOperand)
		rhs := r.(*QueryOperand)

		v.query.And()
		err := v.query.Where(lhs, operator.(ActionOperator), rhs)
		if err != nil {
			return err
		}
	} else if _, ok := v.operators.Peek(); !ok {
		l, hasOperand := v.operands.Pop()
		if hasOperand {
			lhs := l.(*QueryOperand)
			v.query.And()
			err := v.query.Where(lhs, Equals, Value(true))
			if err != nil {
				return err
			}
		}
	}

	// Only close parenthesis if we're nested
	if nested {
		v.query.CloseParenthesis()
	}

	return nil
}

func (v *baseQueryGen) StartFunction(name string) error {
	return nil
}

func (v *baseQueryGen) EndFunction() error {
	return nil
}

func (v *baseQueryGen) StartArgument(num int) error {
	return nil
}

func (v *baseQueryGen) EndArgument() error {
	return nil
}

func (v *baseQueryGen) VisitAnd() error {
	v.query.And()
	return nil
}

func (v *baseQueryGen) VisitOr() error {
	v.query.Or()
	return nil
}

func (v *baseQueryGen) VisitNot() error {
	v.operators.Push(Not)
	return nil
}

func (v *baseQueryGen) VisitOperator(op string) error {
	operator, err := toActionOperator(op)
	if err != nil {
		return err
	}

	v.operators.Push(operator)

	return nil
}

func toActionOperator(op string) (ActionOperator, error) {
	switch op {
	case operators.Equals:
		return Equals, nil
	case operators.NotEquals:
		return NotEquals, nil
	case operators.Greater:
		return GreaterThan, nil
	case operators.GreaterEquals:
		return GreaterThanEquals, nil
	case operators.Less:
		return LessThan, nil
	case operators.LessEquals:
		return LessThanEquals, nil
	case operators.In:
		return OneOf, nil
	default:
		return Unknown, nil
	}
}

func (v *baseQueryGen) VisitLiteral(value any) error {
	if value == nil {
		v.operands.Push(Null())
	} else {
		v.operands.Push(Value(value))
	}
	return nil
}

func (v *baseQueryGen) VisitIdent(ident *parser.ExpressionIdent) error {
	return v.identHandler(v.ctx, v.query, v.schema, ident, v.operands)
}

func (v *baseQueryGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	arr := []string{}
	for _, e := range idents {
		arr = append(arr, e.Fragments[1])
	}

	v.operands.Push(Value(arr))

	return nil
}

func (v *baseQueryGen) Result() (*QueryBuilder, error) {
	return v.query, nil
}
