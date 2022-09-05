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

// List implements a Keel List Action.
// In quick overview this means generating a SQL query
// based on the List operation's Inputs and Where clause,
// running that query, and returning the results.
func List(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema,
	inputs interface{}) (records interface{}, hasNextPage bool, hasPreviousPage bool, err error) {
	listInput, err := buildListInput(operation, inputs)
	if err != nil {
		return nil, false, false, err
	}
	db, err := runtimectx.GetDB(ctx)
	if err != nil {
		return nil, false, false, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	tableName := strcase.ToSnake(model.Name)

	// Initialise a query on the table = to which we'll add Where clauses.
	tx := db.Table(tableName)

	// Add the WHERE clauses derived from the inputs.
	tx, err = addListInputFilters(operation, listInput, tx)
	if err != nil {
		return nil, false, false, err
	}

	// todo
	// Add the WHERE clauses derived from EXPLICIT inputs (i.e. the operation's where clauses).
	// tx, err = addWhereFilters(operation, schema, args, tx)
	// if err != nil {
	// 	return nil, err
	// }

	// Where clause to implement the after/before paging request
	tx = addAfterBefore(tx, listInput.Page)

	// Now ordering
	tx = addOrderByID(tx)

	// Put a LIMIT clause on the sql, if the Page mandate is asking for the first-N after x, or the
	// last-N before, x. The limit it applies one more than the number requested to help detect if
	// there more pages available.
	tx, numRequested := addLimit(tx, listInput.Page)

	// Todo: should we validate the type of the values?, or let postgres object to them later?

	// Execute the SQL query.
	result := []map[string]any{}
	tx = tx.Find(&result)
	if tx.Error != nil {
		return nil, false, false, tx.Error
	}
	res := toLowerCamelMaps(result)

	// Reason over the results to judge if there is a next page and to return only
	// the records requested (not the extra one).
	switch {
	case listInput.Page.After != "" && len(result) > listInput.Page.First:
		hasNextPage = true
		res = res[0 : numRequested-1]
	case listInput.Page.Before != "" && len(result) > listInput.Page.Last:
		hasPreviousPage = true
		res = res[1 : listInput.Page.Last+1]
	}

	// todo, this is not a robust implementation - upgrade it to the lead() / lag() pattern that Tom suggested here:
	// https://teamkeel.slack.com/archives/D03C08FGN5C/p1661959457265179

	return res, hasNextPage, hasPreviousPage, nil
}

// addListInputFilters adds Where clauses to the given gorm.DB corresponding to the
// given ListInput.
func addListInputFilters(op *proto.Operation, listInput *ListInput, tx *gorm.DB) (*gorm.DB, error) {
	// We'll look at each of the fields specified as inputs by the operation in the schema,
	// and then try to find these referenced by the where filters in the given ListInput.
	for _, schemaInput := range op.Inputs {
		if schemaInput.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			return nil, errors.New("not yet supported: explicit inputs for list actions")
		}
		expectedFieldName := schemaInput.Target[0]
		var matchingWhere *Where
		for _, where := range listInput.Wheres {
			if where.Name == expectedFieldName {
				matchingWhere = where
				break
			}
		}
		if matchingWhere == nil {
			return nil, fmt.Errorf("operation expects an input named: <%s>, but none is present on the request", expectedFieldName)
		}
		var err error
		tx, err = addWhere(tx, expectedFieldName, matchingWhere)
		if err != nil {
			return nil, err
		}
	}
	return tx, nil
}

// addWhere updates the given gorm.DB tx with a where clause that represents the given
// query.
func addWhere(tx *gorm.DB, columnName string, where *Where) (*gorm.DB, error) {
	switch where.Operator {
	case OperatorEquals:
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))
		return tx.Where(w, where.Operand), nil
	case OperatorStartsWith:
		operandStr, ok := where.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("cannot case this: %v to a string", where.Operand)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandStr+"%%"), nil
	default:
		return nil, fmt.Errorf("operator: %v is not yet supported", where.Operator)
	}
}

func addAfterBefore(tx *gorm.DB, page Page) *gorm.DB {
	switch {
	case page.After != "":
		return tx.Where("ID > ?", page.After)
	case page.Before != "":
		return tx.Where("ID < ?", page.Before)
	}
	return tx
}

func addOrderByID(tx *gorm.DB) *gorm.DB {
	return tx.Order("id")
}

// addLimit puts a LIMIT clause on the query to return the number of records
// specified by the Page mandate (plus 1). It adds one to make it possible to detect,
// hasNextPage / hasPreviousPage.
func addLimit(tx *gorm.DB, page Page) (txOut *gorm.DB, numRequested int) {
	var n int
	switch {
	case page.First != 0:
		n = page.First + 1
		return tx.Limit(n), n
	case page.Last != 0:
		n = page.Last + 1
		return tx.Limit(n), n
	}
	return tx, n
}

// buildListInput consumes the dictionary that carries the LIST operation input values on the
// incoming request, and composes a corresponding ListInput object that is good
// to pass to the generic List() function.
func buildListInput(operation *proto.Operation, requestInputArgs any) (*ListInput, error) {

	argsMap, ok := requestInputArgs.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot cast this: %+v to map[string]any", requestInputArgs)
	}
	page, err := parsePage(argsMap)
	if err != nil {
		return nil, err
	}
	whereInputs, ok := argsMap["where"]
	if !ok {
		return nil, fmt.Errorf("arguments map does not contain a where key: %v", argsMap)
	}
	whereInputsAsMap, ok := whereInputs.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot cast this: %v to a map[string]any", whereInputs)
	}

	wheres := []*Where{}
	for argName, argValue := range whereInputsAsMap {
		argValueAsMap, ok := argValue.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a map[string]any", argValue)
		}
		for operatorStr, operand := range argValueAsMap {
			op, err := operator(operatorStr)
			if err != nil {
				return nil, err
			}
			where := &Where{
				Name:     argName,
				Operator: op,
				Operand:  operand,
			}
			wheres = append(wheres, where)
		}
	}
	inp := &ListInput{
		Page:   page,
		Wheres: wheres,
	}
	return inp, nil
}

// parsePage extracts page mandate information from the given map and uses it to
// compose a Page.
func parsePage(args map[string]any) (Page, error) {
	page := Page{}

	if first, ok := args["first"]; ok {
		asInt, ok := first.(int)
		if !ok {
			return page, fmt.Errorf("cannot cast this: %v to an int", first)
		}
		page.First = asInt
	}

	if last, ok := args["last"]; ok {
		asInt, ok := last.(int)
		if !ok {
			return page, fmt.Errorf("cannot cast this: %v to an int", last)
		}
		page.Last = asInt
	}

	if after, ok := args["after"]; ok {
		asString, ok := after.(string)
		if !ok {
			return page, fmt.Errorf("cannot cast this: %v to a string", after)
		}
		page.After = asString
	}

	if before, ok := args["before"]; ok {
		asString, ok := before.(string)
		if !ok {
			return page, fmt.Errorf("cannot cast this: %v to a string", before)
		}
		page.Before = asString
	}

	return page, nil
}

// operator converts the given string representation of an operator like
// "eq" into the corresponding Operator value.
func operator(operatorStr string) (op Operator, err error) {
	switch operatorStr {
	case "eq":
		return OperatorEquals, nil
	case "startsWith":
		return OperatorStartsWith, nil
	default:
		return op, fmt.Errorf("unrecognized operator: %s", operatorStr)
	}
}
