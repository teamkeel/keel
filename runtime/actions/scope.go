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

func Execute(scope *Scope, inputs map[string]any) (any, map[string][]string, error) {
	if scope.operation.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		resp, headers, err := functions.CallFunction(scope.context, scope.operation.Name, inputs)
		if err != nil {
			return nil, nil, err
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
			}, headers, nil
		}

		return resp, headers, err
	}

	switch scope.operation.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		v, err := Get(scope, inputs)
		// Get() can return nil, but for some reason if we don't explicitly
		// return nil here too the result becomes an empty map, which is rather
		// odd.
		// Simple repo of this: https://play.golang.com/p/MbBzvhrdOm_f
		if v == nil {
			return nil, nil, err
		}
		return v, nil, err
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		result, err := Update(scope, inputs)
		return result, nil, err
	case proto.OperationType_OPERATION_TYPE_CREATE:
		result, err := Create(scope, inputs)
		return result, nil, err
	case proto.OperationType_OPERATION_TYPE_DELETE:
		result, err := Delete(scope, inputs)
		return result, nil, err
	case proto.OperationType_OPERATION_TYPE_LIST:
		result, err := List(scope, inputs)
		return result, nil, err
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		result, err := Authenticate(scope, inputs)
		return result, nil, err
	default:
		return nil, nil, fmt.Errorf("unhandled operation type %s", scope.operation.Type.String())
	}
}
