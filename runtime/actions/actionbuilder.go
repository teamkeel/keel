package actions

import (
	"context"

	"github.com/teamkeel/keel/proto"
)

type Scope struct {
	context   context.Context
	operation *proto.Operation
	model     *proto.Model
	schema    *proto.Schema
}

func NewScope(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema) (*Scope, error) {

	model := proto.FindModel(schema.Models, operation.ModelName)

	return &Scope{
		context:   ctx,
		operation: operation,
		model:     model,
		schema:    schema,
	}, nil
}
