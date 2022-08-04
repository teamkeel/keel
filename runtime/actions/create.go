package actions

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"

	"github.com/iancoleman/strcase"
)

func Create(ctx context.Context, operation *proto.Operation, schema *proto.Schema, args map[string]any) (map[string]any, error) {
	db, err := runtimectx.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	model := proto.FindModel(schema.Models, operation.ModelName)
	modelMap, err := initialValueForModel(model, schema)
	if err != nil {
		return nil, err
	}

	// Now overwrite the fields for which Inputs have been given accordingly.
	for _, input := range operation.Inputs {
		switch input.Behaviour {
		case proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT:
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
		default:
			return nil, fmt.Errorf("input behaviour %s is not yet supported for Create", input.Behaviour)
		}
	}

	// Write a row to the database.
	if err := db.Table(strcase.ToSnake(model.Name)).Create(modelMap).Error; err != nil {
		return nil, err
	}
	return toLowerCamelMap(modelMap), nil
}

func toLowerCamelMap(m map[string]any) map[string]any {
	res := map[string]any{}
	for key, value := range m {
		res[strcase.ToLowerCamel(key)] = value
	}
	return res
}
