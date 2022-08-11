package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/expressions"
	"gorm.io/gorm"
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

	// Add the WHERE clauses derived from IMPLICIT inputs.
	tx, err = addInputFilters(operation, args, tx)
	if err != nil {
		return nil, err
	}
	// Add the WHERE clauses derived from EXPLICIT inputs (i.e. the operation's where clauses).
	tx, err = addWhereFilters(operation, schema, args, tx)
	if err != nil {
		return nil, err
	}

	// Todo: should we validate the type of the values?, or let postgres object to them later?

	result := []map[string]any{}
	tx = tx.Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	n := len(result)
	if n == 0 {
		return nil, errors.New("no records found for Get() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", n)
	}
	return toLowerCamelMap(result[0]), nil
}

func addInputFilters(op *proto.Operation, args map[string]any, tx *gorm.DB) (*gorm.DB, error) {
	for _, input := range op.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}
		identifier := input.Target[0]
		valueFromArg, ok := args[identifier]
		if !ok {
			return nil, fmt.Errorf("missing argument: %s", identifier)
		}
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(identifier))
		tx = tx.Where(w, valueFromArg)
	}
	return tx, nil
}

func addWhereFilters(
	op *proto.Operation,
	schema *proto.Schema,
	args map[string]any,
	tx *gorm.DB) (*gorm.DB, error) {
	for _, e := range op.WhereExpressions {
		expr, err := expressions.Parse(e.Source)
		if err != nil {
			return nil, err
		}
		identifier, exprValue, err := interpretExpression(expr, op, schema, args)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(identifier))
		tx = tx.Where(w, exprValue)
	}
	return tx, nil
}
