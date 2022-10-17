package actions

import (
	"fmt"
	"time"

	"github.com/teamkeel/keel/proto"
	"gorm.io/gorm"
)

// List implements a Keel List Action.
// In quick overview this means generating a SQL query
// based on the List operation's Inputs and Where clause,
// running that query, and returning the results.
// func List(
// 	ctx context.Context,
// 	operation *proto.Operation,
// 	schema *proto.Schema,
// 	inputs map[string]any) (records interface{}, hasNextPage bool, err error) {

// 	db, err := runtimectx.GetDatabase(ctx)
// 	if err != nil {
// 		return nil, false, err
// 	}

// 	model := proto.FindModel(schema.Models, operation.ModelName)

// 	qry, err := buildQuery(db, model, operation, schema, inputs)
// 	if err != nil {
// 		return nil, false, err
// 	}

// 	// Execute the SQL query.
// 	result := []map[string]any{}
// 	qry = qry.Find(&result)
// 	if qry.Error != nil {
// 		return nil, false, qry.Error
// 	}

// 	// Sort out the hasNextPage value, and clean up the response.
// 	if len(result) > 0 {
// 		last := result[len(result)-1]
// 		hasNextPage = last["hasnext"].(bool)
// 	}
// 	res := toLowerCamelMaps(result)
// 	for _, row := range res {
// 		delete(row, "hasnext")
// 	}
// 	return res, hasNextPage, nil
// }

// func buildQuery(
// 	db *gorm.DB,
// 	model *proto.Model,
// 	op *proto.Operation,
// 	schema *proto.Schema,
// 	args map[string]any,
// ) (*gorm.DB, error) {

// 	listInput, err := buildListInput(op, args)
// 	if err != nil {
// 		return nil, err
// 	}

// 	tableName := strcase.ToSnake(model.Name)

// 	// Initialise a query on the table = to which we'll add Where clauses.
// 	qry := db.Table(tableName)

// 	// Specify the ORDER BY - but also a "LEAD" extra column to harvest extra data
// 	// that helps to determin "hasNextPage".
// 	qry = addOrderingAndLead(qry)

// 	// Add the WHERE clauses derived from the implicit inputs.
// 	qry, err = addListImplicitFilters(op, listInput, qry)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Add the WHERE clauses derived from EXPLICIT inputs (i.e. the operation's where clauses).
// 	qry, err = addListExplicitInputFilters(op, schema, listInput, qry)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Where clause to implement the after/before paging request
// 	qry = addAfterBefore(qry, listInput.Page)

// 	// Put a LIMIT clause on the sql, if the Page mandate is asking for the first-N after x, or the
// 	// last-N before x.
// 	qry = addLimit(qry, listInput.Page)

// 	return qry, nil
// }

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

func isMap(x any) bool {
	if _, ok := x.(map[string]any); ok {
		return true
	}
	return false
}
