package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// todo : move
type CallFunction = func(ctx context.Context, actionName string, opType proto.OperationType, body map[string]any) (any, error)

func NewRpcApi(proto *proto.Schema, api *proto.Api) (*rpcApiBuilder, error) {
	m := &rpcApiBuilder{
		proto: proto,
		Get:   make(map[string]actionHandler),
		Post:  make(map[string]actionHandler),
	}

	return m.build(api, proto)
}

type rpcApiBuilder struct {
	proto *proto.Schema

	callCustomFunc CallFunction

	Get  map[string]actionHandler
	Post map[string]actionHandler
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
			scope, err := actions.NewScope(r.Context(), op, schema, nil)
			var builder actions.GetAction

			inputs := queryParamsToInputs(r.URL.Query())
			if err != nil {
				return nil, err
			}

			result, err := builder.
				Initialise(scope).
				ApplyImplicitFilters(inputs).
				ApplyExplicitFilters(inputs).
				IsAuthorised(inputs).
				Execute(inputs)

			return result.Value.Object, err
		}
		mk.Get[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_LIST:
		var builder actions.ListAction

		handler := func(r *http.Request) (interface{}, error) {

			scope, err := actions.NewScope(r.Context(), op, schema, nil)

			inputs := queryParamsToInputs(r.URL.Query())

			if err != nil {
				return nil, err
			}

			where, err := toArgsMap(inputs, "where")

			if err != nil {
				where = map[string]any{}
			}

			result, err := builder.
				Initialise(scope).
				ApplyImplicitFilters(where).
				ApplyExplicitFilters(where).
				IsAuthorised(inputs).
				Execute(inputs)

			if err != nil {
				return nil, err
			}
			return result.Value.Collection, err
		}
		mk.Get[op.Name] = handler

		// Support post requests which take a full gql query object as the body
		handler = func(r *http.Request) (interface{}, error) {
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
			scope, err := actions.NewScope(r.Context(), op, schema, nil)
			if err != nil {
				return nil, err
			}
			where, err := toArgsMap(inputs, "where")

			if err != nil {
				where = map[string]any{}
			}

			result, err := builder.
				Initialise(scope).
				ApplyImplicitFilters(where).
				ApplyExplicitFilters(where).
				IsAuthorised(inputs).
				Execute(inputs)

			if err != nil {
				return nil, err
			}
			return result.Value.Collection, err
		}
		mk.Post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_CREATE:
		handler := func(r *http.Request) (interface{}, error) {
			var builder actions.CreateAction

			scope, err := actions.NewScope(r.Context(), op, schema, nil)
			if err != nil {
				return nil, err
			}
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}

			if err != nil {
				return nil, err
			}

			return builder.
				Initialise(scope).
				CaptureImplicitWriteInputValues(inputs).
				CaptureSetValues(inputs).
				IsAuthorised(inputs).
				Execute(inputs)
		}
		mk.Post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		handler := func(r *http.Request) (interface{}, error) {
			var builder actions.UpdateAction

			scope, err := actions.NewScope(r.Context(), op, schema, nil)
			if err != nil {
				return nil, err
			}
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
			values, err := toArgsMap(inputs, "values")
			if err != nil {
				return nil, err
			}

			wheres, err := toArgsMap(inputs, "where")
			if err != nil {
				return nil, err
			}

			return builder.
				Initialise(scope).
				// first capture any implicit inputs
				CaptureImplicitWriteInputValues(values).
				// then capture explicitly used inputs
				CaptureSetValues(values).
				// then apply unique filters
				ApplyImplicitFilters(wheres).
				ApplyExplicitFilters(wheres).
				IsAuthorised(inputs).
				Execute(inputs)
		}
		mk.Post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_DELETE:
		handler := func(r *http.Request) (interface{}, error) {
			inputs, err := postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
			var builder actions.DeleteAction

			scope, err := actions.NewScope(r.Context(), op, schema, nil)

			if err != nil {
				return nil, err
			}

			return builder.
				Initialise(scope).
				ApplyImplicitFilters(inputs).
				ApplyExplicitFilters(inputs).
				IsAuthorised(inputs).
				Execute(inputs)
		}
		mk.Post[op.Name] = handler
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		break
	default:
		return fmt.Errorf("addRoute() does not yet support this op.Type: %v", op.Type)
	}

	return nil
}
