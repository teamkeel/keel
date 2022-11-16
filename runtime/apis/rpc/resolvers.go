package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func ActionFunc(schema *proto.Schema, operation *proto.Operation) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var input map[string]any

		switch r.Method {
		case http.MethodGet:
			switch operation.Type {
			case proto.OperationType_OPERATION_TYPE_GET,
				proto.OperationType_OPERATION_TYPE_LIST:
				input = queryParamsToInputs(r.URL.Query())
			default:
				return nil, fmt.Errorf("%s not allowed", r.Method)
			}
		case http.MethodPost:
			var err error
			input, err = postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("%s not allowed", r.Method)
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

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
			return actions.List(scope, input)
		case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
			return actions.Authenticate(scope, input)
		default:
			panic(fmt.Errorf("unhandled operation type %s", operation.Type.String()))
		}
	}
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
