package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	inputs interface{}) (records interface{}, hasNextPage bool, err error) {
	listInput, err := buildListInput(operation, inputs)
	if err != nil {
		return nil, false, err
	}
	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, false, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	qry, err := buildQuery(db, model, operation, listInput)
	if err != nil {
		return nil, false, err
	}

	// Execute the SQL query.
	result := []map[string]any{}
	qry = qry.Find(&result)
	if qry.Error != nil {
		return nil, false, qry.Error
	}

	// Sort out the hasNextPage value, and clean up the response.
	if len(result) > 0 {
		last := result[len(result)-1]
		hasNextPage = last["hasnext"].(bool)
	}
	res := toLowerCamelMaps(result)
	for _, row := range res {
		delete(row, "hasnext")
	}
	return res, hasNextPage, nil
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
		if matchingWhere == nil && schemaInput.Optional {
			// If the input is optional we don't need a where input
			continue
		}
		if matchingWhere == nil {
			return nil, fmt.Errorf("operation expects an input named: <%s>, but none is present on the request", expectedFieldName)
		}

		var err error
		tx, err = addWhere(tx, expectedFieldName, matchingWhere, schemaInput.Type)
		if err != nil {
			return nil, err
		}
	}
	return tx, nil
}

// addWhere updates the given gorm.DB tx with a where clause that represents the given
// query.
func addWhere(tx *gorm.DB, columnName string, where *Where, inputType *proto.TypeInfo) (*gorm.DB, error) {
	switch where.Operator {
	case OperatorEquals:
		operand := where.Operand

		if inputType.Type == proto.Type_TYPE_DATE || inputType.Type == proto.Type_TYPE_DATETIME || inputType.Type == proto.Type_TYPE_TIMESTAMP {
			timeOperand, err := parseTimeOperand(where.Operand, inputType.Type)
			if err != nil {
				return nil, err
			}
			operand = timeOperand
		}

		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))
		return tx.Where(w, operand), nil

	case OperatorStartsWith:
		operandStr, ok := where.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a string", where.Operand)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandStr+"%%"), nil

	case OperatorEndsWith:
		operandStr, ok := where.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a string", where.Operand)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		return tx.Where(w, "%%"+operandStr), nil

	case OperatorContains:
		operandStr, ok := where.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a string", where.Operand)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		return tx.Where(w, "%%"+operandStr+"%%"), nil

	case OperatorOneOf:
		operandStrings, ok := where.Operand.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a []interface{}", where.Operand)
		}
		w := fmt.Sprintf("%s in ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandStrings), nil

	case OperatorLessThan:
		operandInt, ok := where.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", where.Operand)
		}
		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorLessThanEquals:
		operandInt, ok := where.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", where.Operand)
		}
		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorGreaterThan:
		operandInt, ok := where.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", where.Operand)
		}
		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorGreaterThanEquals:
		operandInt, ok := where.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", where.Operand)
		}
		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorBefore:
		operandTime, err := parseTimeOperand(where.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

	case OperatorAfter:
		operandTime, err := parseTimeOperand(where.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

	case OperatorOnOrBefore:
		operandTime, err := parseTimeOperand(where.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

	case OperatorOnOrAfter:
		operandTime, err := parseTimeOperand(where.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

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

// addLimit puts a LIMIT clause on the query to return the number of records
// specified by the Page mandate.
func addLimit(tx *gorm.DB, page Page) *gorm.DB {
	switch {
	case page.First != 0:
		return tx.Limit(page.First)
	case page.Last != 0:
		return tx.Limit(page.Last)
	default:
		return tx
	}
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

// praseTime extract and parses time for date/time based operators
// Supports timestamps passed in map[seconds:int] and dates passesd as map[day:int month:int year:int]
func parseTimeOperand(operand any, inputType proto.Type) (t *time.Time, err error) {
	operandMap, ok := operand.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot cast this: %v to a map[string]interface{}", operand)
	}

	switch inputType {
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		seconds := operandMap["seconds"]
		secondsInt, ok := seconds.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to int", seconds)
		}
		unix := time.Unix(int64(secondsInt), 0).UTC()
		t = &unix

	case proto.Type_TYPE_DATE:
		day := operandMap["day"]
		month := operandMap["month"]
		year := operandMap["year"]

		dayInt, ok := day.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast days: %v to int", day)
		}
		monthInt, ok := month.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast month: %v to int", month)
		}
		yearInt, ok := year.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast year: %v to int", year)
		}

		time, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-%02d", yearInt, monthInt, dayInt))
		if err != nil {
			return nil, fmt.Errorf("cannot parse date %s", err)
		}
		t = &time

	default:
		return nil, fmt.Errorf("unknown time field type")
	}

	return t, nil
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

// addOrderingAndLead puts in a SELECT statement that puts in the ORDER BY clause to support
// paging. It also uses the SQL "lead(1)" idiom to deduces if each row has a following row, wich
// we can then use to determine if a "next" page is available.
func addOrderingAndLead(tx *gorm.DB) *gorm.DB {

	const by = "id"
	selectArgs := `
	 *,
		CASE WHEN lead("id") OVER ( order by ? ) is not null THEN true ELSE false
		END as hasNext
	`

	tx = tx.Select(selectArgs, by)
	tx = tx.Order(by)
	return tx
}

func buildQuery(
	db *gorm.DB,
	model *proto.Model,
	op *proto.Operation,
	listInput *ListInput) (*gorm.DB, error) {

	tableName := strcase.ToSnake(model.Name)

	// Initialise a query on the table = to which we'll add Where clauses.
	qry := db.Table(tableName)

	// Specify the ORDER BY - but also a "LEAD" extra column to harvest extra data
	// that helps to determin "hasNextPage".
	qry = addOrderingAndLead(qry)

	// Add the WHERE clauses derived from the inputs.
	qry, err := addListInputFilters(op, listInput, qry)
	if err != nil {
		return nil, err
	}

	// todo
	// Add the WHERE clauses derived from EXPLICIT inputs (i.e. the operation's where clauses).
	// tx, err = addWhereFilters(operation, schema, args, tx)
	// if err != nil {
	// 	return nil, err
	// }

	// Where clause to implement the after/before paging request
	qry = addAfterBefore(qry, listInput.Page)

	// Put a LIMIT clause on the sql, if the Page mandate is asking for the first-N after x, or the
	// last-N before x.
	qry = addLimit(qry, listInput.Page)

	return qry, nil
}
