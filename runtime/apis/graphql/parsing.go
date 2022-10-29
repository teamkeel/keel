package graphql

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/runtime/actions"
)

type GraphQlArgParser struct {
}

// TODO: In here we must inspect the data structures and parse arguments accordingly.  E.g. for date and time to time.Time

func (parser *GraphQlArgParser) ParseGet(requestInput map[string]any) (*actions.Args, error) {
	input, ok := requestInput["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not a map")
	}

	if len(input) == 0 {
		return nil, errors.New("arguments cannot be empty")
	}

	values := map[string]any{}
	wheres := input

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseCreate(requestInput map[string]any) (*actions.Args, error) {
	input, ok := requestInput["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not a map")
	}

	if len(input) == 0 {
		return nil, errors.New("arguments cannot be empty")
	}

	values := input
	wheres := map[string]any{}

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseUpdate(requestInput map[string]any) (*actions.Args, error) {
	input, ok := requestInput["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not a map")
	}

	values, err := toArgsMap(input, "values")
	if err != nil {
		return nil, err
	}

	wheres, err := toArgsMap(input, "where")
	if err != nil {
		return nil, err
	}

	if len(wheres) == 0 {
		return nil, errors.New("wheres cannot be empty")
	}

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseList(requestInput map[string]any) (*actions.Args, error) {
	input, ok := requestInput["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not a map")
	}

	values, err := toArgsMap(input, "values")
	if err != nil {
		return nil, err
	}

	wheres, err := toArgsMap(input, "where")
	if err != nil {
		return nil, err
	}

	return actions.NewArgs(values, wheres), nil
}

func (parser *GraphQlArgParser) ParseDelete(requestInput map[string]any) (*actions.Args, error) {
	input, ok := requestInput["input"].(map[string]any)
	if !ok {
		return nil, errors.New("input not a map")
	}

	if len(input) == 0 {
		return nil, errors.New("arguments cannot be empty")
	}

	values := map[string]any{}
	wheres := input

	return actions.NewArgs(values, wheres), nil
}

func toArgsMap(input map[string]any, key string) (map[string]any, error) {
	subKey, ok := input[key]

	if !ok {
		return nil, fmt.Errorf("%s missing", key)
	}

	subMap, ok := subKey.(map[string]any)

	if !ok {
		return nil, fmt.Errorf("%s does not match expected format", key)
	}

	return subMap, nil
}
