package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func GetFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		input := p.Args["input"].(map[string]any)

		return actions.Get(scope, input)
	}
}

func CreateFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		input := p.Args["input"].(map[string]any)

		return actions.Create(scope, input)
	}
}

func ListFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseList(operation, p.Args)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		result, err := scope.List(args)
		if err != nil {
			return nil, err
		}

		resp, err := connectionResponse(result.Results, result.HasNextPage)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

func DeleteFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		input := p.Args["input"].(map[string]any)

		return actions.Delete(scope, input)
	}
}

func UpdateFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseUpdate(operation, p.Args)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(p.Context, operation, schema)
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
