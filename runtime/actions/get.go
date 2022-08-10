package actions

import (
	"context"
	"errors"
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

	db, err := runtimectx.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	tableName := strcase.ToSnake(model.Name)
	// Initialise a query on the table.
	tx := db.Table(tableName)

	// A Get operation (conceptually) needs exactly one filter on a unique field.
	// But this can be in the form of an schema Input, or a schema Where clause.

	switch {
	case len(operation.Inputs) == 1:
		input := operation.Inputs[0]
		identifier := input.Target[0]
		valueFromArg, ok := args[identifier]
		if !ok {
			return nil, fmt.Errorf("missing argument: %s", identifier)
		}
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(identifier))
		tx = tx.Where(w, valueFromArg)
	case len(operation.WhereExpressions) == 1:
		return nil, errors.New("where expressions not implemented for Get operations")
	default:
		return nil, errors.New("get operation must have either one input or one where clause")
	}

	// Todo: should we validate the type of the values?, or let postgres object to them later?

	result := map[string]any{}
	tx = tx.Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// TODO.
	// The gorm docs say that Find() should raise ErrRecordNotFound, but when used as
	// above it does not - for reasons I don't understand.
	// However it seems the RowsAffected field can tell us.
	//
	// See: https://gorm.io/docs/query.html#Retrieving-a-single-object
	if tx.RowsAffected == 0 {
		return nil, errors.New("no records found for Get() operation")
	}
	return toLowerCamelMap(result), nil
}
