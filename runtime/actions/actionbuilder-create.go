package actions

import (
	"context"

	"github.com/teamkeel/keel/proto"
)

type CreateAction struct {
	Action
}

func (action *CreateAction) Instantiate(ctx context.Context, schema *proto.Schema, operation *proto.Operation) ActionBuilder {
	// can we call instantiate on Action embedded field?

	action.Scope.values = &DbValues{}

	return action
}

func (c *CreateAction) Execute() (*Result, error) {
	return &Result{}, nil
}
