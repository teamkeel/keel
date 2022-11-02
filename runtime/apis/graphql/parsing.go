package graphql

import (
	"errors"
	"time"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

type GraphQlArgParser struct{}

func (parser *GraphQlArgParser) ParseGet(operation *proto.Operation, requestInput interface{}) (*actions.Args, error) {
	data, ok := requestInput.(map[string]any)
	if !ok {
		return nil, errors.New("request data not of type map[string]any")
	}

	input, ok := data["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not of type map[string]any")
	}

	values := map[string]any{}
	wheres := input

	wheres = convertArgsMap(operation, wheres)

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseCreate(operation *proto.Operation, requestInput interface{}) (*actions.Args, error) {
	data, ok := requestInput.(map[string]any)
	if !ok {
		return nil, errors.New("request data not of type map[string]any")
	}

	input, ok := data["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not of type map[string]any")
	}

	// Add explicit inputs to wheres because as they can be used in @permission
	explicitInputs := lo.FilterMap(operation.Inputs, func(in *proto.OperationInput, _ int) (string, bool) {
		_, ok := data[in.Name]
		return in.Name, ok
	})
	explicitInputArgs := lo.PickByKeys(data, explicitInputs)

	values := input
	wheres := explicitInputArgs

	values = convertArgsMap(operation, values)
	wheres = convertArgsMap(operation, wheres)

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseUpdate(operation *proto.Operation, requestInput interface{}) (*actions.Args, error) {
	data, ok := requestInput.(map[string]any)
	if !ok {
		return nil, errors.New("request data not of type map[string]any")
	}

	input, ok := data["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not of type map[string]any")
	}

	values, ok := input["values"].(map[string]any)
	if !ok {
		values = map[string]any{}
	}

	wheres, ok := input["where"].(map[string]any)
	if !ok {
		wheres = map[string]any{}
	}

	// Add explicit inputs to wheres as well because as they can be used in @permission
	explicitInputs := lo.FilterMap(operation.Inputs, func(in *proto.OperationInput, _ int) (string, bool) {
		isExplicit := in.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_EXPLICIT
		_, isArg := values[in.Name]

		return in.Name, (isExplicit && isArg)
	})
	explicitInputArgs := lo.PickByKeys(values, explicitInputs)
	wheres = lo.Assign(wheres, explicitInputArgs)

	values = convertArgsMap(operation, values)
	wheres = convertArgsMap(operation, wheres)

	if len(wheres) == 0 {
		return nil, errors.New("wheres cannot be empty")
	}

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseList(operation *proto.Operation, requestInput interface{}) (*actions.Args, error) {
	data, ok := requestInput.(map[string]any)
	if !ok {
		return nil, errors.New("request data not of type map[string]any")
	}

	wheres := map[string]any{}
	values := map[string]any{}

	input, ok := data["input"].(map[string]any)
	if ok {
		wheres, ok = input["where"].(map[string]any)
		if !ok {
			wheres = map[string]any{}
		}
	}

	wheres = convertArgsMap(operation, wheres)

	first, firstPresent := input["first"]

	if firstPresent {
		firstInt, ok := first.(int)
		if !ok {
			wheres["first"] = nil
		} else {
			wheres["first"] = firstInt
		}
	}
	after, afterPresent := input["after"]

	if afterPresent {
		afterStr, ok := after.(string)
		if !ok {
			wheres["after"] = nil
		} else {
			wheres["after"] = afterStr
		}
	}

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseDelete(operation *proto.Operation, requestInput interface{}) (*actions.Args, error) {
	data, ok := requestInput.(map[string]any)
	if !ok {
		return nil, errors.New("request data not of type map[string]any")
	}

	input, ok := data["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not of type map[string]any")
	}

	if len(input) == 0 {
		return nil, errors.New("arguments cannot be empty")
	}

	values := map[string]any{}
	wheres := input

	wheres = convertArgsMap(operation, wheres)

	return actions.NewArgs(values, wheres), nil
}

func convertArgsMap(operation *proto.Operation, values map[string]any) map[string]any {
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
