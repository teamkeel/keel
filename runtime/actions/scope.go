package actions

import (
	"context"
	"fmt"

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
	schema *proto.Schema) *Scope {

	model := proto.FindModel(schema.Models, operation.ModelName)

	return &Scope{
		context:   ctx,
		operation: operation,
		model:     model,
		schema:    schema,
	}
}

func Execute(scope *Scope, inputs map[string]any) (any, error) {
	switch scope.operation.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		return Get(scope, inputs)
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		return Update(scope, inputs)
	case proto.OperationType_OPERATION_TYPE_CREATE:
		return Create(scope, inputs)
	case proto.OperationType_OPERATION_TYPE_DELETE:
		return Delete(scope, inputs)
	case proto.OperationType_OPERATION_TYPE_LIST:
		return List(scope, inputs)
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		return Authenticate(scope, inputs)
	default:
		return nil, fmt.Errorf("unhandled operation type %s", scope.operation.Type.String())
	}
}
