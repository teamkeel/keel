package rpc

import (
	"fmt"
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func GetFn(schema *proto.Schema, operation *proto.Operation, argParser *RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			return nil, fmt.Errorf("%s not allowed", r.Method)
		}

		var input map[string]any
		var err error

		switch r.Method {
		case http.MethodGet:
			input = queryParamsToInputs(r.URL.Query())
		case http.MethodPost:
			input, err = postParamsToInputs(r.Body)
			if err != nil {
				return nil, err
			}
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

		return actions.Get(scope, input)
	}
}

func CreateFn(schema *proto.Schema, operation *proto.Operation, argParser *RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		if r.Method != http.MethodPost {
			return nil, fmt.Errorf("%s not allowed", r.Method)
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

		input, err := postParamsToInputs(r.Body)
		if err != nil {
			return nil, err
		}

		return actions.Create(scope, input)
	}
}

func DeleteFn(schema *proto.Schema, operation *proto.Operation, argParser *RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		if r.Method != http.MethodPost {
			return nil, fmt.Errorf("%s not allowed", r.Method)
		}

		args, err := argParser.ParseDelete(operation, r)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

		result, err := scope.Delete(args)
		if err != nil {
			return false, err
		}

		return result, nil
	}
}

func UpdateFn(schema *proto.Schema, operation *proto.Operation, argParser *RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		if r.Method != http.MethodPost {
			return nil, fmt.Errorf("%s not allowed", r.Method)
		}

		args, err := argParser.ParseUpdate(operation, r)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

		result, err := scope.Update(args)

		if err != nil {
			return nil, err
		}

		return result.Object, nil
	}
}

func ListFn(schema *proto.Schema, operation *proto.Operation, argParser *RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			return nil, fmt.Errorf("%s not allowed", r.Method)
		}

		args, err := argParser.ParseList(operation, r)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(r.Context(), operation, schema)
		if err != nil {
			return nil, err
		}

		return scope.List(args)
	}
}

func AuthenticateFn(schema *proto.Schema, operation *proto.Operation, argParser *RpcArgParser) func(r *http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		if r.Method != http.MethodPost {
			return nil, fmt.Errorf("%s not allowed", r.Method)
		}

		data, err := postParamsToInputs(r.Body)
		if err != nil {
			return nil, err
		}

		authArgs := actions.AuthenticateArgs{
			CreateIfNotExists: data["createIfNotExists"].(bool),
			Email:             data["emailPassword"].(map[string]any)["email"].(string),
			Password:          data["emailPassword"].(map[string]any)["password"].(string),
		}

		token, identityCreated, err := actions.Authenticate(r.Context(), schema, &authArgs)
		if err != nil {
			return nil, err
		}

		identityId, err := actions.ParseBearerToken(token)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"identityId":      identityId.String(),
			"identityCreated": identityCreated,
			"token":           token,
		}, nil
	}
}
