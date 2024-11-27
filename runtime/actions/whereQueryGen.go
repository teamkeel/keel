package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/teamkeel/keel/proto"
)

func WhereQueryGen(ctx context.Context, query *QueryBuilder, schema *proto.Schema, inputs map[string]any) expressionVisitor[bool] {
	return &whereQueryGen{
		ctx:       ctx,
		query:     query,
		schema:    schema,
		inputs:    inputs,
		operators: arraystack.New(),
		operands:  arraystack.New(),
	}
}

var _ expressionVisitor[bool] = new(whereQueryGen)

type whereQueryGen struct {
	ctx       context.Context
	query     *QueryBuilder
	schema    *proto.Schema
	operators *arraystack.Stack
	operands  *arraystack.Stack
	inputs    map[string]any
}

func (v *whereQueryGen) result() bool {
	return true
}

func (v *whereQueryGen) startCondition(parenthesis bool) error {
	if parenthesis {
		v.query.OpenParenthesis()
	}

	return nil
}

func (v *whereQueryGen) endCondition(parenthesis bool) error {
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

func (v *whereQueryGen) visitAnd() error {
	v.query.And()
	return nil
}

func (v *whereQueryGen) visitOr() error {
	v.query.Or()
	return nil
}

func (v *whereQueryGen) visitOperator(op ActionOperator) error {
	v.operators.Push(op)
	return nil
}

func (v *whereQueryGen) visitLiteral(value any) error {
	if value == nil {
		v.operands.Push(Null())
	} else {
		v.operands.Push(Value(value))
	}
	return nil
}

func (v *whereQueryGen) visitInput(name string) error {
	value, ok := v.inputs[name]
	if !ok {
		return fmt.Errorf("implicit or explicit input '%s' does not exist in arguments", name)
	}

	v.operands.Push(Value(value))
	return nil
}

func (v *whereQueryGen) visitField(fragments []string) error {
	operand, err := generateOperand(v.ctx, v.schema, v.query.Model, fragments)
	if err != nil {
		return err
	}

	err = v.query.AddJoinFromFragments(v.schema, fragments)
	if err != nil {
		return err
	}

	v.operands.Push(operand)

	return nil
}

func (v *whereQueryGen) modelName() string {
	return v.query.Model.Name
}
