package testing

import (
	"errors"
	"time"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

type IntegrationTestArgParser struct{}

func (parser *IntegrationTestArgParser) ParseList(operation *proto.Operation, requestInput interface{}) (*actions.Args, error) {
	data, ok := requestInput.(map[string]any)
	if !ok {
		return nil, errors.New("request data not of type map[string]any")
	}

	values := map[string]any{}
	wheres := convertArgsMap(operation, data)

	first, firstPresent := data["first"]

	if firstPresent {
		firstInt, ok := first.(int)
		if !ok {
			wheres["first"] = nil
		} else {
			wheres["first"] = firstInt
		}
	}
	after, afterPresent := data["after"]

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
	stringValue, ok := value.(string)
	if !ok {
		panic("date must be a string")
	}

	time, err := time.Parse(time.RFC3339, stringValue)
	if err != nil {
		panic(err.Error())
	}

	return time
}

func convertTimestamp(value any) time.Time {
	stringValue, ok := value.(string)
	if !ok {
		panic("timestamp must be a string")
	}

	time, _ := time.Parse(time.RFC3339, stringValue)

	return time
}
