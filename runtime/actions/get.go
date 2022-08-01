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

	// If there is a where clause, there can be no inputs, but we are not
	// dealing with that case.
	expectedInput := operation.Inputs[0]

	// todo - where clause

	// todo: do we need name case coercion of the name?

	// todo: remind self if should be looking at target, not name, and when so? Or is it already resolved in proto.

	expectedInputIdentifier := expectedInput.Target[0]
	inputValue, ok := args[expectedInputIdentifier]
	if !ok {
		return nil, fmt.Errorf("missing argument: %s", expectedInputIdentifier)
	}

	// do we need to unpack the inputValue from the arg?

	// Todo: some argument types need mapping to different database types

	// Todo: should we validate the type of the values?, or let postgres object to them later?

	db, err := runtimectx.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	result := map[string]any{}
	tableName := strcase.ToSnake(model.Name)
	columnName := strcase.ToSnake(expectedInputIdentifier)
	w := fmt.Sprintf("%s = ?", columnName)
	if err := db.Table(tableName).Where(w, inputValue).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
