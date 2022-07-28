package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

func Get(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema,
	args map[string]any) (interface{}, error) {

	model := proto.FindModel(schema.Models, operation.ModelName)
	expectedInput := operation.Inputs[0] // Always exactly one for a Get

	// todo: do we need name case coercion of the name?

	// todo: remind self if should be looking at target, not name, and when so? Or is it already resolved in proto.

	inputValue, ok := args[expectedInput.Name]
	if !ok {
		return nil, fmt.Errorf("missing argument: %s", expectedInput.Name)
	}

	// do we need to unpack the inputValue from the arg?

	// Todo: some argument types need mapping to different database types

	// Todo: should we validate the type of the values?, or let postgres object to them later?

	db, err := runtimectx.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	result := map[string]any{}
	w := fmt.Sprintf("%s = ?", expectedInput.Name)
	tableName := strcase.ToSnake(model.Name)
	if err := db.Table(tableName).Where(w, inputValue).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
