package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"gorm.io/gorm"
)

// Get implements a Keel Get Action.
// In quick overview this means generating a SQL query
// based on the Get operation's Inputs and Where clause,
// running that query, and returning the results.
func Get(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema,
	args map[string]any) (interface{}, error) {

	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	tableName := strcase.ToSnake(model.Name)

	// Initialise a query on the table = to which we'll add Where clauses.
	tx := db.Table(tableName)

	// Add the WHERE clauses derived from IMPLICIT inputs.
	tx, err = addGetImplicitInputFilters(operation, args, tx)
	if err != nil {
		return nil, err
	}
	// Add the WHERE clauses derived from EXPLICIT inputs (i.e. the operation's where clauses).
	tx, err = addGetExplicitInputFilters(operation, schema, args, tx)
	if err != nil {
		return nil, err
	}

	// Todo: should we validate the type of the values?, or let postgres object to them later?

	// Execute the SQL query.
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

	resultMap := toLowerCamelMap(result[0])

	// todo: permissions to evaluate at the database-level where applicable
	// https://linear.app/keel/issue/RUN-129/expressions-to-evaluate-at-database-level-where-applicable
	authorized, err := EvaluatePermissions(ctx, operation, schema, resultMap)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.New("not authorized to access this operation")
	}

	return resultMap, nil
}

// addGetImplicitInputFilters adds Where clauses for all the operation inputs, which have type
// IMPLICIT. E.g. "get getPerson(id)"
func addGetImplicitInputFilters(op *proto.Operation, args map[string]any, tx *gorm.DB) (*gorm.DB, error) {
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

// addGetExplicitInputFilters adds Where clauses for all the operation's Where clauses.
// E.g.
//
//	get getPerson(name: Text) {
//		@where(person.name == name)
//	}
func addGetExplicitInputFilters(
	op *proto.Operation,
	schema *proto.Schema,
	args map[string]any,
	tx *gorm.DB) (*gorm.DB, error) {
	for _, e := range op.WhereExpressions {
		expr, err := parser.ParseExpression(e.Source)
		if err != nil {
			return nil, err
		}
		// This call gives us the column and the value to use like this:
		// tx.Where(fmt.Sprintf("%s = ?", column), value)
		identifier, exprValue, err := interpretExpressionGivenArgs(expr, op, schema, args)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(identifier))
		tx = tx.Where(w, exprValue)
	}
	return tx, nil
}
