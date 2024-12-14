package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/google/cel-go/common/operators"
	"github.com/teamkeel/keel/expressions/visitor"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// FilterQueryGen visits the expression and adds filter conditions to the provided query builder
func FilterQueryGen(ctx context.Context, query *QueryBuilder, schema *proto.Schema, model *proto.Model, action *proto.Action, inputs map[string]any) visitor.Visitor[bool] {
	return &whereQueryGen{
		ctx:       ctx,
		query:     query,
		schema:    schema,
		model:     model,
		action:    action,
		inputs:    inputs,
		operators: arraystack.New(),
		operands:  arraystack.New(),
	}
}

var _ visitor.Visitor[bool] = new(whereQueryGen)

type whereQueryGen struct {
	ctx       context.Context
	query     *QueryBuilder
	schema    *proto.Schema
	model     *proto.Model
	action    *proto.Action
	inputs    map[string]any
	operators *arraystack.Stack
	operands  *arraystack.Stack
}

func (v *whereQueryGen) Result() (bool, error) {
	return true, nil
}

func (v *whereQueryGen) StartCondition(parenthesis bool) error {
	if parenthesis {
		v.query.OpenParenthesis()
	}

	return nil
}

func (v *whereQueryGen) EndCondition(parenthesis bool) error {
	if _, ok := v.operators.Peek(); ok && v.operands.Size() == 2 {
		// This handles duel operand conditions, such is post.IsActive == true
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

		err := v.query.Where(lhs, operator.(ActionOperator), rhs)
		if err != nil {
			return err
		}
	} else if _, ok := v.operators.Peek(); !ok {
		// This handles single operand conditions, such is post.IsActive
		l, hasOperand := v.operands.Pop()
		if hasOperand {
			lhs := l.(*QueryOperand)
			err := v.query.Where(lhs, Equals, Value(true))
			if err != nil {
				return err
			}
		}
	}

	if parenthesis {
		v.query.CloseParenthesis()
	}

	return nil
}

func (v *whereQueryGen) VisitAnd() error {
	v.query.And()
	return nil
}

func (v *whereQueryGen) VisitOr() error {
	v.query.Or()
	return nil
}

func (v *whereQueryGen) VisitOperator(op string) error {
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
		return Unknown, fmt.Errorf("not implemeneted: %s", op)
	}
}

func (v *whereQueryGen) VisitLiteral(value any) error {
	if value == nil {
		v.operands.Push(Null())
	} else {
		v.operands.Push(Value(value))
	}
	return nil
}

func (v *whereQueryGen) VisitIdent(ident *parser.ExpressionIdent) error {
	operand, err := generateOperand(v.ctx, v.schema, v.model, v.action, v.inputs, ident.Fragments)
	if err != nil {
		return err
	}

	err = v.query.AddJoinFromFragments(v.schema, ident.Fragments)
	if err != nil {
		return err
	}

	v.operands.Push(operand)

	return nil
}

func (v *whereQueryGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	arr := []string{}
	for _, e := range idents {
		arr = append(arr, e.Fragments[1])
	}

	v.operands.Push(Value(arr))

	return nil
}
