package actions

import (
	"context"
	"errors"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"golang.org/x/exp/maps"

	"github.com/iancoleman/strcase"
)

func Create(ctx context.Context, operation *proto.Operation, schema *proto.Schema, args map[string]any) (map[string]any, error) {

	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}
	model := proto.FindModel(schema.Models, operation.ModelName)
	modelMap, err := initialValueForModel(model, schema)
	if err != nil {
		return nil, err
	}

	// Now overwrite the fields for which Inputs have been given accordingly.
	implicitInputs := lo.Filter(operation.Inputs, func(input *proto.OperationInput, _ int) bool {
		return input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT
	})

	for _, input := range implicitInputs {
		modelFieldName := input.Target[0]

		// If this argument is missing it must be optional.
		v, ok := args[input.Name]
		if !ok {
			continue
		}
		v, err := toMap(v, input.Type.Type)
		if err != nil {
			return nil, err
		}
		modelMap[strcase.ToSnake(modelFieldName)] = v
	}

	setArgs, err := SetExpressionInputsToModelMap(operation, args, schema, ctx)

	if err != nil {
		return nil, err
	}

	// todo: clashing keys between implicit / explicit args (is this possible?)
	maps.Copy(modelMap, setArgs)

	maps.DeleteFunc(modelMap, func(k string, v any) bool {
		match := lo.SomeBy(model.Fields, func(f *proto.Field) bool {
			return strcase.ToSnake(f.Name) == k
		})

		return !match
	})

	authorized, err := EvaluatePermissions(ctx, operation, schema, toLowerCamelMap(modelMap))
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.New("not authorized to access this operation")
	}

	// Write a row to the database.
	if err := db.Table(strcase.ToSnake(model.Name)).Create(modelMap).Error; err != nil {
		return nil, err
	}
	return toLowerCamelMap(modelMap), nil
}

// toLowerCamelMap returns a copy of the given map, in which all
// of the key strings are converted to LowerCamelCase.
// It is good for converting identifiers typically used as database
// table or column names, to the case requirements stipulated by the Keel schema.
func toLowerCamelMap(m map[string]any) map[string]any {
	res := map[string]any{}
	for key, value := range m {
		res[strcase.ToLowerCamel(key)] = value
	}
	return res
}

// toLowerCamelMaps is a convenience wrapper around toLowerCamelMap
// that operates on a list of input maps - rather than just a single map.
func toLowerCamelMaps(maps []map[string]any) []map[string]any {
	res := []map[string]any{}
	for _, m := range maps {
		res = append(res, toLowerCamelMap(m))
	}
	return res
}
