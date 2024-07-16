package graphql

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/teamkeel/graphql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
)

func ActionFunc(schema *proto.Schema, action *proto.Action) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		span := trace.SpanFromContext(p.Context)

		scope := actions.NewScope(p.Context, action, schema)
		input := p.Args["input"]

		res, meta, err := actions.Execute(scope, input)
		if err != nil {
			var runtimeErr common.RuntimeError
			if !errors.As(err, &runtimeErr) {
				logrus.Error(err)
				err = common.RuntimeError{
					Code:    common.ErrInternal,
					Message: "error executing request",
				}

			}

			span.SetAttributes(
				attribute.String("error.code", runtimeErr.Code),
				attribute.String("error.message", runtimeErr.Message),
			)

			return nil, err
		}

		rootValue := p.Info.RootValue.(map[string]interface{})
		headersValue := rootValue["headers"].(map[string][]string)

		if meta != nil {
			for k, v := range meta.Headers {
				headersValue[k] = v
			}
		}

		if action.Type == proto.ActionType_ACTION_TYPE_LIST {
			// actions.Execute() returns any but a list action will return a map
			m, _ := res.(map[string]any)
			return connectionResponse(m)
		}

		return res, nil
	}
}
