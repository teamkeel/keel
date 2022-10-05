package actions

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"

	"github.com/iancoleman/strcase"
)

func Create(ctx context.Context, operation *proto.Operation, schema *proto.Schema, inputs map[string]any) (map[string]any, error) {
	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}
	model := proto.FindModel(schema.Models, operation.ModelName)
	modelMap, err := initialValueForModel(model, schema)
	if err != nil {
		return nil, err
	}

<<<<<<< Updated upstream
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

=======
>>>>>>> Stashed changes
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
