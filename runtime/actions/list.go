package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/expressions"
	"gorm.io/gorm"
)

func main() {
	builder := NewQueryBuilder(nil, nil, nil, nil)
	builder.
}


// List implements a Keel List Action.
// In quick overview this means generating a SQL query
// based on the List operation's Inputs and Where clause,
// running that query, and returning the results.
func List(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema,
	inputs map[string]any) (records interface{}, hasNextPage bool, err error) {

	db, err := runtimectx.GetDatabase(ctx)
	if err != nil {
		return nil, false, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	qry, err := buildQuery(db, model, operation, schema, inputs)
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

func buildQuery(
	db *gorm.DB,
	model *proto.Model,
	op *proto.Operation,
	schema *proto.Schema,
	args map[string]any,
) (*gorm.DB, error) {

	listInput, err := buildListInput(op, args)
	if err != nil {
		return nil, err
	}

	tableName := strcase.ToSnake(model.Name)

	// Initialise a query on the table = to which we'll add Where clauses.
	qry := db.Table(tableName)

	// Specify the ORDER BY - but also a "LEAD" extra column to harvest extra data
	// that helps to determin "hasNextPage".
	qry = addOrderingAndLead(qry)

	// Add the WHERE clauses derived from the implicit inputs.
	qry, err = addListImplicitFilters(op, listInput, qry)
	if err != nil {
		return nil, err
	}

	// Add the WHERE clauses derived from EXPLICIT inputs (i.e. the operation's where clauses).
	qry, err = addListExplicitInputFilters(op, schema, listInput, qry)
	if err != nil {
		return nil, err
	}

	// Where clause to implement the after/before paging request
	qry = addAfterBefore(qry, listInput.Page)

	// Put a LIMIT clause on the sql, if the Page mandate is asking for the first-N after x, or the
	// last-N before x.
	qry = addLimit(qry, listInput.Page)

	return qry, nil
}

// addListImplicitFilters adds Where clauses to the given gorm.DB corresponding to the
// implicit inputs present in the given ListInput.
func addListImplicitFilters(op *proto.Operation, listInput *ListInput, tx *gorm.DB) (*gorm.DB, error) {
	// We'll look at each of the fields specified as implicit inputs by the operation in the schema,
	// and then try to find these referenced by the where filters in the given ListInput.
	for _, schemaInput := range op.Inputs {
		if schemaInput.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		expectedFieldName := schemaInput.Target[0]
		var matchingWhere *ImplicitFilter
		for _, where := range listInput.ImplicitFilters {
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
		tx, err = addWhereForImplicitFilter(tx, expectedFieldName, matchingWhere, schemaInput.Type)
		if err != nil {
			return nil, err
		}
	}
	return tx, nil
}

// addListExplicitInputFilters adds Where clauses for all the operation's Where clauses.
// E.g.
//
//	list getPerson(name: Text) {
//		@where(person.name == name)
//	}
func addListExplicitInputFilters(
	op *proto.Operation,
	schema *proto.Schema,
	listInput *ListInput,
	tx *gorm.DB) (*gorm.DB, error) {
	for _, e := range op.WhereExpressions {
		expr, err := expressions.Parse(e.Source)
		if err != nil {
			return nil, err
		}
		// This call gives us the column and the value to use like this:
		// tx.Where(fmt.Sprintf("%s = ?", column), value)
		fieldName, err := interpretExpressionField(expr, op, schema)
		if err != nil {
			return nil, err
		}
		// Find the ExplicitInputFilter that belongs to the field being targeted by the
		// expression.
		filter, ok := lo.Find(listInput.ExplicitFilters, func(f *ExplicitFilter) bool {
			return f.Name == fieldName
		})
		if !ok {
			return nil, fmt.Errorf("input does not provide a filter for key: %s", fieldName)
		}
		scalarValue := filter.ScalarValue

		w := fmt.Sprintf("%s = ?", strcase.ToSnake(fieldName))
		tx = tx.Where(w, scalarValue)
	}
	return tx, nil
}

// addWhereForImplicitFilter updates the given gorm.DB tx with a where clause that represents the given
// query.
func addWhereForImplicitFilter(tx *gorm.DB, columnName string, filter *ImplicitFilter, inputType *proto.TypeInfo) (*gorm.DB, error) {
	switch filter.Operator {
	case OperatorEquals:
		operand := filter.Operand

		if inputType.Type == proto.Type_TYPE_DATE || inputType.Type == proto.Type_TYPE_DATETIME || inputType.Type == proto.Type_TYPE_TIMESTAMP {
			timeOperand, err := parseTimeOperand(filter.Operand, inputType.Type)
			if err != nil {
				return nil, err
			}
			operand = timeOperand
		}

		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))
		return tx.Where(w, operand), nil

	case OperatorStartsWith:
		operandStr, ok := filter.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a string", filter.Operand)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandStr+"%%"), nil

	case OperatorEndsWith:
		operandStr, ok := filter.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a string", filter.Operand)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		return tx.Where(w, "%%"+operandStr), nil

	case OperatorContains:
		operandStr, ok := filter.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a string", filter.Operand)
		}
		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		return tx.Where(w, "%%"+operandStr+"%%"), nil

	case OperatorOneOf:
		operandStrings, ok := filter.Operand.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a []interface{}", filter.Operand)
		}
		w := fmt.Sprintf("%s in ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandStrings), nil

	case OperatorLessThan:
		operandInt, ok := filter.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", filter.Operand)
		}
		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorLessThanEquals:
		operandInt, ok := filter.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", filter.Operand)
		}
		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorGreaterThan:
		operandInt, ok := filter.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", filter.Operand)
		}
		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorGreaterThanEquals:
		operandInt, ok := filter.Operand.(int)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to an int", filter.Operand)
		}
		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandInt), nil

	case OperatorBefore:
		operandTime, err := parseTimeOperand(filter.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

	case OperatorAfter:
		operandTime, err := parseTimeOperand(filter.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

	case OperatorOnOrBefore:
		operandTime, err := parseTimeOperand(filter.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

	case OperatorOnOrAfter:
		operandTime, err := parseTimeOperand(filter.Operand, inputType.Type)
		if err != nil {
			return nil, err
		}
		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		return tx.Where(w, operandTime), nil

	default:
		return nil, fmt.Errorf("operator: %v is not yet supported", filter.Operator)
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
func buildListInput(operation *proto.Operation, argsMap map[string]any) (*ListInput, error) {

	allOptionalInputs := true
	for _, in := range operation.Inputs {
		if !in.Optional {
			allOptionalInputs = false
		}
	}

	implicitFilters := []*ImplicitFilter{}
	explicitFilters := []*ExplicitFilter{}

	fmt.Printf("args: %v", argsMap)

	if allOptionalInputs && argsMap == nil {
		// No inputs and nothing required. Set default paging
		return &ListInput{
			Page:            Page{First: 50},
			ImplicitFilters: implicitFilters,
			ExplicitFilters: explicitFilters,
		}, nil
	}

	page, err := parsePage(argsMap)
	if err != nil {
		return nil, err
	}
	whereInputs, ok := argsMap["where"]
	if !ok {
		// We have some required inputs but there is no where key
		if !allOptionalInputs {
			return nil, fmt.Errorf("arguments map does not contain a where key: %v", argsMap)
		}
	} else {
		whereInputsAsMap, ok := whereInputs.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot cast this: %v to a map[string]any", whereInputs)
		}

		for argName, argValue := range whereInputsAsMap {

			switch {
			case isMap(argValue):
				argValueAsMap := argValue.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("cannot cast this: %v to a map[string]any", argValue)
				}

				for operatorStr, operand := range argValueAsMap {
					op, err := operator(operatorStr)
					if err != nil {
						return nil, err
					}
					implicitFilter := &ImplicitFilter{
						Name:     argName,
						Operator: op,
						Operand:  operand,
					}
					implicitFilters = append(implicitFilters, implicitFilter)
				}
			default:
				explicitFilter := &ExplicitFilter{
					Name:        argName,
					ScalarValue: argValue,
				}
				explicitFilters = append(explicitFilters, explicitFilter)
			}
		}
	}

	inp := &ListInput{
		Page:            page,
		ImplicitFilters: implicitFilters,
		ExplicitFilters: explicitFilters,
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
