package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// TODO: this logic will be exactly the same on RPC since we'll have abstractions like ArgParser.  We should decouple this from graphql and make it reusable

func GetFn(schema *proto.Schema, operation *proto.Operation, argParser actions.ArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseGet(p.Args)
		if err != nil {
			return nil, err
		}

		var builder actions.GetAction
		scope, err := actions.NewScope(p.Context, operation, schema, nil)
		if err != nil {
			return nil, err
		}

		result, err := builder.
			Initialise(scope).
			ApplyImplicitFilters(args.Wheres()).
			ApplyExplicitFilters(args.Wheres()).
			IsAuthorised(args.Wheres()).
			Execute(args.Wheres())

		if result != nil {
			return result.Value.Object, err
		}
		return nil, err
	}
}

func CreateFn(schema *proto.Schema, operation *proto.Operation, argParser actions.ArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseGet(p.Args)
		if err != nil {
			return nil, err
		}

		var builder actions.CreateAction
		scope, err := actions.NewScope(p.Context, operation, schema, nil)
		if err != nil {
			return nil, err
		}

		result, err := builder.
			Initialise(scope).
			CaptureImplicitWriteInputValues(args.Values()).
			CaptureSetValues(args.Values()).
			IsAuthorised(args.Wheres()).
			Execute(args.Wheres())

		if result != nil {
			return result.Value.Object, err
		}
		return nil, err
	}
}

func ListFn(schema *proto.Schema, operation *proto.Operation, argParser actions.ArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseGet(p.Args)
		if err != nil {
			return nil, err
		}

		var builder actions.ListAction
		scope, err := actions.NewScope(p.Context, operation, schema, nil)
		if err != nil {
			return nil, err
		}

		result, err := builder.
			Initialise(scope).
			ApplyImplicitFilters(args.Wheres()).
			ApplyExplicitFilters(args.Wheres()).
			IsAuthorised(args.Wheres()).
			Execute(args.Wheres())

		if err != nil {
			return nil, err
		}

		records := result.Value.Collection

		hasNextPage := result.Value.HasNextPage

		resp, err := connectionResponse(records, hasNextPage)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func DeleteFn(schema *proto.Schema, operation *proto.Operation, argParser actions.ArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseGet(p.Args)
		if err != nil {
			return nil, err
		}

		var builder actions.DeleteAction
		scope, err := actions.NewScope(p.Context, operation, schema, nil)
		if err != nil {
			return nil, err
		}

		result, err := builder.
			Initialise(scope).
			ApplyImplicitFilters(args.Wheres()).
			ApplyExplicitFilters(args.Wheres()).
			IsAuthorised(args.Wheres()).
			Execute(args.Wheres())

		if result != nil {
			return result.Value.Success, err
		}

		return false, err
	}
}

func UpdateFn(schema *proto.Schema, operation *proto.Operation, argParser actions.ArgParser) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		args, err := argParser.ParseGet(p.Args)
		if err != nil {
			return nil, err
		}

		var builder actions.UpdateAction

		scope, err := actions.NewScope(p.Context, operation, schema, nil)
		if err != nil {
			return nil, err
		}

		result, err := builder.
			Initialise(scope).
			CaptureImplicitWriteInputValues(args.Values()).
			CaptureSetValues(args.Values()).
			ApplyImplicitFilters(args.Wheres()).
			ApplyExplicitFilters(args.Wheres()).
			IsAuthorised(args.Wheres()).
			Execute(args.Wheres())

		if result != nil {
			return result.Value.Object, err
		}

		return nil, err
	}
}
