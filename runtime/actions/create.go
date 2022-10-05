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
