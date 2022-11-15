package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func getInput(args map[string]any) map[string]any {
	input, ok := args["input"].(map[string]any)
	if !ok {
		input = map[string]any{}
	}

	return input
}

func ActionFunc(schema *proto.Schema, operation *proto.Operation) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		input := getInput(p.Args)

		switch operation.Type {
		case proto.OperationType_OPERATION_TYPE_GET:
			return actions.Get(scope, input)
		case proto.OperationType_OPERATION_TYPE_UPDATE:
			return actions.Update(scope, input)
		case proto.OperationType_OPERATION_TYPE_CREATE:
			return actions.Create(scope, input)
		case proto.OperationType_OPERATION_TYPE_DELETE:
			return actions.Delete(scope, input)
		case proto.OperationType_OPERATION_TYPE_LIST:
			res, err := actions.List(scope, input)
			if err != nil {
				return nil, err
			}
			return connectionResponse(res.Results, res.HasNextPage)
		default:
			panic(fmt.Errorf("unhandled operation type %s", operation.Type.String()))
		}
	}
}
