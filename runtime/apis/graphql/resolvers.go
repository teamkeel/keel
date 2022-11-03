package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func GetFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseGet(operation, p.Args)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		result, err := scope.Get(args)
		if err != nil {
			return nil, err
		}

		return result.Object, nil
	}
}

func CreateFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseCreate(operation, p.Args)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(p.Context, operation, schema)
		if err != nil {
			return nil, err
		}

		result, err := scope.Create(args)
		if err != nil {
			return nil, err
		}

		return result.Object, nil
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

		resp, err := connectionResponse(result.Collection, result.HasNextPage)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

func DeleteFn(schema *proto.Schema, operation *proto.Operation, argParser *GraphQlArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseDelete(operation, p.Args)
		if err != nil {
			return nil, err
		}

		scope, err := actions.NewScope(p.Context, operation, schema)
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
