package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/karlseguin/typed"
	"github.com/samber/lo"
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

	return TryParseObjectResponse(res, op)
}

func ParseCreateObjectResponse(context context.Context, op *proto.Operation, args WhereArgs) (map[string]any, error) {
	res, err := functions.CallFunction(context, op.Name, op.Type, args)
	if err != nil {
		return nil, err
	}

	return TryParseObjectResponse(res, op)
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

	return TryParseObjectResponse(res, op)
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
		collection = lo.Map(collection, func(item map[string]interface{}, _ int) map[string]any {
			return transformResponse(item, op)
		})
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
func TryParseObjectResponse(res any, operation *proto.Operation) (map[string]any, error) {
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

		objectMap = transformResponse(objectMap, operation)

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

func transformResponse(response map[string]any, op *proto.Operation) map[string]any {
	for key, value := range response {
		input, found := lo.Find(op.Inputs, func(i *proto.OperationInput) bool {
			return i.Name == key
		})

		if !found {
			continue
		}

		switch input.Type.Type {
		case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			timeStr, ok := value.(string)

			if !ok {
				continue
			}

			t, err := time.Parse(time.RFC3339, timeStr)

			if err != nil {
				continue
			}

			response[key] = t
		default:
			continue
		}
	}

	return response
}
