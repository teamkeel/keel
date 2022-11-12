package rpc

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

type RpcArgParser struct {
}

func (parser *RpcArgParser) ParseUpdate(operation *proto.Operation, request *http.Request) (*actions.Args, error) {
	data, err := postParamsToInputs(request.Body)
	if err != nil {
		return nil, err
	}

	values, _ := data["values"].(map[string]any)
	wheres, _ := data["where"].(map[string]any)

	// Add explicit inputs to wheres as well because as they can be used in @permission
	explicitInputs := lo.FilterMap(operation.Inputs, func(in *proto.OperationInput, _ int) (string, bool) {
		isExplicit := in.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_EXPLICIT
		_, isArg := values[in.Name]

		return in.Name, (isExplicit && isArg)
	})
	explicitInputArgs := lo.PickByKeys(values, explicitInputs)
	wheres = lo.Assign(wheres, explicitInputArgs)

	values, err = convertArgsMap(operation, values)
	if err != nil {
		return nil, err
	}
	wheres, err = convertArgsMap(operation, wheres)
	if err != nil {
		return nil, err
	}

	return actions.NewArgs(values, wheres), nil
}

func (parser *RpcArgParser) ParseList(operation *proto.Operation, request *http.Request) (*actions.Args, error) {
	var data map[string]any
	var err error

	switch request.Method {
	case http.MethodGet:
		data = queryParamsToInputs(request.URL.Query())
	case http.MethodPost:
		data, err = postParamsToInputs(request.Body)
		if err != nil {
			return nil, err
		}
	}

	wheres, _ := data["where"].(map[string]any)

	// TODO: don't put the pagination args into the "wheres" map
	first, firstPresent := data["first"]
	if firstPresent {
		firstInt, ok := first.(int)
		if !ok {
			wheres["first"] = nil
		} else {
			wheres["first"] = firstInt
		}
	}

	// TODO: don't put the pagination args into the "wheres" map
	after, afterPresent := data["after"]
	if afterPresent {
		afterStr, ok := after.(string)
		if !ok {
			wheres["after"] = nil
		} else {
			wheres["after"] = afterStr
		}
	}

	return actions.NewArgs(map[string]any{}, wheres), nil
}

func (parser *RpcArgParser) ParseDelete(operation *proto.Operation, request *http.Request) (*actions.Args, error) {
	data, err := postParamsToInputs(request.Body)
	if err != nil {
		return nil, err
	}

	values := map[string]any{}
	wheres := data

	wheres, err = convertArgsMap(operation, wheres)
	if err != nil {
		return nil, err
	}

	return actions.NewArgs(values, wheres), nil
}

func convertArgsMap(operation *proto.Operation, values map[string]any) (map[string]any, error) {
	var err error
	for k, v := range values {
		input, found := lo.Find(operation.Inputs, func(in *proto.OperationInput) bool {
			return in.Name == k
		})

		if !found {
			continue
		}

		if operation.Type == proto.OperationType_OPERATION_TYPE_LIST && input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			if input.Type.Type == proto.Type_TYPE_DATE || input.Type.Type == proto.Type_TYPE_DATETIME {
				listOpMap := v.(map[string]any)

				for kListOp, vListOp := range listOpMap {
					listOpMap[kListOp], err = convertDate(vListOp)
					if err != nil {
						return nil, err
					}
				}
				values[k] = listOpMap
			}
		} else {
			if input.Type.Type == proto.Type_TYPE_DATE || input.Type.Type == proto.Type_TYPE_DATETIME {
				values[k], err = convertDate(v)
				if err != nil {
					return nil, err
				}
			}
		}

	}

	return values, nil
}

func convertDate(value any) (time.Time, error) {
	isoDate, ok := value.(string)
	if !ok {
		panic("date must be an ISO string")
	}

	return time.Parse(time.RFC3339, isoDate)
}

func queryParamsToInputs(q url.Values) map[string]any {
	inputs := make(map[string]any)
	for k := range q {
		inputs[k] = q.Get(k)
	}
	return inputs
}

func postParamsToInputs(b io.ReadCloser) (inputs map[string]any, err error) {

	body, err := io.ReadAll(b)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &inputs)
	if err != nil {
		return nil, err
	}

	return inputs, nil
}
