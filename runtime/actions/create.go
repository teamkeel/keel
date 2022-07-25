package actions

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"

	"github.com/iancoleman/strcase"
)

//func Create(ctx context.Context, model *proto.Model, op *proto.Operation, args map[string]any) (map[string]any, error) {
func Create(ctx context.Context, operation *proto.Operation, args map[string]any) (map[string]any, error) {
	schema := runtimectx.GetSchema(ctx)
	model := proto.FindModel(schema.Models, operation.ModelName)
	db := runtimectx.GetDB(ctx)

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

	// for future reference, this is how you would do a GET
	// var myMap map[string]any
	// db.Table(model.Name).Where("id = ?", someID).First(myMap).Error

	return modelMap, nil
}
