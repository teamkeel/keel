package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/expressions/visitor"
	"github.com/teamkeel/keel/proto"
)

// SelectQueryGen visits the expression and adds select clauses to the provided query builder
func SelectQueryGen(ctx context.Context, query *QueryBuilder, schema *proto.Schema, model *proto.Model, action *proto.Action, inputs map[string]any) visitor.Visitor[*QueryOperand] {
	return &setQueryGen{
		ctx:    ctx,
		query:  query,
		schema: schema,
		model:  model,
		action: action,
		inputs: inputs,
	}
}

var _ visitor.Visitor[*QueryOperand] = new(setQueryGen)

type setQueryGen struct {
	ctx     context.Context
	query   *QueryBuilder
	operand *QueryOperand
	schema  *proto.Schema
	model   *proto.Model
	action  *proto.Action
	inputs  map[string]any
}

func (v *setQueryGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *setQueryGen) EndCondition(parenthesis bool) error {
	return nil
}

func (v *setQueryGen) VisitAnd() error {
	return errors.New("and operator not supported with set")
}

func (v *setQueryGen) VisitOr() error {
	return errors.New("or operator not supported with set")
}

func (v *setQueryGen) VisitOperator(op string) error {
	return errors.New(fmt.Sprintf("%s operator not supported with set", op))
}

func (v *setQueryGen) VisitLiteral(value any) error {
	if value == nil {
		v.operand = Null()
	} else {
		v.operand = Value(value)
	}
	return nil
}

func (v *setQueryGen) VisitVariable(name string) error {
	operand, err := generateOperand(v.ctx, v.schema, v.model, v.action, v.inputs, []string{name})
	if err != nil {
		return err
	}
	v.operand = operand

	return nil
}

func (v *setQueryGen) VisitField(fragments []string) error {
	operand, err := generateOperand(v.ctx, v.schema, v.model, v.action, v.inputs, fragments)
	if err != nil {
		return err
	}
	v.operand = operand

	return nil
}

func (v *setQueryGen) VisitIdentArray(fragments [][]string) error {
	arr := []string{}
	for _, e := range fragments {
		arr = append(arr, e[1])
	}

	v.operand = Value(arr)

	return nil
}

func (v *setQueryGen) ModelName() string {
	return v.query.Model.Name
}

func (v *setQueryGen) Result() (*QueryOperand, error) {
	return v.operand, nil
}
