package actions

import (
	"context"
	"strings"

	"github.com/google/cel-go/interpreter"
	"github.com/teamkeel/keel/proto"
)

// OperandResolver is used to resolve expressions without database access if possible
// i.e. early evaluation.
type OperandResolver struct {
	context context.Context
	schema  *proto.Schema
	model   *proto.Model
	action  *proto.Action
	inputs  map[string]any
}

func (a *OperandResolver) ResolveName(name string) (any, bool) {
	fragments := strings.Split(name, ".")

	fragments, err := NormaliseFragments(a.schema, fragments)
	if err != nil {
		return nil, false
	}

	operand, err := generateOperand(a.context, a.schema, a.model, a.action, a.inputs, fragments)
	if err != nil {
		return nil, false
	}

	if !operand.IsValue() {
		return nil, false
	}

	return operand.value, true
}

func (a *OperandResolver) Parent() interpreter.Activation {
	return nil
}
