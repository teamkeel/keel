package rpc

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

func NewRpcApi(p *proto.Schema, api *proto.Api) func(r *http.Request) (interface{}, error) {
	namesOfModelsUsedByAPI := lo.Map(api.ApiModels, func(m *proto.ApiModel, _ int) string {
		return m.ModelName
	})

	modelInstances := proto.FindModels(p.Models, namesOfModelsUsedByAPI)

	return func(r *http.Request) (interface{}, error) {
		trimmedPath := strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/%s/", api.Name))
		trimmedPath = strings.TrimPrefix(trimmedPath, fmt.Sprintf("/%s/", strings.ToLower(api.Name)))

		var operation *proto.Operation

	modelsLoop:
		for _, model := range modelInstances {
			for _, op := range model.Operations {
				if op.Name == trimmedPath {
					operation = op
					break modelsLoop
				}
			}
		}

		if operation == nil {
			return nil, errors.New("not found")
		}

		var handler func(r *http.Request) (interface{}, error)

		switch operation.Type {
		case proto.OperationType_OPERATION_TYPE_GET:
			handler = GetFn(p, operation, &RpcArgParser{})
		case proto.OperationType_OPERATION_TYPE_CREATE:
			handler = CreateFn(p, operation, &RpcArgParser{})
		case proto.OperationType_OPERATION_TYPE_LIST:
			handler = ListFn(p, operation, &RpcArgParser{})
		case proto.OperationType_OPERATION_TYPE_UPDATE:
			handler = UpdateFn(p, operation, &RpcArgParser{})
		case proto.OperationType_OPERATION_TYPE_DELETE:
			handler = DeleteFn(p, operation, &RpcArgParser{})
		case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
			handler = AuthenticateFn(p, operation, &RpcArgParser{})
		}

		return handler(r)
	}
}
