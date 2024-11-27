package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

func SetQueryGen(ctx context.Context, query *QueryBuilder, schema *proto.Schema, model *proto.Model, action *proto.Action, inputs map[string]any) expressionVisitor[*QueryOperand] {
	return &setQueryGen{
		ctx:    ctx,
		query:  query,
		schema: schema,
		model:  model,
		action: action,
		inputs: inputs,
	}
}

var _ expressionVisitor[*QueryOperand] = new(setQueryGen)

type setQueryGen struct {
	ctx     context.Context
	query   *QueryBuilder
	operand *QueryOperand
	schema  *proto.Schema
	model   *proto.Model
	action  *proto.Action
	inputs  map[string]any
}

func (v *setQueryGen) startCondition(parenthesis bool) error {
	return nil
}

func (v *setQueryGen) endCondition(parenthesis bool) error {
	return nil
}

func (v *setQueryGen) visitAnd() error {
	return errors.New("and operator not supported with set")
}

func (v *setQueryGen) visitOr() error {
	return errors.New("or operator not supported with set")
}

func (v *setQueryGen) visitOperator(op ActionOperator) error {
	return errors.New(fmt.Sprintf("%s operator not supported with set", op))
}

func (v *setQueryGen) visitLiteral(value any) error {
	if value == nil {
		v.operand = Null()
	} else {
		v.operand = Value(value)
	}
	return nil
}

func (v *setQueryGen) visitInput(name string) error {
	operand, err := generateOperand(v.ctx, v.schema, v.model, v.action, v.inputs, []string{name})
	if err != nil {
		return err
	}
	v.operand = operand

	return nil
}

func (v *setQueryGen) visitField(fragments []string) error {
	operand, err := generateOperand(v.ctx, v.schema, v.model, v.action, v.inputs, fragments)
	if err != nil {
		return err
	}
	v.operand = operand

	return nil
}

func (v *setQueryGen) modelName() string {
	return v.query.Model.Name
}

func (v *setQueryGen) result() *QueryOperand {
	return v.operand
}
