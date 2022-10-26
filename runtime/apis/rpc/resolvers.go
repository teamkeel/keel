package rpc

import (
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func GetFn(schema *proto.Schema, operation *proto.Operation, argParser RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		args, err := argParser.ParseGet(r.URL.Query())
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

		result, err := scope.Get(args)

		return result.Object, err
	}
}

func ListFn(schema *proto.Schema, operation *proto.Operation, argParser RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		args, err := argParser.ParseList(r.URL.Query())
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

		result, err := scope.List(args)

		return result.Collection, err
	}
}
