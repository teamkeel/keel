package graphql

import (
	"fmt"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func getInput(operation *proto.Operation, args map[string]any) map[string]any {
	input, ok := args["input"].(map[string]any)
	if !ok {
		input = map[string]any{}
	}

	switch operation.Type {
	case proto.OperationType_OPERATION_TYPE_GET, proto.OperationType_OPERATION_TYPE_CREATE, proto.OperationType_OPERATION_TYPE_DELETE:
		input = parseTypes(operation, input)
	case proto.OperationType_OPERATION_TYPE_UPDATE, proto.OperationType_OPERATION_TYPE_LIST:
		if where, ok := input["where"].(map[string]any); ok {
			input["where"] = parseTypes(operation, where)
		}
		if values, ok := input["values"].(map[string]any); ok {
			input["values"] = parseTypes(operation, values)
		}
		return input
	}

	return input
}

func ActionFunc(schema *proto.Schema, operation *proto.Operation) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		input := getInput(operation, p.Args)

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
		case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
			return actions.Authenticate(scope, input)
		default:
			panic(fmt.Errorf("unhandled operation type %s", operation.Type.String()))
		}
	}
}

func parseTypes(operation *proto.Operation, values map[string]any) map[string]any {
	for k, v := range values {
		input, found := lo.Find(operation.Inputs, func(in *proto.OperationInput) bool {
			return in.Name == k
		})

		if !found {
			continue
		}

		if operation.Type == proto.OperationType_OPERATION_TYPE_LIST && input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			if input.Type.Type == proto.Type_TYPE_DATE {
				listOpMap := v.(map[string]any)

				for kListOp, vListOp := range listOpMap {
					listOpMap[kListOp] = convertDate(vListOp)
				}
				values[k] = listOpMap
			}
			if input.Type.Type == proto.Type_TYPE_DATETIME {
				listOpMap := v.(map[string]any)
				for kListOp, vListOp := range listOpMap {
					listOpMap[kListOp] = convertTimestamp(vListOp)
				}
				values[k] = listOpMap
			}
		} else {
			if input.Type.Type == proto.Type_TYPE_DATE {
				values[k] = convertDate(v)
			}
			if input.Type.Type == proto.Type_TYPE_DATETIME {
				values[k] = convertTimestamp(v)
			}
		}

	}

	return values
}

func convertDate(value any) time.Time {
	dateMap, ok := value.(map[string]any)
	if !ok {
		panic("date must be a map")
	}

	day, okDay := dateMap["day"].(int)
	month, okMonth := dateMap["month"].(int)
	year, okYear := dateMap["year"].(int)

	if !(okDay && okMonth && okYear) {
		panic("date badly formatted")
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func convertTimestamp(value any) time.Time {
	timeMap, ok := value.(map[string]any)
	if !ok {
		panic("date must be a map")
	}
	seconds, ok := timeMap["seconds"].(int) // todo: should be int64
	if !ok {
		panic("time badly formatted")
	}

	return time.Unix(int64(seconds), 0).UTC()
}
