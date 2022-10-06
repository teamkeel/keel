package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

// Delete implements a Keel Delete Action.
// In quick overview this means generating a SQL query
// based on the Delete operation's Inputs and Where clause,
// running that query, and returning the results.
func Delete(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema,
	args map[string]any) (bool, error) {

	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return false, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	tableName := strcase.ToSnake(model.Name)

	// Initialise a query on the table = to which we'll add Where clauses.
	tx := db.Table(tableName)

	// Add the WHERE clauses derived from IMPLICIT inputs.
	tx, err = addDeleteImplicitInputFilters(operation, args, tx)
	if err != nil {
		return false, err
	}

	// todo: reading the model can be removed once permissions can evaluate at the database-level
	// https://linear.app/keel/issue/RUN-129/expressions-to-evaluate-at-database-level
	result := []map[string]any{}
	tx = tx.Find(&result)
	if tx.Error != nil {
		return false, tx.Error
	}
	n := len(result)
	if n == 0 {
		return false, errors.New("no records found for Delete() operation")
	}
	if n > 1 {
		return false, fmt.Errorf("Delete() operation should find only one record, it found: %d", n)
	}

	resultMap := toLowerCamelMap(result[0])

	authorized, err := EvaluatePermissions(ctx, operation, schema, resultMap)
	if err != nil {
		return false, err
	}
	if !authorized {
		return false, errors.New("not authorized to access this operation")
	}

	// Execute the SQL query.
	record := []map[string]any{}
	tx = tx.Delete(&record)

	if tx.Error != nil || tx.RowsAffected == 0 {
		return false, tx.Error
	}

	return true, nil
}

func addDeleteImplicitInputFilters(op *proto.Operation, args map[string]any, tx *gorm.DB) (*gorm.DB, error) {
	for _, input := range op.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		identifier := input.Target[0]
		valueFromArg, ok := args[identifier]
		if !ok {
			return nil, fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", identifier, args)
		}
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(identifier))
		tx = tx.Where(w, valueFromArg)
	}
	return tx, nil
}
