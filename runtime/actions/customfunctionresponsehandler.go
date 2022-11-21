package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/karlseguin/typed"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
)

type Obj struct {
	Object map[string]any
}

func ParseGetObjectResponse(context context.Context, op *proto.Operation, args WhereArgs) (map[string]any, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)
	if err != nil {
		return nil, err
	}

	return TryParseObjectResponse(res)
}

func ParseCreateObjectResponse(context context.Context, op *proto.Operation, args WhereArgs) (map[string]any, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)
	if err != nil {
		return nil, err
	}

	return TryParseObjectResponse(res)
}

func ParseDeleteResponse(context context.Context, op *proto.Operation, args WhereArgs) (bool, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)

	if err != nil {
		return false, err
	}
	resMap, ok := res.(map[string]any)

	if !ok {
		panic("custom function response not a map")
	}

	success, successPresent := resMap["success"]
	errors, errorsPresent := resMap["errors"]

	if successPresent {
		success, ok := success.(bool)

		if !ok {
			return false, fmt.Errorf("invalid response from custom function: success was not a bool")
		}

		return success, nil
	} else if errorsPresent {
		errorArr, ok := errors.([]map[string]any)

		if ok && len(errorArr) > 0 {
			messages := []string{}

			for _, err := range errorArr {
				message, ok := err["message"]

				if !ok {
					continue
				}

				messageStr, ok := message.(string)

				if !ok {
					continue
				}

				messages = append(messages, messageStr)
			}

			return false, fmt.Errorf(strings.Join(messages, ","))
		}
	}

	return false, fmt.Errorf("invalid response from custom function: success was not a bool")
}

func ParseUpdateResponse(context context.Context, op *proto.Operation, args WhereArgs) (map[string]any, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)

	if err != nil {
		return nil, err
	}

	return TryParseObjectResponse(res)
}

func ParseListResponse(context context.Context, op *proto.Operation, args WhereArgs) (*ListResult, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)
	if err != nil {
		return nil, err
	}

	resMap, ok := res.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("custom function response is not a map")
	}

	collection := typed.New(resMap).Maps("collection")
	if collection != nil {
		return &ListResult{
			Results: collection,
		}, nil
	}

	errors := typed.New(resMap).Maps("errors")
	if errors != nil {
		messages := []string{}
		for _, err := range errors {
			messages = append(messages, typed.New(err).String("message"))
		}
		return nil, fmt.Errorf(strings.Join(messages, ","))
	}

	return nil, fmt.Errorf("invalid response from custom function")
}

// Tries to parse object returned from custom functions runtime into correct data type
// Otherwise, tries to format error messages returned from custom functions runtime in a nice way in the error return type
// Otherwise panics
func TryParseObjectResponse(res any) (map[string]any, error) {
	resMap, ok := res.(map[string]any)

	if !ok {
		panic("custom function response not a map")
	}

	object, objectPresent := resMap["object"]
	errors, errorsPresent := resMap["errors"]

	if objectPresent {
		objectMap, ok := object.(map[string]any)

		if !ok {
			panic("custom functions object not a map")
		}

		return objectMap, nil
	} else if errorsPresent {
		errorArr, ok := errors.([]map[string]any)

		if ok && len(errorArr) > 0 {

			messages := []string{}

			for _, err := range errorArr {
				message, ok := err["message"]

				if !ok {
					continue
				}

				messageStr, ok := message.(string)

				if !ok {
					continue
				}

				messages = append(messages, messageStr)
			}

			return nil, fmt.Errorf(strings.Join(messages, ","))

		}

		panic("errors in unexpected format")
	}

	return nil, fmt.Errorf("incorrect data returned from custom function")
}
