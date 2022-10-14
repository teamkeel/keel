package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func NewRpcApi(proto *proto.Schema, api *proto.Api) (*rpcApiBuilder, error) {
	m := &rpcApiBuilder{
		proto: proto,
		get:   make(map[string]actionHandler),
		post:  make(map[string]actionHandler),
	}

	return m.build(api, proto)
}

type rpcApiBuilder struct {
	proto *proto.Schema
	get   map[string]actionHandler
	post  map[string]actionHandler
}

type actionHandler func(r *http.Request) (interface{}, error)

func (mk *rpcApiBuilder) build(api *proto.Api, schema *proto.Schema) (*rpcApiBuilder, error) {

	namesOfModelsUsedByAPI := lo.Map(api.ApiModels, func(m *proto.ApiModel, _ int) string {
		return m.ModelName
	})

	modelInstances := proto.FindModels(mk.proto.Models, namesOfModelsUsedByAPI)

	for _, model := range modelInstances {
		for _, op := range model.Operations {
			err := mk.addRoute(op, schema)
			if err != nil {
				return nil, err
			}
		}
	}

	return mk, nil
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

type AuthResponse struct {
	Token   string
	Created bool
}

func (mk *rpcApiBuilder) addRoute(
	op *proto.Operation,
	schema *proto.Schema) error {

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		handler := func(r *http.Request) (interface{}, error) {
			inputs := queryParamsToInputs(r.URL.Query())
			return actions.Get(r.Context(), op, schema, inputs)
		}
		mk.get[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_LIST:
		handler := func(r *http.Request) (interface{}, error) {
			inputs := queryParamsToInputs(r.URL.Query())
			res, _, err := actions.List(r.Context(), op, schema, inputs)
			return res, err
		}
		mk.get[op.Name] = handler

		// Support post requests which take a full gql query object as the body
		handler = func(r *http.Request) (interface{}, error) {
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
			res, _, err := actions.List(r.Context(), op, schema, inputs)
			return res, err
		}
		mk.post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_CREATE:
		handler := func(r *http.Request) (interface{}, error) {
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
			return actions.Create(r.Context(), op, schema, inputs)
		}
		mk.post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		handler := func(r *http.Request) (interface{}, error) {
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
			return actions.Update(r.Context(), op, schema, inputs)
		}
		mk.post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_DELETE:
		handler := func(r *http.Request) (interface{}, error) {
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
			return actions.Delete(r.Context(), op, schema, inputs)
		}
		mk.post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		break
	default:
		return fmt.Errorf("addRoute() does not yet support this op.Type: %v", op.Type)
	}

	return nil
}
