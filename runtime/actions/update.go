package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"golang.org/x/exp/maps"
	"gorm.io/gorm"
)

// Update implements a Keel Update Action.
// In quick overview this means generating a SQL query
// based on the Update operation's Inputs and Where clause,
// running that query, and returning the results.
func Update(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema,
	args map[string]any) (map[string]any, error) {

	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	tableName := strcase.ToSnake(model.Name)

	// Initialise a query on the table = to which we'll add Where clauses.
	tx := db.Table(tableName)

	// Add the WHERE clauses derived from IMPLICIT inputs.
	tx, err = addUpdateImplicitInputFilters(operation, args, tx)
	if err != nil {
		return nil, err
	}

	// todo: permissions to evaluate at the database-level (expressions-to-SQL)
	// https://linear.app/keel/issue/RUN-129/expressions-to-evaluate-at-database-level-where-applicable
	result := []map[string]any{}
	tx = tx.Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	n := len(result)
	if n == 0 {
		return nil, errors.New("no records found for Update() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Update() operation should find only one record, it found: %d", n)
	}
	authorized, err := EvaluatePermissions(ctx, operation, schema, toLowerCamelMap(result[0]))
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.New("not authorized to access this operation")
	}

	values := map[string]any{}
	if args["values"] != nil {
		argValues, ok := args["values"].(map[string]any)

		if !ok {
			return nil, fmt.Errorf("values not provided")
		}

		values = argValues
	}

	modelMap := map[string]any{}
	for k, v := range values {
		modelMap[strcase.ToSnake(k)] = v
	}

	setArgs, err := SetExpressionInputsToModelMap(operation, values, schema, ctx)

	if err != nil {
		return nil, err
	}

	maps.Copy(modelMap, setArgs)

	maps.DeleteFunc(modelMap, func(k string, v any) bool {
		match := lo.SomeBy(model.Fields, func(f *proto.Field) bool {
			return strcase.ToSnake(f.Name) == k
		})

		return !match
	})

	tx.Updates(modelMap)

	if tx.Error != nil || tx.RowsAffected == 0 {
		return nil, tx.Error
	}

	// todo: figure out how to make tx.Clauses(clause.Returning{}).Updates(values) work with dynamically created structs
	// usually in a non dynamic model, you would use .Model(User{}) but we do not know what the Model is, and havent built
	// a struct for it; Gorm assumes you know what your model looks like upfront
	// As a shortcut, we just do a select to hydrate the latest state of the record
	resultMap := map[string]any{}

	tx.Take(&resultMap)

	return toLowerCamelMap(resultMap), nil
}

func addUpdateImplicitInputFilters(op *proto.Operation, args map[string]any, tx *gorm.DB) (*gorm.DB, error) {
	wheres, ok := args["where"].(map[string]any)

	if !ok {
		return nil, fmt.Errorf("where constraint not provided")
	}

	for _, input := range op.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		if input.Mode != proto.InputMode_INPUT_MODE_READ {
			continue
		}

		identifier := input.Target[0]
		valueFromArg, ok := wheres[identifier]
		if !ok {
			return nil, fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", identifier, args)
		}
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(identifier))
		tx = tx.Where(w, valueFromArg)
	}
	return tx, nil
}
