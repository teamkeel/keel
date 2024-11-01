package actions

import (
	"context"
	"strings"

	"github.com/google/cel-go/interpreter"
	"github.com/teamkeel/keel/proto"
)

type OperandResolver struct {
	context context.Context
	schema  *proto.Schema
	model   *proto.Model
	action  *proto.Action
	inputs  map[string]any
}

func (a *OperandResolver) ResolveName(name string) (any, bool) {
	fragments := strings.Split(name, ".")

	fragments, err := normalisedFragments(a.schema, fragments)
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

// Parent returns the parent of the current activation, may be nil.
// If non-nil, the parent will be searched during resolve calls.
func (a *OperandResolver) Parent() interpreter.Activation {
	return nil
}
