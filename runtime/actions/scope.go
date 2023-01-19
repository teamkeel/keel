package actions

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/functions"
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
	if scope.operation.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		resp, err := functions.CallFunction(scope.context, scope.operation.Name, inputs)
		if err != nil {
			return nil, err
		}

		// For now a custom list function just returns a list of records, but the API's
		// all return an objects containing results and pagination info. So we need
		// to "wrap" the results here.
		// TODO: come up with a better implementation for list functions that can support
		// pagination
		if scope.operation.Type == proto.OperationType_OPERATION_TYPE_LIST {
			results, _ := resp.([]any)
			return map[string]any{
				"results":     results,
				"hasNextPage": false,
			}, nil
		}

		return resp, err
	}

	switch scope.operation.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		v, err := Get(scope, inputs)
		// Get() can return nil, but for some reason if we don't explicitly
		// return nil here too the result becomes an empty map, which is rather
		// odd.
		// Simple repo of this: https://play.golang.com/p/MbBzvhrdOm_f
		if v == nil {
			return nil, err
		}
		return v, err
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
