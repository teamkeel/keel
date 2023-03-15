package graphql

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/sirupsen/logrus"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
)

func ActionFunc(schema *proto.Schema, operation *proto.Operation) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		scope := actions.NewScope(p.Context, operation, schema)
		input := p.Args["input"]

		res, headers, err := actions.Execute(scope, input)
		if err != nil {
			var runtimeErr common.RuntimeError
			if !errors.As(err, &runtimeErr) {
				logrus.Error(err)
				err = common.RuntimeError{
					Code:    common.ErrInternal,
					Message: "error executing request",
				}
			}
			return nil, err
		}

		rootValue := p.Info.RootValue.(map[string]interface{})
		headersValue := rootValue["headers"].(map[string][]string)
		for k, v := range headers {
			headersValue[k] = v
		}

		if operation.Type == proto.OperationType_OPERATION_TYPE_LIST {
			// actions.Execute() returns any but a list action will return a map
			m, _ := res.(map[string]any)
			return connectionResponse(m)
		}

		return res, nil
	}
}
