package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
)

type Obj struct {
	Object map[string]any
}

func ParseGetObjectResponse(context context.Context, op *proto.Operation, args WhereArgs) (Row, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)
	if err != nil {
		return nil, err
	}

	return TryParseObjectResponse(res)
}

func ParseCreateObjectResponse(context context.Context, op *proto.Operation, args WhereArgs) (Row, error) {
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

func ParseUpdateResponse(context context.Context, op *proto.Operation, args WhereArgs) (Row, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)

	if err != nil {
		return nil, err
	}

	return TryParseObjectResponse(res)
}

func ParseListResponse(context context.Context, op *proto.Operation, args WhereArgs) (*ActionResult[ListResult], error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)

	if err != nil {
		return nil, err
	}
	resMap, ok := res.(map[string]any)

	if !ok {
		panic("custom function response not a map")
	}

	collection, collectionPresent := resMap["collection"]
	errors, errorsPresent := resMap["errors"]

	if collectionPresent {
		collectionAny, ok := collection.([]any)

		if !ok {
			panic("custom functions object not an array")
		}

		results := []map[string]any{}

		for _, item := range collectionAny {
			item, ok := item.(map[string]any)

			if !ok {
				continue
			}

			results = append(results, item)
		}

		if !ok {
			panic("custom functions object not an array")
		}

		return &ActionResult[ListResult]{
			Value: ListResult{
				Results: results,
			}}, nil
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
	}

	panic("errors in unexpected format")
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
