package graphql

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

func GetFn(schema *proto.Schema) func(p graphql.ResolveParams) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		input := p.Args["input"]
		arguments, ok := input.(map[string]any)
		if !ok {
			return nil, errors.New("input not a map")
		}

		var builder actions.GetAction
		scope, err := actions.NewScope(p.Context, op, schema, nil)

		if err != nil {
			return nil, err
		}

		result, err := builder.
			Initialise(scope).
			ApplyImplicitFilters(arguments).
			ApplyExplicitFilters(arguments).
			IsAuthorised(arguments).
			Execute(arguments)

		if result != nil {
			return result.Value.Object, err
		}
		return nil, err
	}
}

func CreateFn(p graphql.ResolveParams) (interface{}, error) {
	input := p.Args["input"]
	arguments, ok := input.(map[string]any)
	if !ok {
		return nil, errors.New("input not a map")
	}

	var builder actions.CreateAction

	scope, err := actions.NewScope(p.Context, op, schema, nil)

	if err != nil {
		return nil, err
	}

	result, err := builder.
		Initialise(scope).
		CaptureImplicitWriteInputValues(arguments). // todo: err?
		CaptureSetValues(arguments).
		IsAuthorised(arguments).
		Execute(arguments)

	if result != nil {
		return result.Value.Object, err
	}
	return nil, err
}

func ListFn(p graphql.ResolveParams) (interface{}, error) {
	input := p.Args["input"].(map[string]any)

	args, err := mk.NormalizeArgs(input)

	// If no inputs have been specified then we need to initialise an empty
	// input map with no where conditions
	if err != nil {
		args = map[string]any{
			"where": map[string]any{},
		}
	}

	var builder actions.ListAction

	scope, err := actions.NewScope(p.Context, op, schema, nil)

	if err != nil {
		return nil, err
	}

	result, err := builder.
		Initialise(scope).
		ApplyImplicitFilters(args).
		ApplyExplicitFilters(args).
		IsAuthorised(args).
		Execute(args)

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

func DeleteFn(p graphql.ResolveParams) (interface{}, error) {
	input := p.Args["input"]
	arguments, ok := input.(map[string]any)

	if !ok {
		return nil, errors.New("input not a map")
	}

	arguments, err = mk.NormalizeArgs(arguments)

	var builder actions.DeleteAction

	scope, err := actions.NewScope(p.Context, op, schema, nil)

	if err != nil {
		return nil, err
	}

	result, err := builder.
		Initialise(scope).
		ApplyImplicitFilters(arguments).
		ApplyExplicitFilters(arguments).
		IsAuthorised(arguments).
		Execute(arguments)

	if result != nil {
		return result.Value.Success, err
	}

	return false, err
}

func UpdateFn(p graphql.ResolveParams) (interface{}, error) {
	input := p.Args["input"]
	arguments, ok := input.(map[string]any)
	if !ok {
		return nil, errors.New("input not a map")
	}

	args := actions.NewArgs(arguments)

	if err != nil {
		return nil, err
	}

	var builder actions.UpdateAction

	scope, err := actions.NewScope(p.Context, op, schema, nil)

	if err != nil {
		return nil, err
	}

	arguments, err = mk.NormalizeArgs(arguments)

	result, err := builder.
		Initialise(scope).
		// first capture any implicit inputs
		CaptureImplicitWriteInputValues().
		// then capture explicitly used inputs
		CaptureSetValues().
		// then apply unique filters
		ApplyImplicitFilters().
		ApplyExplicitFilters().
		IsAuthorised().
		Execute()

	if result != nil {
		return result.Value.Object, err
	}

	return nil, err
}
